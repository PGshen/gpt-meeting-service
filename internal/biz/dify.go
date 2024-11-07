package biz

import (
	"context"
	"errors"
	"gpt-meeting-service/internal/domain"
	"gpt-meeting-service/internal/lib/utils"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type DifyRepo interface {
	Create(context.Context, *domain.Dify) (string, error)
	List(ctx context.Context, searchReq *domain.DifySearchReq) (*[]domain.Dify, int64, error)
	FindOne(ctx context.Context, id string) (*domain.Dify, error)
	Update(ctx context.Context, dify *domain.Dify) error
}

type DifyUsecase struct {
	difyRepo DifyRepo
	log      *log.Helper
}

func NewDifyUsecase(difyRepo DifyRepo, logger log.Logger) *DifyUsecase {
	return &DifyUsecase{
		difyRepo: difyRepo,
		log:      log.NewHelper(logger),
	}
}

func (d *DifyUsecase) Search(ctx http.Context, condition *domain.DifySearchReq) (*domain.DifyResp, error) {
	difyList, cnt, err := d.difyRepo.List(ctx, condition)
	if err != nil {
		return nil, err
	}
	difyDataList := []domain.DifyData{}
	for _, dify := range *difyList {
		difyDataList = append(difyDataList, domain.DifyData{
			Id:          dify.Id,
			Name:        dify.Name,
			Description: dify.Description,
			Author:      dify.Author,
			AppType:     dify.AppType,
			Category:    dify.Category,
			Yml:         dify.Yml,
			Images:      dify.Images,
			DownloadCnt: dify.DownloadCnt,
			LikeCnt:     dify.LikeCnt,
		})
	}
	resp := domain.DifyResp{
		Cnt:  cnt,
		List: &difyDataList,
	}
	return &resp, nil
}

func (d *DifyUsecase) Create(ctx http.Context, newDify *domain.DifyData) (*domain.DifyEmpty, error) {
	if newDify.Name == "" {
		return nil, errors.New("name is required")
	}
	if newDify.Description == "" {
		return nil, errors.New("descript is required")
	}
	if newDify.AppType == "" {
		return nil, errors.New("appType is required")
	}
	if len(newDify.Category) == 0 {
		return nil, errors.New("category is required")
	}
	if newDify.Yml == "" {
		return nil, errors.New("yml is required")
	}
	if len(newDify.Images) == 0 {
		return nil, errors.New("image is required")
	}
	dify := &domain.Dify{
		Name:        newDify.Name,
		Description: newDify.Description,
		Author:      newDify.Author,
		AppType:     newDify.AppType,
		Category:    newDify.Category,
		Yml:         newDify.Yml,
		Images:      newDify.Images,
		DownloadCnt: 0,
		DownloadIps: []string{},
		LikeCnt:     0,
		LikeIps:     []string{},
		Quality:     1,
		CreateTime:  time.Now().Unix(),
		UpdateTime:  time.Now().Unix(),
		DeleteTime:  0,
	}
	_, err := d.difyRepo.Create(ctx, dify)
	if err != nil {
		return nil, err
	}
	return &domain.DifyEmpty{}, nil
}

func (d *DifyUsecase) IncrLike(ctx http.Context, difyId string) (*domain.DifyEmpty, error) {
	userIp := utils.GetClientIp(ctx.Request())
	dify, err := d.difyRepo.FindOne(ctx, difyId)
	if err != nil {
		return nil, err
	}
	likeIps := dify.LikeIps
	likeIps = append(likeIps, userIp)
	likeIps = utils.RemoveDuplicate(likeIps)
	likeCnt := len(likeIps)
	uParam := &domain.Dify{
		Id:      difyId,
		LikeCnt: int64(likeCnt),
		LikeIps: likeIps,
	}
	err = d.difyRepo.Update(ctx, uParam)
	if err != nil {
		return nil, err
	}
	return &domain.DifyEmpty{}, nil
}

func (d *DifyUsecase) IncrDislike(ctx http.Context, difyId string) (*domain.DifyEmpty, error) {
	userIp := utils.GetClientIp(ctx.Request())
	dify, err := d.difyRepo.FindOne(ctx, difyId)
	if err != nil {
		return nil, err
	}
	dislikeIps := dify.DislikeIps
	dislikeIps = append(dislikeIps, userIp)
	dislikeIps = utils.RemoveDuplicate(dislikeIps)
	dislikeCnt := len(dislikeIps)
	uParam := &domain.Dify{
		Id:         difyId,
		DislikeCnt: int64(dislikeCnt),
		DislikeIps: dislikeIps,
	}
	err = d.difyRepo.Update(ctx, uParam)
	if err != nil {
		return nil, err
	}
	return &domain.DifyEmpty{}, nil
}

func (d *DifyUsecase) IncrDownload(ctx http.Context, difyId string) (*domain.DifyEmpty, error) {
	userIp := utils.GetClientIp(ctx.Request())
	dify, err := d.difyRepo.FindOne(ctx, difyId)
	if err != nil {
		return nil, err
	}
	downloadIps := dify.DownloadIps
	downloadIps = append(downloadIps, userIp)
	downloadIps = utils.RemoveDuplicate(downloadIps)
	downloadCnt := len(downloadIps)
	uParam := &domain.Dify{
		Id:          difyId,
		DownloadCnt: int64(downloadCnt),
		DownloadIps: downloadIps,
	}
	err = d.difyRepo.Update(ctx, uParam)
	if err != nil {
		return nil, err
	}
	return &domain.DifyEmpty{}, nil
}
