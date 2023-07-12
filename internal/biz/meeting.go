/*
 * @Date: 2023-05-24 22:39:01
 * @LastEditors: Please set LastEditors
 * @LastEditTime: 2023-07-12 21:27:04
 * @FilePath: /gpt-meeting-service/internal/biz/meeting.go
 */
package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gpt-meeting-service/internal/domain"
	"gpt-meeting-service/internal/lib/datastruct"
	"gpt-meeting-service/internal/lib/utils"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/sashabaranov/go-openai"
)

const LIMIT_COUNT = 3

type MeetingRepo interface {
	Create(context.Context, *domain.Meeting) (string, error)
	FindOne(context.Context, string) (*domain.Meeting, error)
	List(ctx context.Context, pageNum, pageSize int64, createdBy string) (*[]domain.Meeting, int64, error)
	CountByUser(ctx context.Context, createdBy string) (int64, error)
	UpdateMeeting(context.Context, *domain.Meeting) error
	UpdateStatus(context.Context, string, string, domain.Status) error
	UpdateMeetingFlow(context.Context, string, string, domain.FlowItem) error
	UpdateMeetingData(context.Context, string, string, domain.MeetingDataItem) error
	UpdateMeetingChatTreeNode(ctx context.Context, meetingId string, meetingDataItemId string, chatNodeIdPath []string, treeNode datastruct.TreeNode[domain.Conversation]) error
	UpdateMeetingChatTreeNodeData(ctx context.Context, meetingId string, meetingDataItemId string, chatNodeIdPath []string, conversation domain.Conversation) error
}

type MeetingUsecase struct {
	meetingRepo         MeetingRepo
	meetingTemplateRepo MeetingTemplateRepo
	gpt                 *Gpt
	log                 *log.Helper
}

func NewMeetingUsecase(meetingRepo MeetingRepo, meetingTemplateRepo MeetingTemplateRepo, gpt *Gpt, logger log.Logger) *MeetingUsecase {
	return &MeetingUsecase{
		meetingRepo:         meetingRepo,
		meetingTemplateRepo: meetingTemplateRepo,
		gpt:                 gpt,
		log:                 log.NewHelper(logger),
	}
}

/**
 * @description: 新建会议
 * @param {context.Context} ctx
 * @param {*domain.NewMeetingRequest} newMeeting
 * @return {*}
 */
func (m *MeetingUsecase) NewMeeting(ctx http.Context, newMeeting *domain.NewMeetingRequest) (*domain.NewMeetingResponse, error) {
	// 检查无apiKey用户的创建次数
	request := ctx.Request()
	headers := request.Header
	re := regexp.MustCompile(`:\d+$`)
	userIp := re.ReplaceAllString(request.RemoteAddr, "")
	fakeUid := fmt.Sprintf("IP=%s,UserAgent=%s", userIp, request.UserAgent()) // 未登录的伪用户ID
	if len(headers["Apikey"]) == 0 || headers["Apikey"][0] == "" {
		count, err := m.meetingRepo.CountByUser(ctx, fakeUid)
		if err != nil {
			return nil, err
		}
		if count >= LIMIT_COUNT {
			return nil, errors.New("体验次数已用完～")
		}
	}
	// 找到对应的模版
	meetingTemplate, err := m.meetingTemplateRepo.FindOne(ctx, newMeeting.TemplateId)
	if err != nil {
		return nil, err
	}
	templateGraph := meetingTemplate.TemplateGraph
	execSort, err := templateGraph.TopologicalSort() // 执行顺序
	if err != nil {
		return nil, err
	}
	nodeAncestorsMap := templateGraph.GetAncestors() // 祖先节点列表

	// 执行流程
	meetingFlow := []domain.FlowItem{}
	for _, item := range execSort {
		upstream := []string{}
		if _, ok := nodeAncestorsMap[item.GetId()]; ok {
			upstream = nodeAncestorsMap[item.GetId()]
		}
		meetingFlow = append(meetingFlow, domain.FlowItem{
			NodeInfo: item,
			Status:   domain.Wait,
			Upstream: upstream,
		})
	}
	meeting := &domain.Meeting{
		Name:        newMeeting.Name,
		TemplateId:  newMeeting.TemplateId,
		Status:      domain.Idle,
		MeetingFlow: meetingFlow,
		CreatedBy:   fakeUid,
		CreatedTime: time.Now().Unix(),
	}
	meetingId, err := m.meetingRepo.Create(ctx, meeting)
	if err != nil {
		return nil, err
	}
	// 返回响应
	resp := &domain.NewMeetingResponse{
		MeetingId:   meetingId,
		MeetingFlow: meetingFlow,
	}
	return resp, nil
}

/**
 * @description: 历史会议记录
 * @param {context.Context} ctx
 * @param {*domain.HistoryMeetingRequest} condition
 * @return {*}
 */
func (m *MeetingUsecase) HistoryMeeting(ctx http.Context, condition *domain.HistoryMeetingRequest) (*domain.HistoryMeetingResponse, error) {
	request := ctx.Request()
	re := regexp.MustCompile(`:\d+$`)
	userIp := re.ReplaceAllString(request.RemoteAddr, "")
	fakeUid := fmt.Sprintf("IP=%s,UserAgent=%s", userIp, request.UserAgent()) // 未登录的伪用户ID
	meetingList, total, err := m.meetingRepo.List(ctx, condition.PageNum, condition.PageSize, fakeUid)
	if err != nil {
		return nil, err
	}
	resp := domain.HistoryMeetingResponse{
		Total: total,
		Data:  meetingList,
	}
	return &resp, nil
}

/**
 * @description: 查询任务进展
 * @param {context.Context} ctx
 * @param {string} meetingId
 * @return {*}
 */
func (m *MeetingUsecase) Progress(ctx context.Context, meetingId string) (*domain.ProcessResponse, error) {
	var nowStatus domain.Status = domain.Done
	var nowFlowItem domain.FlowItem
	var nextFlowItem domain.FlowItem
	meeting, err := m.meetingRepo.FindOne(ctx, meetingId)
	if err != nil {
		return nil, err
	}
	meetingFlow := meeting.MeetingFlow
	var activedNum = len(meetingFlow)
	utils.Reverse(meetingFlow) // 反转顺序获取进行中和下一步需要执行的节点
	for _, flowItem := range meetingFlow {
		if flowItem.Status == domain.Processing { // 运行中
			nowStatus = domain.Running
			nowFlowItem = flowItem
			activedNum--
		} else if flowItem.Status == domain.Wait {
			nowStatus = domain.Idle
			nextFlowItem = flowItem
			activedNum--
		}
	}
	// b, _ := json.Marshal(meeting.MeetingData)
	// fmt.Println(string(b))
	utils.Reverse(meetingFlow)
	result := &domain.ProcessResponse{
		MeetingId:    meetingId,
		MeetingName:  meeting.Name,
		NowStatus:    nowStatus,
		NowFlowItem:  &nowFlowItem,
		NextFlowItem: &nextFlowItem,
		ActivedNum:   activedNum,
		Progress:     meetingFlow,
	}
	return result, nil
}

/**
 * @description: 开始一个环节
 * @param {context.Context} ctx
 * @param {string} meetingId
 * @param {string} flowItemId
 * @return {*}
 */
func (m *MeetingUsecase) StartFlowItem(ctx context.Context, meetingId string, flowItemId string) (interface{}, error) {
	m.meetingRepo.UpdateMeeting(ctx, &domain.Meeting{
		Id:     meetingId,
		Status: domain.Running,
	})
	return nil, m.meetingRepo.UpdateStatus(ctx, meetingId, flowItemId, domain.Processing)
}

/**
 * @description: 结束一个环节
 * @param {context.Context} ctx
 * @param {string} meetingId
 * @param {string} flowItemId
 * @return {*}
 */
func (m *MeetingUsecase) EndFlowItem(ctx context.Context, meetingId string, flowItemId string) (interface{}, error) {
	err := m.meetingRepo.UpdateStatus(ctx, meetingId, flowItemId, domain.Finish)
	if err != nil {
		return nil, err
	}
	// 判断是否所有环节均已结束
	meeting, err := m.meetingRepo.FindOne(ctx, meetingId)
	if err != nil {
		return nil, err
	}
	isAllFinished := true
	for _, item := range meeting.MeetingFlow {
		if item.Status != "finish" {
			isAllFinished = false
			break
		}
	}
	status := domain.Idle
	if isAllFinished {
		overview := ""
		conclusion := ""
		status = domain.Done
		// 所有环节完成
		meetingFlow := meeting.MeetingFlow
		meetingData := meeting.MeetingData
		for _, item := range meetingFlow {
			flowItemId := item.NodeInfo.Id
			meetingDataItem := meetingData[flowItemId]
			if item.NodeInfo.NodeType != domain.NT_Output && item.NodeInfo.NodeType != domain.NT_Input {
				// 非output类型节点都归为过程
				overview += fmt.Sprintf("## %s\n%s\n\n", item.NodeInfo.NodeName, meetingDataItem.Output)
			} else if item.NodeInfo.NodeType == domain.NT_Output {
				conclusion += fmt.Sprintf("## %s\n%s\n\n", item.NodeInfo.NodeName, meetingDataItem.Output)
			}
		}
		err := m.meetingRepo.UpdateMeeting(ctx, &domain.Meeting{
			Id:         meetingId,
			Status:     status,
			OverView:   overview,
			Conclusion: conclusion,
		})
		if err != nil {
			return false, err
		}
	}
	return nil, nil
}

/**
 * @description: 获取input预设信息
 * @param {context.Context} ctx
 * @param {*domain.MeetingFlowItemRequest} meetingFlowItem
 * @return {*}
 */
func (m *MeetingUsecase) GetInputData(ctx context.Context, meetingFlowItem *domain.MeetingFlowItemRequest) (*domain.InputData, error) {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	meeting, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return nil, err
	}
	nodeInfo := flowItem.NodeInfo
	curFlowItemName := nodeInfo.NodeName
	input := ""
	output := ""
	prompt := nodeInfo.OptimizationPrompt
	meetingData := meeting.MeetingData
	if _, ok := meetingData[flowItemId]; ok {
		meetingFlowItem := meetingData[flowItemId]
		if meetingFlowItem.Input != "" {
			input = meetingFlowItem.Input
		}
		if meetingFlowItem.Output != "" {
			output = meetingFlowItem.Output
		}
	}
	return &domain.InputData{
		Presets: &domain.InputPresets{
			MeetingId:          meetingId,
			FlowItemId:         flowItemId,
			Characters:         *nodeInfo.Characters,
			CurFlowItemName:    curFlowItemName,
			OptimizationPrompt: prompt,
		},
		Input:  input,
		Output: output,
	}, nil
}

/**
 * @description: 输入
 * @param {http.Context} ctx
 * @param {*domain.MeetingFlowItemRequest} meetingFlowItem
 * @return {*}
 */
func (m *MeetingUsecase) Input(ctx http.Context, meetingFlowItem *domain.MeetingFlowItemRequest) error {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	topicGoal := meetingFlowItem.Text
	characters := meetingFlowItem.Characters
	optimizationPrompt := meetingFlowItem.Prompt
	meeting, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return err
	}
	if characters == "" {
		characters = flowItem.NodeInfo.Characters.Description
	}
	if optimizationPrompt == "" {
		optimizationPrompt = flowItem.NodeInfo.OptimizationPrompt
	}
	flowItem.NodeInfo.OptimizationPrompt = optimizationPrompt
	flowItem.NodeInfo.Characters.Description = characters
	// 更新到meetingFlow
	err = m.meetingRepo.UpdateMeetingFlow(ctx, meetingId, flowItemId, *flowItem)
	if err != nil {
		return err
	}
	meetingDataItem := domain.MeetingDataItem{
		Input:   topicGoal,
		Process: domain.Process{},
		Output:  "",
	}
	err = m.meetingRepo.UpdateMeetingData(ctx, meetingId, flowItemId, meetingDataItem)
	if err != nil {
		return err
	}
	// 基于用户输入的内容进行优化
	messages := getContext(meeting, flowItem, flowItem.NodeInfo.Characters) // 上下文
	// 当前环节
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("当前环节是[%s]", flowItem.NodeInfo.NodeName),
	})
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: optimizationPrompt + topicGoal,
	})
	_, err = m.gpt.ChatCompletion(ctx, messages)
	return err
}

/**
 * @description: 提交输入
 * @param {http.Context} ctx
 * @param {*domain.MeetingFlowItemRequest} meetingFlowItem
 * @return {*}
 */
func (m *MeetingUsecase) SubmitInput(ctx context.Context, meetingFlowItem *domain.MeetingFlowItemRequest) (interface{}, error) {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	topicGoal := meetingFlowItem.Text
	err := m.meetingRepo.UpdateMeeting(ctx, &domain.Meeting{
		Id:        meetingId,
		TopicGoal: topicGoal,
	})
	if err != nil {
		return false, err
	}
	// 更新数据
	meeting, err := m.meetingRepo.FindOne(ctx, meetingId)
	if err != nil {
		return false, err
	}
	meetingData := meeting.MeetingData
	var meetingDataItem domain.MeetingDataItem
	if _, ok := meetingData[flowItemId]; ok {
		meetingDataItem = meetingData[flowItemId]
		meetingDataItem.Output = topicGoal
	} else {
		meetingDataItem = domain.MeetingDataItem{
			Input:   topicGoal,
			Process: domain.Process{},
			Output:  topicGoal,
		}
	}
	err = m.meetingRepo.UpdateMeetingData(ctx, meetingId, flowItemId, meetingDataItem)
	return true, err
}

/**
 * @description: 获取thinking预设信息
 * @param {http.Context} ctx
 * @param {*domain.MeetingFlowItemRequest} meetingFlowItem
 * @return {*}
 */
func (m *MeetingUsecase) GetThinkingData(ctx context.Context, meetingFlowItem *domain.MeetingFlowItemRequest) (*domain.ThinkingData, error) {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	meeting, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return nil, err
	}
	// 背景信息
	messages := getContext(meeting, flowItem, nil)
	background := ""
	for _, message := range messages {
		background += message.Content + "\n"
	}

	nodeInfo := flowItem.NodeInfo
	curFlowItemName := nodeInfo.NodeName
	// 如果存在在覆盖
	if str, ok := nodeInfo.OtherInfo["background"].(string); ok {
		background = str
	}
	if str, ok := nodeInfo.OtherInfo["curFlowItemName"].(string); ok {
		curFlowItemName = str
	}
	associationCharacters := domain.Member{}
	if nodeInfo.AssociationCharacters != nil {
		associationCharacters = *nodeInfo.AssociationCharacters
	}
	quizCharacters := domain.Member{}
	if nodeInfo.QuizCharacters != nil {
		quizCharacters = *nodeInfo.QuizCharacters
	}
	thinkingData := domain.ThinkingData{
		Presets: &domain.ThinkingPresets{
			MeetingId:             meetingId,
			FlowItemId:            flowItemId,
			AssociationCharacters: associationCharacters,
			QuizCharacters:        quizCharacters,
			Background:            background,
			CurFlowItemName:       curFlowItemName,
			ProloguePrompt:        nodeInfo.ProloguePrompt,
			QuizPrompt:            nodeInfo.QuizPrompt,
			QuizRound:             nodeInfo.QuizRound,
			QuizNum:               nodeInfo.QuizNum,
			SummaryCharacters:     *nodeInfo.Characters,
			SummaryPrompt:         nodeInfo.SummaryPrompt,
		},
	}
	if flowItem.Status == "finish" {
		meetingDataItem := meeting.MeetingData[flowItemId]
		chatTree := meetingDataItem.Process.ChatTree
		if chatTree != nil {
			// 读取树
			conversationMap := make(map[string]domain.SimpleConversation)
			thinkingTree := formatThinkingTree(chatTree.Root, conversationMap)
			thinkingData.Thinking = &domain.ThinkingThinking{
				ConversationMap: conversationMap,
				ThinkingTree:    *thinkingTree,
			}
		}
		thinkingData.Summary = &domain.ThinkingSummary{
			SummaryText: meetingDataItem.Output,
		}
	}

	return &thinkingData, nil
}

// 格式化树
func formatThinkingTree(treeNode *datastruct.TreeNode[domain.Conversation], conversationMap map[string]domain.SimpleConversation) *domain.ThinkingTree {
	if treeNode == nil {
		return nil
	}
	conversationMap[treeNode.Data.GetId()] = domain.SimpleConversation{
		Id:       treeNode.Data.GetId(),
		Question: treeNode.Data.Question.Text,
		Answer:   treeNode.Data.Answer.Text,
	}
	thinkingTree := &domain.ThinkingTree{
		Id:        treeNode.Data.GetId(),
		ParentId:  treeNode.Parent,
		Question:  treeNode.Data.Question.Text,
		Answer:    treeNode.Data.Answer.Text,
		Width:     380,
		Height:    62,
		Collapsed: false,
		Children:  []*domain.ThinkingTree{},
	}
	for _, child := range treeNode.Children {
		if qaNode := formatThinkingTree(child, conversationMap); qaNode != nil {
			thinkingTree.Children = append(thinkingTree.Children, qaNode)
		}
	}
	return thinkingTree
}

/**
 * @description: 保存thinking预设信息
 * @param {http.Context} ctx
 * @return {*}
 */
func (m *MeetingUsecase) SaveThinkingPresets(ctx context.Context, presets *domain.ThinkingPresets) (interface{}, error) {
	meetingId := presets.MeetingId
	flowItemId := presets.FlowItemId
	_, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return nil, err
	}
	// 调整后的信息保存至flowData
	flowItem.NodeInfo.AssociationCharacters = &presets.AssociationCharacters
	flowItem.NodeInfo.QuizCharacters = &presets.QuizCharacters
	flowItem.NodeInfo.ProloguePrompt = presets.ProloguePrompt
	flowItem.NodeInfo.QuizPrompt = presets.QuizPrompt
	flowItem.NodeInfo.QuizRound = presets.QuizRound
	flowItem.NodeInfo.QuizNum = presets.QuizNum
	flowItem.NodeInfo.OtherInfo = map[string]interface{}{
		"background":      presets.Background,
		"curFlowItemName": presets.CurFlowItemName,
	}
	err = m.meetingRepo.UpdateMeetingFlow(ctx, meetingId, flowItemId, *flowItem)
	if err != nil {
		return nil, errors.New("update meeting flow error")
	}

	// 预设信息将作为chatTree的根节点
	conversation := domain.Conversation{
		Id: "thinking", // 根节点ID固定为thinking
		Question: domain.Message{
			Text:     presets.ProloguePrompt,
			Time:     int(time.Now().Unix()),
			TokenCnt: 0,
		},
		Answer: domain.Message{},
	}
	process := domain.Process{
		ChatTree: datastruct.NewTree[domain.Conversation](),
	}
	process.ChatTree.Insert("", conversation)
	meetingDataItem := domain.MeetingDataItem{
		Input:   presets.ProloguePrompt,
		Process: process,
		Output:  "",
	}
	err = m.meetingRepo.UpdateMeetingData(ctx, meetingId, flowItemId, meetingDataItem) // 第一个节点，直接写入
	return true, err
}

/**
 * @description: 思考并提出问题
 * @param {http.Context} ctx
 * @param {*domain.MeetingFlowItemRequest} meetingFlowItem
 * @return {*}
 */
func (m *MeetingUsecase) ThinkAndQuiz(ctx http.Context, meetingFlowItem *domain.MeetingFlowItemRequest) error {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	chatTreeNodeId := meetingFlowItem.ChatTreeNodeId
	chatTreeParentNodeId := meetingFlowItem.ChatTreeParentNodeId
	meeting, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return err
	}
	// 人设 + 上下文
	nodeInfo := flowItem.NodeInfo
	messages := getContext(meeting, flowItem, nodeInfo.QuizCharacters)
	// 当前环节
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("当前环节是[%s], 以下是本环节的历史对话", flowItem.NodeInfo.NodeName),
	})
	// 获取meetingDataItem
	meetingData := meeting.MeetingData
	var meetingDataItem domain.MeetingDataItem
	if _, ok := meetingData[flowItemId]; !ok {
		return fmt.Errorf("[meetingId=%s, flowItemId=%s]meetingDataItem not found", meetingId, flowItemId)
	}
	meetingDataItem = meetingData[flowItemId]
	// 先将节点插入树中, 保证能找到前置消息
	conversation := domain.Conversation{
		Id:       chatTreeNodeId,
		Question: domain.Message{},
		Answer:   domain.Message{},
	}
	meetingDataItem.Process.ChatTree.Insert(chatTreeParentNodeId, conversation)
	meeting.MeetingData[meetingFlowItem.FlowItemId] = meetingDataItem

	if treeMessage, err := getChatTreeMessage(meeting, meetingFlowItem.FlowItemId, meetingFlowItem.ChatTreeNodeId); err == nil {
		messages = append(messages, treeMessage...)
	}
	// 提问prompt
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: flowItem.NodeInfo.QuizPrompt,
	})
	gptReply, err := m.gpt.ChatCompletion(ctx, messages)
	if err != io.EOF {
		return err
	}
	// 更新question
	conversation.Question = domain.Message{
		Text:     gptReply,
		Time:     int(time.Now().Unix()),
		TokenCnt: 0,
	}
	meetingDataItem.Process.ChatTree.Update(chatTreeNodeId, conversation)
	treeNode := meetingDataItem.Process.ChatTree.GetTreeNode(chatTreeNodeId, nil)
	if treeNode == nil {
		return errors.New("treeNode not found")
	}
	// 找到树节点路径
	path, _ := meetingDataItem.Process.ChatTree.FindPath(nil, chatTreeNodeId)
	idPath := []string{}
	for _, item := range path {
		if item.Data.GetId() == chatTreeNodeId { // 因为是插入节点，所以最后一个节点暂时还不存在
			continue
		}
		idPath = append(idPath, item.Data.GetId())
	}
	newCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 需要重新设置超时时长，前面耗时较长会导致超时
	defer cancel()
	err = m.meetingRepo.UpdateMeetingChatTreeNode(newCtx, meetingId, flowItemId, idPath, *treeNode)

	return err
}

/**
 * @description: 思考并回答问题
 * @param {http.Context} ctx
 * @param {*domain.MeetingFlowItemRequest} meetingFlowItem
 * @return {*}
 */
func (m *MeetingUsecase) ThinkAndAnswer(ctx http.Context, meetingFlowItem *domain.MeetingFlowItemRequest) error {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	chatTreeNodeId := meetingFlowItem.ChatTreeNodeId
	chatTreeParentNodeId := meetingFlowItem.ChatTreeParentNodeId
	text := meetingFlowItem.Text
	meeting, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return err
	}
	if chatTreeParentNodeId != "" && text != "" { // 说明是主动提问的
		// 先将节点插入树中, 保证能找到前置消息
		meetingDataItem := meeting.MeetingData[flowItemId]
		conversation := domain.Conversation{
			Id: chatTreeNodeId,
			Question: domain.Message{
				Text:     text,
				Time:     int(time.Now().Unix()),
				TokenCnt: 0,
			},
			Answer: domain.Message{},
		}
		meetingDataItem.Process.ChatTree.Insert(chatTreeParentNodeId, conversation)
		meeting.MeetingData[meetingFlowItem.FlowItemId] = meetingDataItem
		treeNode := meetingDataItem.Process.ChatTree.GetTreeNode(chatTreeNodeId, nil)
		// 找到树节点路径
		path, _ := meetingDataItem.Process.ChatTree.FindPath(nil, chatTreeNodeId)
		idPath := []string{}
		for _, item := range path {
			if item.Data.GetId() == chatTreeNodeId { // 因为是插入节点，所以最后一个节点暂时还不存在
				continue
			}
			idPath = append(idPath, item.Data.GetId())
		}
		// 提问先写入树中
		err = m.meetingRepo.UpdateMeetingChatTreeNode(ctx, meetingId, flowItemId, idPath, *treeNode)
		if err != nil {
			return err
		}
	}
	// 人设 + 上下文
	nodeInfo := flowItem.NodeInfo
	messages := getContext(meeting, flowItem, nodeInfo.AssociationCharacters)
	// 当前环节
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("当前环节是[%s], 以下是本环节的历史对话", flowItem.NodeInfo.NodeName),
	})
	// chatTree上游记录
	if treeMessage, err := getChatTreeMessage(meeting, flowItemId, chatTreeNodeId); err == nil {
		messages = append(messages, treeMessage...)
	}
	messages[len(messages)-1].Content = "请回答：" + messages[len(messages)-1].Content + "(字数尽量控制在100以内)"
	var gptReply string
	gptReply, err = m.gpt.ChatCompletion(ctx, messages)
	if err != io.EOF {
		log.Error(err.Error())
		return err
	}
	if gptReply != "" {
		// 保存
		meetingDataItem := meeting.MeetingData[flowItemId]
		// 更新写入Question
		node := meetingDataItem.Process.ChatTree.GetTreeNode(chatTreeNodeId, nil)
		if node == nil {
			log.Errorf("ThinkingAndAnswer getTreeNode fail: %s", chatTreeNodeId)
			return errors.New("chatTreeNode not found")
		}
		conversation := node.Data
		conversation.Answer = domain.Message{
			Text:     gptReply,
			Time:     int(time.Now().Unix()),
			TokenCnt: 0,
		}
		// meetingDataItem.Process.ChatTree.Update(chatTreeNodeId, conversation)
		// 找到树节点路径
		path, _ := meetingDataItem.Process.ChatTree.FindPath(nil, chatTreeNodeId)
		idPath := []string{}
		for _, item := range path {
			idPath = append(idPath, item.Data.GetId())
		}

		newCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 需要重新设置超时时长，前面耗时较长会导致超时
		defer cancel()
		err = m.meetingRepo.UpdateMeetingChatTreeNodeData(newCtx, meetingId, flowItemId, idPath, conversation)
		return err
	}
	return nil
}

/**
 * @description: 总结
 * @param {http.Context} ctx
 * @param {*domain.MeetingFlowItemRequest} meetingFlowItem
 * @return {*}
 */
func (m *MeetingUsecase) ThinkAndSummary(ctx http.Context, meetingFlowItem *domain.MeetingFlowItemRequest) error {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	characters := meetingFlowItem.Characters
	summaryPrompt := meetingFlowItem.Prompt
	meeting, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return err
	}
	flowItem.NodeInfo.SummaryPrompt = summaryPrompt
	flowItem.NodeInfo.Characters.Description = characters
	// 更新到meetingFlow
	err = m.meetingRepo.UpdateMeetingFlow(ctx, meetingId, flowItemId, *flowItem)
	if err != nil {
		return err
	}
	// 人设 + 上下文
	nodeInfo := flowItem.NodeInfo
	messages := getContext(meeting, flowItem, nodeInfo.Characters)
	// 当前环节 + 本轮记录
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("当前环节是[%s]", flowItem.NodeInfo.NodeName),
	})
	// chatTree格式转为json
	chatTree := meeting.MeetingData[flowItemId].Process.ChatTree
	qaTree := formatQaTree(chatTree.Root)
	b, err := json.Marshal(qaTree)
	if err != nil {
		return err
	}
	qaTreeJson := string(b)
	messages = append(messages, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("本环节的数据格式为Json树结构，一个节点代表一次对答，子节点则表示基于祖父链上所有对答的背景下的对答。例如："+
			"{'question': '围绕本次会议的主题，提出一个点子。', 'answer': '可以给手表增加健康监测功能。', 'children': [{'question': '可以列举下有哪些健康监测功能吗？', 'answer': '例如：心率、血脂等', 'children': []}]}") +
			" 其中children下的节点就是基于父节点对答背景所提出的问题和回答。下面是本环节的所有对答数据：" +
			qaTreeJson,
	})
	// 总结prompt
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: flowItem.NodeInfo.SummaryPrompt,
	})
	gptReply, err := m.gpt.ChatCompletion(ctx, messages)
	if err != io.EOF {
		log.Error(err.Error())
	}
	// 保存
	meetingDataItem := meeting.MeetingData[flowItemId]
	meetingDataItem.Output = gptReply
	newCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 需要重新设置超时时长，前面耗时较长会导致超时
	defer cancel()
	err = m.meetingRepo.UpdateMeetingData(newCtx, meetingId, flowItemId, meetingDataItem)
	if err != nil {
		return err
	}
	return nil
}

// 问答树
type QaTree struct {
	Question string    `json:"question"`
	Answer   string    `json:"answer"`
	Children []*QaTree `json:"children"`
}

// 格式化树
func formatQaTree(treeNode *datastruct.TreeNode[domain.Conversation]) *QaTree {
	if treeNode == nil {
		return nil
	}
	qaTree := &QaTree{
		Question: treeNode.Data.Question.Text,
		Answer:   treeNode.Data.Answer.Text,
		Children: []*QaTree{},
	}
	for _, child := range treeNode.Children {
		if qaNode := formatQaTree(child); qaNode != nil {
			qaTree.Children = append(qaTree.Children, qaNode)
		}
	}
	return qaTree
}

func (m *MeetingUsecase) GetDiscussionData(ctx context.Context, meetingFlowItem *domain.MeetingFlowItemRequest) (*domain.DiscussionData, error) {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	meeting, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return nil, err
	}
	// 背景信息
	messages := getContext(meeting, flowItem, nil)
	background := ""
	for _, message := range messages {
		background += message.Content + "\n"
	}
	nodeInfo := flowItem.NodeInfo
	curFlowItemName := nodeInfo.NodeName
	// 如果存在在覆盖
	if str, ok := nodeInfo.OtherInfo["background"].(string); ok {
		background = str
	}
	if str, ok := nodeInfo.OtherInfo["curFlowItemName"].(string); ok {
		curFlowItemName = str
	}
	memberList := []domain.Member{}
	if nodeInfo.MemberList != nil {
		memberList = nodeInfo.MemberList
	}
	replyRound := nodeInfo.ReplyRound
	summaryCharacters := domain.Member{}
	if nodeInfo.Characters != nil {
		summaryCharacters = *nodeInfo.Characters
	}
	discussionData := domain.DiscussionData{
		Presets: &domain.DiscussionPresets{
			MeetingId:         meetingId,
			FlowItemId:        flowItemId,
			Background:        background,
			CurFlowItemName:   curFlowItemName,
			MemberList:        memberList,
			ProloguePrompt:    nodeInfo.ProloguePrompt,
			ReplyRound:        replyRound,
			SummaryCharacters: summaryCharacters,
			SummaryPrompt:     nodeInfo.SummaryPrompt,
		},
	}
	if flowItem.Status == "finish" {
		meetingDataItem := meeting.MeetingData[flowItemId]
		chatList := meetingDataItem.Process.ChatList
		if chatList != nil {
			discussionData.Discussion = &domain.DiscussionDiscusstion{
				ChatList: chatList,
			}
		}
		discussionData.Summary = &domain.DiscussionSummary{
			SummaryText: meetingDataItem.Output,
		}
	}
	return &discussionData, nil
}

func (m *MeetingUsecase) SaveDiscussionPresets(ctx context.Context, presets *domain.DiscussionPresets) (interface{}, error) {
	meetingId := presets.MeetingId
	flowItemId := presets.FlowItemId
	_, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return nil, err
	}
	// 调整后的信息保存至flowData
	flowItem.NodeInfo.ProloguePrompt = presets.ProloguePrompt // todo 改为discussionPrompt
	flowItem.NodeInfo.SummaryPrompt = presets.SummaryPrompt
	flowItem.NodeInfo.MemberList = presets.MemberList
	flowItem.NodeInfo.ReplyRound = presets.ReplyRound
	flowItem.NodeInfo.OtherInfo = map[string]interface{}{
		"background":      presets.Background,
		"curFlowItemName": presets.CurFlowItemName,
	}
	err = m.meetingRepo.UpdateMeetingFlow(ctx, meetingId, flowItemId, *flowItem)
	if err != nil {
		return nil, errors.New("update meeting flow error")
	}

	// 预设信息作为chatList的第一条记录
	process := domain.Process{
		ChatList: []domain.Message{
			{
				Text: presets.Background + presets.ProloguePrompt,
				Member: domain.Member{
					MemberId:   "assistant",
					MemberName: "assistant",
				},
				Time:     int(time.Now().Unix()),
				TokenCnt: 0,
			},
		},
	}
	meetingDataItem := domain.MeetingDataItem{
		Input:   presets.ProloguePrompt,
		Process: process,
		Output:  "",
	}
	err = m.meetingRepo.UpdateMeetingData(ctx, meetingId, flowItemId, meetingDataItem) // 第一个，直接写入
	return true, err
}

func (m *MeetingUsecase) Discuss(ctx http.Context, meetingFlowItem *domain.MeetingFlowItemRequest) error {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	memberId := meetingFlowItem.MemberId
	meeting, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return err
	}
	// 与会人员
	characters := ""
	curMember := domain.Member{}
	otherMember := ""
	nodeInfo := flowItem.NodeInfo
	otherMemberIndex := 1
	for _, member := range nodeInfo.MemberList {
		if member.MemberId == memberId {
			characters = member.Description
			curMember = member
		} else {
			otherMember += fmt.Sprintf("%d. 名称：%s 特点：%s\n", otherMemberIndex, member.MemberName, member.Description)
			otherMemberIndex += 1
		}
	}
	if characters == "" {
		return errors.New("memberId not found")
	}
	// 人设 + 上下文
	messages := getContext(meeting, flowItem, nil)
	// 当前环节
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("当前环节是[%s], 参与本次讨论除你之外，还有以下成员：\n%s\n", flowItem.NodeInfo.NodeName, otherMember),
	})
	// 历史对话
	if discussMessage, err := getDiscussMessage(meeting, flowItemId); err == nil && len(discussMessage) > 0 {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: "以下是本环节的对话",
		})
		messages = append(messages, discussMessage...)
	}
	// 人设 + 引导prompt
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: fmt.Sprintf("你是%s %s (你不必重复需解释你是谁，只需要按自己的特点回答即可)", characters, flowItem.NodeInfo.ProloguePrompt),
	})
	gptReply, err := m.gpt.ChatCompletion(ctx, messages)
	if err != io.EOF {
		return err
	}
	meetingDataItem := meeting.MeetingData[flowItemId]
	meetingDataItem.Process.ChatList = append(meetingDataItem.Process.ChatList, domain.Message{
		Text:     gptReply,
		Member:   curMember,
		Time:     int(time.Now().Unix()),
		TokenCnt: 0,
	})
	newCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 需要重新设置超时时长，前面耗时较长会导致超时
	defer cancel()
	// 理论上来说这个环节是串行的，所以直接更新即可
	err = m.meetingRepo.UpdateMeetingData(newCtx, meetingId, flowItemId, meetingDataItem)
	return err
}

func (m *MeetingUsecase) DiscussAndQuiz(ctx context.Context, meetingFlowItem *domain.MeetingFlowItemRequest) (interface{}, error) {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	meeting, _, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return false, err
	}
	meetingDataItem := meeting.MeetingData[flowItemId]
	meetingDataItem.Process.ChatList = append(meetingDataItem.Process.ChatList, domain.Message{
		Text: meetingFlowItem.Text,
		Member: domain.Member{
			MemberId:   "user",
			MemberName: "user",
		},
		Time:     int(time.Now().Unix()),
		TokenCnt: 0,
	})
	err = m.meetingRepo.UpdateMeetingData(ctx, meetingId, flowItemId, meetingDataItem)
	return true, err
}

func (m *MeetingUsecase) DiscussAndSummary(ctx http.Context, meetingFlowItem *domain.MeetingFlowItemRequest) error {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	characters := meetingFlowItem.Characters
	summaryPrompt := meetingFlowItem.Prompt
	meeting, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return err
	}
	flowItem.NodeInfo.SummaryPrompt = summaryPrompt
	flowItem.NodeInfo.Characters.Description = characters
	// 更新到meetingFlow
	err = m.meetingRepo.UpdateMeetingFlow(ctx, meetingId, flowItemId, *flowItem)
	if err != nil {
		return err
	}
	// 人设 + 上下文
	nodeInfo := flowItem.NodeInfo
	messages := getContext(meeting, flowItem, nodeInfo.Characters)
	// 当前环节
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("当前环节是[%s], 以下是本环节的对话", flowItem.NodeInfo.NodeName),
	})
	// 历史对话
	if discussMessage, err := getDiscussMessage(meeting, flowItemId); err != nil {
		messages = append(messages, discussMessage...)
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: flowItem.NodeInfo.SummaryPrompt,
	})
	gptReply, err := m.gpt.ChatCompletion(ctx, messages)
	if err != io.EOF {
		log.Error(err.Error())
		return err
	}
	// 保存
	meetingDataItem := meeting.MeetingData[flowItemId]
	meetingDataItem.Output = gptReply
	newCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 需要重新设置超时时长，前面耗时较长会导致超时
	defer cancel()
	err = m.meetingRepo.UpdateMeetingData(newCtx, meetingId, flowItemId, meetingDataItem)
	if err != nil {
		return err
	}
	return nil
}

func (m *MeetingUsecase) GetProcessingData(ctx context.Context, meetingFlowItem *domain.MeetingFlowItemRequest) (*domain.ProcessingData, error) {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	meeting, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return nil, err
	}
	// 背景信息
	messages := getContext(meeting, flowItem, nil)
	background := ""
	for _, message := range messages {
		background += message.Content + "\n"
	}
	nodeInfo := flowItem.NodeInfo
	curFlowItemName := nodeInfo.NodeName
	prompt := nodeInfo.ProcessingPrompt
	meetingData := meeting.MeetingData
	if _, ok := meetingData[flowItemId]; ok {
		meetingFlowItem := meetingData[flowItemId]
		if meetingFlowItem.Process.Prompt != "" {
			prompt = meetingFlowItem.Process.Prompt
		}
	}
	var output string
	if flowItem.Status == "finish" {
		meetingDataItem := meeting.MeetingData[flowItemId]
		background = meetingDataItem.Input
		output = meetingDataItem.Output
	}
	characters := nodeInfo.Characters

	return &domain.ProcessingData{
		Presets: &domain.ProcessingPresets{
			MeetingId:        meetingId,
			FlowItemId:       flowItemId,
			Characters:       *characters,
			Background:       background,
			CurFlowItemName:  curFlowItemName,
			ProcessingPrompt: prompt,
		},
		Ouput: output,
	}, err
}

func (m *MeetingUsecase) Process(ctx http.Context, meetingFlowItem *domain.MeetingFlowItemRequest) error {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	characters := meetingFlowItem.Characters
	inputData := meetingFlowItem.Text
	processingPrompt := meetingFlowItem.Prompt
	meeting, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return err
	}
	if characters == "" {
		characters = flowItem.NodeInfo.Characters.Description
	}
	if processingPrompt == "" {
		processingPrompt = flowItem.NodeInfo.OptimizationPrompt
	}
	flowItem.NodeInfo.ProcessingPrompt = processingPrompt
	flowItem.NodeInfo.Characters.Description = characters
	// 更新到meetingFlow
	err = m.meetingRepo.UpdateMeetingFlow(ctx, meetingId, flowItemId, *flowItem)
	if err != nil {
		return err
	}
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: characters,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: inputData,
		},
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: processingPrompt,
		},
	}
	gptReply, err := m.gpt.ChatCompletion(ctx, messages)
	if err != io.EOF {
		log.Error(err.Error())
	}
	// 保存
	meetingData := meeting.MeetingData
	var meetingDataItem domain.MeetingDataItem
	if _, ok := meetingData[flowItemId]; ok {
		meetingDataItem = domain.MeetingDataItem{}
	} else {
		meetingDataItem = meetingData[flowItemId]
	}
	meetingDataItem.Input = inputData
	meetingDataItem.Process.Prompt = processingPrompt
	meetingDataItem.Output = gptReply
	newCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 需要重新设置超时时长，前面耗时较长会导致超时
	defer cancel()
	err = m.meetingRepo.UpdateMeetingData(newCtx, meetingId, flowItemId, meetingDataItem)
	return err
}

func (m *MeetingUsecase) GetOutputData(ctx context.Context, meetingFlowItem *domain.MeetingFlowItemRequest) (*domain.OutputData, error) {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	meeting, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return nil, err
	}
	nodeInfo := flowItem.NodeInfo
	curFlowItemName := nodeInfo.NodeName
	prompt := nodeInfo.SummaryPrompt
	meetingData := meeting.MeetingData
	if _, ok := meetingData[flowItemId]; ok {
		meetingFlowItem := meetingData[flowItemId]
		if meetingFlowItem.Process.Prompt != "" {
			prompt = meetingFlowItem.Process.Prompt
		}
	}
	var output string
	if flowItem.Status == "finish" {
		meetingDataItem := meeting.MeetingData[flowItemId]
		output = meetingDataItem.Output
	}
	return &domain.OutputData{
		Presets: &domain.OutputPresets{
			MeetingId:       meetingId,
			FlowItemId:      flowItemId,
			Characters:      *nodeInfo.Characters,
			CurFlowItemName: curFlowItemName,
			SummaryPrompt:   prompt,
		},
		Output: output,
	}, err
}

func (m *MeetingUsecase) Output(ctx http.Context, meetingFlowItem *domain.MeetingFlowItemRequest) error {
	meetingId := meetingFlowItem.MeetingId
	flowItemId := meetingFlowItem.FlowItemId
	characters := meetingFlowItem.Characters
	summaryPrompt := meetingFlowItem.Prompt
	meeting, flowItem, err := m.getBaseInfo(ctx, meetingId, flowItemId)
	if err != nil {
		return err
	}
	if characters == "" {
		characters = flowItem.NodeInfo.Characters.Description
	}
	meetingData := meeting.MeetingData
	var meetingDataItem domain.MeetingDataItem
	if _, ok := meetingData[flowItemId]; ok {
		meetingDataItem = domain.MeetingDataItem{}
	} else {
		meetingDataItem = meetingData[flowItemId]
	}
	if summaryPrompt == "" {
		summaryPrompt = flowItem.NodeInfo.SummaryPrompt
	}
	flowItem.NodeInfo.SummaryPrompt = summaryPrompt
	flowItem.NodeInfo.Characters.Description = characters
	// 更新到meetingFlow
	err = m.meetingRepo.UpdateMeetingFlow(ctx, meetingId, flowItemId, *flowItem)
	if err != nil {
		return err
	}
	// 人设 + 上下文
	nodeInfo := flowItem.NodeInfo
	messages := getContext(meeting, flowItem, nodeInfo.Characters)
	// 当前环节
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("当前环节是[%s]", flowItem.NodeInfo.NodeName),
	})
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: summaryPrompt,
	})
	gptReply, err := m.gpt.ChatCompletion(ctx, messages)
	if err != io.EOF {
		log.Error(err.Error())
	}
	// 保存
	meetingDataItem.Input = ""
	meetingDataItem.Output = gptReply
	newCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 需要重新设置超时时长，前面耗时较长会导致超时
	defer cancel()
	err = m.meetingRepo.UpdateMeetingData(newCtx, meetingId, flowItemId, meetingDataItem)
	return err
}

func (m *MeetingUsecase) Chat(ctx http.Context, content string) error {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: content,
		},
	}
	_, err := m.gpt.ChatCompletion(ctx, messages)
	return err
}

/**
 * @description: 获取基础信息
 * @param {context.Context} ctx
 * @param {*} meetingId
 * @param {string} flowItemId
 * @return {*}
 */
func (m *MeetingUsecase) getBaseInfo(ctx context.Context, meetingId, flowItemId string) (*domain.Meeting, *domain.FlowItem, error) {
	if meetingId == "" || flowItemId == "" {
		return nil, nil, errors.New("参数错误")
	}
	meeting, err := m.meetingRepo.FindOne(ctx, meetingId)
	if err != nil {
		return nil, nil, errors.New("meetingId not found")
	}
	// 当前环节
	flowItem := getFlowItem(meeting.MeetingFlow, flowItemId)
	if flowItem == nil {
		return nil, nil, errors.New("flowItemId not found")
	}
	// 前置环节上下文
	return meeting, flowItem, nil
}

/**
 * @description: 获取flowItem
 * @param {domain.MeetingFlow} meetingFlow
 * @param {string} flowItemId
 * @return {*}
 */
func getFlowItem(meetingFlow domain.MeetingFlow, flowItemId string) *domain.FlowItem {
	for _, flowItem := range meetingFlow {
		if flowItem.NodeInfo.GetId() == flowItemId {
			return &flowItem
		}
	}
	return nil
}

/**
 * @description: 上游议程概要
 * @param {*domain.Meeting} meeting
 * @param {*domain.FlowItem} flowItem
 * @return {*}
 */
func getPreviouslyOn(meeting *domain.Meeting, flowItem *domain.FlowItem) (string, bool) {
	previouslyOn := ""
	meetingData := meeting.MeetingData
	upstream := flowItem.Upstream
	utils.Reverse(upstream)
	for index, itemId := range upstream {
		if itemData, ok := meetingData[itemId]; ok {
			item := getFlowItem(meeting.MeetingFlow, itemId)
			if item.NodeInfo.NodeType == "Input" { // 跳过Input
				continue
			}
			previouslyOn += fmt.Sprintf("---\n议程%d\n议题: %s\n结论: %s\n", index, item.NodeInfo.NodeName, strings.Trim(itemData.Output, "\n"))
		}
	}
	if previouslyOn != "" {
		previouslyOn = "已完成的议程如下，供参考(议程之间使用'---'分割)：\n" + previouslyOn + "---\n"
		return previouslyOn, true
	} else {
		return "", false
	}
}

/**
 * @description: 获取聊天树的前置聊天记录
 * @param {*domain.Meeting} meeting
 * @param {string} flowItemId
 * @param {string} chatTreeNodeId
 * @return {*}
 */
func getChatTreeMessage(meeting *domain.Meeting, flowItemId string, chatTreeNodeId string) ([]openai.ChatCompletionMessage, error) {
	messages := []openai.ChatCompletionMessage{}
	meetingData := meeting.MeetingData
	if meetingDataItem, ok := meetingData[flowItemId]; ok {
		chatTree := meetingDataItem.Process.ChatTree
		chatPath, found := chatTree.FindPath(nil, chatTreeNodeId)
		if found {
			for _, chatTreeNode := range chatPath {
				if chatTreeNode.Data.Question.Text != "" {
					messages = append(messages, openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleUser,
						Content: chatTreeNode.Data.Question.Text,
					})
				}
				if chatTreeNode.Data.Answer.Text != "" {
					messages = append(messages, openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: chatTreeNode.Data.Answer.Text,
					})
				}
			}
		}
	}
	return messages, nil
}

/**
 * @description: 获取历史讨论记录
 * @param {*domain.Meeting} meeting
 * @param {string} flowItemId
 * @return {*}
 */
func getDiscussMessage(meeting *domain.Meeting, flowItemId string) ([]openai.ChatCompletionMessage, error) {
	messages := []openai.ChatCompletionMessage{}
	meetingDataItem := meeting.MeetingData[flowItemId]
	chatList := meetingDataItem.Process.ChatList
	for index, chat := range chatList {
		if index == 0 { // 第一条不要，只是前端看的
			continue
		}
		role := openai.ChatMessageRoleAssistant
		if chat.Member.MemberId == "user" {
			role = openai.ChatMessageRoleUser
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: fmt.Sprintf("===\n与会成员：%s\n发言内容：%s\n", chat.Member.MemberName, chat.Text),
		})
	}
	return messages, nil
}

/**
 * @description: 获取上下文
 * @param {*domain.Meeting} meeting
 * @param {*domain.FlowItem} curFlowItem
 * @param {*domain.Member} member
 * @return {*}
 */
func getContext(meeting *domain.Meeting, curFlowItem *domain.FlowItem, member *domain.Member) []openai.ChatCompletionMessage {
	var messages []openai.ChatCompletionMessage
	// curNode := curFlowItem.NodeInfo
	// 1. 人设characters
	if member != nil {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: member.Description,
		})
	}
	// 背景说明
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "你正在参加一场会议（讨论会）",
	})
	// 2. 会议名 meetingName
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("会议（讨论会）名称是：%s", meeting.Name),
	})
	// 3. 主题&目标 topicGoal
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("主题和目标是：%s", meeting.TopicGoal),
	})
	// 4. 议程 meetingFlow
	agenda := "会议议程表如下:\n"
	meetingFlow := meeting.MeetingFlow
	for index, meetingFlowItem := range meetingFlow {
		agenda += fmt.Sprintf("议程%d: %s\n", index, meetingFlowItem.NodeInfo.NodeName)
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: agenda,
	})
	// 5. 前驱环节概要 previouslyOn
	previouslyOn, found := getPreviouslyOn(meeting, curFlowItem)
	if found {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: previouslyOn,
		})
	}
	// 6. 当前环节 curFlowItem -> 这部分放在外部，因为各个环节各不相同
	//   a. Thinking 需要获取树路径记录
	//   b. Discussion 需要获取历史记录
	return messages
}
