/*
 * @Descripttion:
 * @version:
 * @Date: 2023-05-20 19:42:14
 * @LastEditTime: 2023-07-09 21:11:56
 */
package domain

import "gpt-meeting-service/internal/lib/datastruct"

// 状态
type Status string

const (
	Wait       Status = "wait"
	Processing Status = "processing"
	Finish     Status = "finish"
	Success    Status = "success"
	Error      Status = "error"
)

// 整个任务的当前状态
const (
	Idle    Status = "idle"
	Running Status = "running"
	Done    Status = "done"
)

// 会议
type Meeting struct {
	Id          string      `json:"id" bson:"_id,omitempty"`
	TemplateId  string      `json:"templateId" bson:"templateId"`
	Name        string      `json:"name" bson:"name"`
	TopicGoal   string      `json:"topicGoal" bson:"topicGoal"`
	OverView    string      `json:"overview" bson:"overview"`
	Conclusion  string      `json:"conclusion" bson:"conclusion"`
	Status      Status      `json:"status"`
	MeetingFlow MeetingFlow `json:"meetingFlow" bson:"meetingFlow"`
	MeetingData MeetingData `json:"meetingData" bson:"meetingData,omitempty"`
	TokenCnt    int64       `json:"tokenCnt" bson:"tokenCnt"`
	CreatedBy   string      `json:"createdBy" bson:"createdBy"`
	CreatedTime int64       `json:"createdTime" bson:"createdTime"`
}

// 会议流程
type MeetingFlow []FlowItem

type FlowItem struct {
	NodeInfo MeetingNode `json:"nodeInfo" bson:"nodeInfo"`
	Upstream []string    `json:"upstream" bson:"upstream"`
	Status   Status      `json:"status" bson:"status"`
}

// 会议数据
type MeetingData map[string]MeetingDataItem

type MeetingDataItem struct {
	Input   string  `json:"input"`
	Process Process `json:"process"`
	Output  string  `json:"output"`
}

// 过程数据
type Process struct {
	Prompt   string                         `bson:"prompt,omitempty"`
	ChatList []Message                      `bson:"chatList,omitempty"`
	ChatTree *datastruct.Tree[Conversation] `bson:"chatTree,omitempty"`
}

// 一次对话
type Conversation struct {
	Id       string  `bson:"id"`
	Question Message `bson:"question"`
	Answer   Message `bson:"answer"`
}

func (c Conversation) GetId() string {
	return c.Id
}

type SimpleConversation struct {
	Id       string `json:"id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type ThinkingTree struct {
	Id        string          `json:"id"`
	ParentId  string          `json:"parentId"`
	Question  string          `json:"question"`
	Answer    string          `json:"answer"`
	Width     int32           `json:"width"`
	Height    int32           `json:"height"`
	Collapsed bool            `json:"collapsed"`
	Children  []*ThinkingTree `json:"children"`
}

type Message struct {
	Text     string `bson:"text" json:"text"`
	Member   Member `bson:"member" json:"member"`
	Time     int    `bson:"time" json:"time"`
	TokenCnt int    `bson:"tokenCnt" json:"tokenCnt"`
}

type NewMeetingRequest struct {
	Name       string `json:"name"`
	TemplateId string `json:"templateId"`
	CreatedBy  string `json:"createdBy"`
}

type NewMeetingResponse struct {
	MeetingId   string      `json:"meetingId"`
	MeetingFlow MeetingFlow `json:"meetingFlow"`
}

type HistoryMeetingRequest struct {
	PageNum   int64  `json:"pageNum"`
	PageSize  int64  `json:"pageSize"`
	CreatedBy string `json:"createdBy"`
}

type HistoryMeetingResponse struct {
	Total int64      `json:"total"`
	Data  *[]Meeting `json:"data"`
}

type ProcessResponse struct {
	MeetingId    string     `json:"meetingId"`
	MeetingName  string     `json:"meetingName"`
	NowStatus    Status     `json:"nowStatus"`
	NowFlowItem  *FlowItem  `json:"nowFlowItem"`
	NextFlowItem *FlowItem  `json:"nextFlowItem"`
	ActivedNum   int        `json:"activedNum"`
	Progress     []FlowItem `json:"progress"`
}

type UpdateStatusRequest struct {
	MeetingId  string `json:"meetingId"`
	FlowItemId string `json:"flowItemId"`
}

// Thinking预设信息
type ThinkingPresets struct {
	MeetingId             string `json:"meetingId"`
	FlowItemId            string `json:"flowItemId"`
	AssociationCharacters Member `json:"associationCharacters"`
	QuizCharacters        Member `json:"quizCharacters"`
	Background            string `json:"background"`
	CurFlowItemName       string `json:"curFlowItemName"`
	ProloguePrompt        string `json:"prologuePrompt"`
	QuizPrompt            string `json:"quizPrompt"`
	QuizRound             int32  `json:"quizRound"`
	QuizNum               int32  `json:"quizNum"`
	SummaryCharacters     Member `json:"summaryCharacters"`
	SummaryPrompt         string `json:"summaryPrompt"`
}

type ThinkingThinking struct {
	ConversationMap map[string]SimpleConversation `json:"conversationMap"`
	ThinkingTree    ThinkingTree                  `json:"thinkingTree"`
}

type ThinkingSummary struct {
	SummaryText string `json:"summaryText"`
}

type ThinkingData struct {
	Presets  *ThinkingPresets  `json:"presets"`
	Thinking *ThinkingThinking `json:"thinking"`
	Summary  *ThinkingSummary  `json:"summary"`
}

type DiscussionPresets struct {
	MeetingId         string   `json:"meetingId"`
	FlowItemId        string   `json:"flowItemId"`
	Background        string   `json:"background"`
	CurFlowItemName   string   `json:"curFlowItemName"`
	MemberList        []Member `json:"memberList"`
	ProloguePrompt    string   `json:"prologuePrompt"`
	ReplyRound        int32    `json:"replyRound"`
	SummaryCharacters Member   `json:"summaryCharacters"`
	SummaryPrompt     string   `json:"summaryPrompt"`
}

type DiscussionDiscusstion struct {
	ChatList []Message `json:"chatList"`
}

type DiscussionSummary struct {
	SummaryText string `json:"summaryText"`
}

type DiscussionData struct {
	Presets    *DiscussionPresets     `json:"presets"`
	Discussion *DiscussionDiscusstion `json:"discussion"`
	Summary    *DiscussionSummary     `json:"summary"`
}

type ProcessingPresets struct {
	MeetingId        string `json:"meetingId"`
	FlowItemId       string `json:"flowItemId"`
	Characters       Member `json:"characters"`
	Background       string `json:"background"`
	CurFlowItemName  string `json:"curFlowItemName"`
	ProcessingPrompt string `json:"processingPrompt"`
}

type ProcessingData struct {
	Presets *ProcessingPresets `json:"presets"`
	Ouput   string             `json:"output"`
}

type InputPresets struct {
	MeetingId          string `json:"meetingId"`
	FlowItemId         string `json:"flowItemId"`
	Characters         Member `json:"characters"`
	CurFlowItemName    string `json:"curFlowItemName"`
	OptimizationPrompt string `json:"optimizationPrompt"`
}

type InputData struct {
	Presets *InputPresets `json:"presets"`
	Input   string        `json:"input"`
	Output  string        `json:"output"`
}

type OutputPresets struct {
	MeetingId       string `json:"meetingId"`
	FlowItemId      string `json:"flowItemId"`
	Characters      Member `json:"characters"`
	CurFlowItemName string `json:"curFlowItemName"`
	SummaryPrompt   string `json:"summaryPrompt"`
}

type OutputData struct {
	Presets *OutputPresets `json:"presets"`
	Output  string         `json:"output"`
}

type MeetingFlowItemRequest struct {
	MeetingId            string `json:"meetingId"`
	FlowItemId           string `json:"flowItemId"`
	ChatTreeParentNodeId string `json:"chatTreeParentNodeId,omitempty"` // 父节点ID
	ChatTreeNodeId       string `json:"chatTreeNodeId,omitempty"`       // 思考树节点ID
	MemberId             string `json:"memberId,omitempty"`             // 成员ID
	Text                 string `json:"text,omitempty"`                 // 输入的文本
	Characters           string `json:"characters,omitempty"`           // ai人设
	Prompt               string `json:"prompt,omitempty"`               // prompt
}

type MeetingFlowItemReponse struct {
}
