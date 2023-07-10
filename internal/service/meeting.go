/*
 * @Descripttion:
 * @version:
 * @Date: 2023-05-20 17:50:43
 * @LastEditTime: 2023-07-02 17:04:03
 */
package service

import (
	"context"
	"encoding/json"
	"gpt-meeting-service/internal/biz"
	"gpt-meeting-service/internal/domain"
	"io/ioutil"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type Message struct {
	Message string `json:"message"`
}

type MeetingService struct {
	mu  *biz.MeetingUsecase
	log *log.Helper
}

func NewMeetingService(mu *biz.MeetingUsecase, logger log.Logger) *MeetingService {
	return &MeetingService{
		mu:  mu,
		log: log.NewHelper(logger),
	}
}

/**
 * @description: 新建会议
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) NewMeeting(ctx http.Context) error {
	var in domain.NewMeetingRequest
	if err := ctx.Bind(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	ms.log.Debug(in)
	reply, err := ms.mu.NewMeeting(ctx, &in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	return ctx.Result(200, Resp(200, "success", reply))
}

/**
 * @description: 查询历史记录
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) History(ctx http.Context) error {
	var in domain.HistoryMeetingRequest
	if err := ctx.BindQuery(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	ms.log.Debug(in)
	reply, err := ms.mu.HistoryMeeting(ctx, &in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	return ctx.Result(200, Resp(200, "success", reply))
}

/**
 * @description: 查询进度
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) Progress(ctx http.Context) error {
	meeeingId := ctx.Query().Get("meetingId")
	if meeeingId == "" {
		return ctx.JSON(200, Resp(400, "meetingId is empty", nil))
	}
	h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		return ms.mu.Progress(ctx, req.(string))
	})
	out, err := h(ctx, meeeingId)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	reply := out.(*domain.ProcessResponse)
	return ctx.Result(200, Resp(200, "success", reply))
}

/**
 * @description: 开始
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) Start(ctx http.Context) error {
	var in domain.UpdateStatusRequest
	if err := ctx.Bind(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		return ms.mu.StartFlowItem(ctx, in.MeetingId, in.FlowItemId)
	})
	out, err := h(ctx, in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	return ctx.Result(200, Resp(200, "success", out))
}

/**
 * @description: 结束
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) End(ctx http.Context) error {
	var in domain.UpdateStatusRequest
	if err := ctx.Bind(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		return ms.mu.EndFlowItem(ctx, in.MeetingId, in.FlowItemId)
	})
	out, err := h(ctx, in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	return ctx.Result(200, Resp(200, "success", out))
}

/**
 * @description: 获取输入
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) GetInputData(ctx http.Context) error {
	var in domain.MeetingFlowItemRequest
	if err := ctx.BindQuery(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		return ms.mu.GetInputData(ctx, &in)
	})
	out, err := h(ctx, in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	reply := out.(*domain.InputData)
	return ctx.Result(200, Resp(200, "success", reply))
}

/**
 * @description: 输入
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) Input(ctx http.Context) error {
	return meetingLink(ms.mu.Input, ctx)
}

/**
 * @description: 提交输入
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) SubmitInput(ctx http.Context) error {
	var in domain.MeetingFlowItemRequest
	if err := ctx.Bind(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		return ms.mu.SubmitInput(ctx, &in)
	})
	out, err := h(ctx, in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	return ctx.Result(200, Resp(200, "success", out))
}

/**
 * @description: 获取thinking预设
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) GetThinkingData(ctx http.Context) error {
	var in domain.MeetingFlowItemRequest
	if err := ctx.BindQuery(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		return ms.mu.GetThinkingData(ctx, &in)
	})
	out, err := h(ctx, in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	reply := out.(*domain.ThinkingData)
	return ctx.Result(200, Resp(200, "success", reply))
}

/**
 * @description: 提交输入
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) SaveThinkingPresets(ctx http.Context) error {
	var in domain.ThinkingPresets
	if err := ctx.Bind(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		return ms.mu.SaveThinkingPresets(ctx, &in)
	})
	out, err := h(ctx, in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	return ctx.Result(200, Resp(200, "success", out))
}

/**
 * @description: 思考并提问
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) ThinkAndQuiz(ctx http.Context) error {
	return meetingLink(ms.mu.ThinkAndQuiz, ctx)
}

/**
 * @description: 思考并回答
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) ThinkAndAnswer(ctx http.Context) error {
	return meetingLink(ms.mu.ThinkAndAnswer, ctx)
}

/**
 * @description: thinking总结
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) ThinkAndSummary(ctx http.Context) error {
	return meetingLink(ms.mu.ThinkAndSummary, ctx)
}

/**
 * @description: 获取thinking预设
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) GetDiscussionData(ctx http.Context) error {
	var in domain.MeetingFlowItemRequest
	if err := ctx.BindQuery(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		return ms.mu.GetDiscussionData(ctx, &in)
	})
	out, err := h(ctx, in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	reply := out.(*domain.DiscussionData)
	return ctx.Result(200, Resp(200, "success", reply))
}

/**
 * @description: 提交输入
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) SaveDiscussionPresets(ctx http.Context) error {
	var in domain.DiscussionPresets
	if err := ctx.Bind(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		return ms.mu.SaveDiscussionPresets(ctx, &in)
	})
	out, err := h(ctx, in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	return ctx.Result(200, Resp(200, "success", out))
}

/**
 * @description: 多人讨论
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) Discuss(ctx http.Context) error {
	return meetingLink(ms.mu.Discuss, ctx)
}

/**
 * @description: 提交输入
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) DiscussAndQuiz(ctx http.Context) error {
	var in domain.MeetingFlowItemRequest
	if err := ctx.Bind(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		return ms.mu.DiscussAndQuiz(ctx, &in)
	})
	out, err := h(ctx, in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	return ctx.Result(200, Resp(200, "success", out))
}

/**
 * @description: thinking总结
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) DiscussAndSummary(ctx http.Context) error {
	return meetingLink(ms.mu.DiscussAndSummary, ctx)
}

/**
 * @description: 通用
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) Process(ctx http.Context) error {
	return meetingLink(ms.mu.Process, ctx)
}

func (ms *MeetingService) GetProcessingData(ctx http.Context) error {
	var in domain.MeetingFlowItemRequest
	if err := ctx.BindQuery(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		return ms.mu.GetProcessingData(ctx, &in)
	})
	out, err := h(ctx, in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	reply := out.(*domain.ProcessingData)
	return ctx.Result(200, Resp(200, "success", reply))
}

/**
 * @description: 输出
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) Output(ctx http.Context) error {
	return meetingLink(ms.mu.Output, ctx)
}

func (ms *MeetingService) GetOutputData(ctx http.Context) error {
	var in domain.MeetingFlowItemRequest
	if err := ctx.BindQuery(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		return ms.mu.GetOutputData(ctx, &in)
	})
	out, err := h(ctx, in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	reply := out.(*domain.OutputData)
	return ctx.Result(200, Resp(200, "success", reply))
}

func meetingLink(bizDeal func(ctx http.Context, req *domain.MeetingFlowItemRequest) error, ctx http.Context) error {
	// 设置流式响应
	setStreamHeader(ctx.Response())
	// 参数解析
	var in domain.MeetingFlowItemRequest
	req := ctx.Request()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, &in); err != nil {
		return err
	}
	return bizDeal(ctx, &in)
}

/**
 * @description: chat
 * @param {http.Context} ctx
 * @return {*}
 */
func (ms *MeetingService) Chat(ctx http.Context) error {
	setStreamHeader(ctx.Response())
	// fmt.Println("----")
	req := ctx.Request()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	var in Message
	if err = json.Unmarshal(body, &in); err != nil {
		return err
	}
	log.Info("message: " + in.Message)
	return ms.mu.Chat(ctx, in.Message)
}

func (ms *MeetingService) MeetingOptions(ctx http.Context) error {
	resp := ctx.Response()
	resp.Header().Set("Access-Control-Allow-Origin", "*")
	resp.Header().Set("Access-Control-Allow-Methods", "*")
	resp.Header().Set("Access-Control-Allow-Headers", "*")
	return nil
}

/**
 * @description: 设置流式响应
 * @param {http.ResponseWriter} resp
 * @return {*}
 */
func setStreamHeader(resp http.ResponseWriter) {
	resp.Header().Set("Content-Type", "text/event-stream")
	resp.Header().Set("Cache-Control", "no-cache")
	resp.Header().Set("Connection", "keep-alive")
	resp.Header().Set("Access-Control-Allow-Origin", "*")
	resp.Header().Set("Access-Control-Allow-Methods", "*")
	resp.Header().Set("Access-Control-Allow-Headers", "*")
}
