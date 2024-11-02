package service

import (
	"gpt-meeting-service/internal/biz"
	"gpt-meeting-service/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type DifyService struct {
	du  *biz.DifyUsecase
	log *log.Helper
}

func NewDifyService(du *biz.DifyUsecase, logger log.Logger) *DifyService {
	return &DifyService{
		du:  du,
		log: log.NewHelper(logger),
	}
}

type DifyId struct {
	TemplateId string `json:"templateId"`
}

// 分享dify模板
func (ds *DifyService) Share(ctx http.Context) error {
	var in domain.DifyData
	if err := ctx.Bind(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	ds.log.Debug(in)
	reply, err := ds.du.Create(ctx, &in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	return ctx.Result(200, Resp(200, "success", reply))
}

func (ds *DifyService) Search(ctx http.Context) error {
	var in domain.DifySearchReq
	if err := ctx.Bind(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	ds.log.Debug(in)
	reply, err := ds.du.Search(ctx, &in)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	return ctx.Result(200, Resp(200, "success", reply))
}

func (ds *DifyService) IncrLike(ctx http.Context) error {
	var in DifyId
	if err := ctx.Bind(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	ds.log.Debug(in)
	reply, err := ds.du.IncrLike(ctx, in.TemplateId)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	return ctx.Result(200, Resp(200, "success", reply))
}

func (ds *DifyService) IncrDislike(ctx http.Context) error {
	var in DifyId
	if err := ctx.Bind(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	ds.log.Debug(in)
	reply, err := ds.du.IncrDislike(ctx, in.TemplateId)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	return ctx.Result(200, Resp(200, "success", reply))
}

func (ds *DifyService) IncrDownload(ctx http.Context) error {
	var in DifyId
	if err := ctx.Bind(&in); err != nil {
		return ctx.JSON(200, Resp(400, err.Error(), nil))
	}
	ds.log.Debug(in)
	reply, err := ds.du.IncrDownload(ctx, in.TemplateId)
	if err != nil {
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	return ctx.Result(200, Resp(200, "success", reply))
}
