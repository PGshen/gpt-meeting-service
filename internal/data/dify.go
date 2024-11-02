package data

import (
	"context"
	"gpt-meeting-service/internal/biz"
	"gpt-meeting-service/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const DIFY_COLLECTION = "dify"

type difyRepo struct {
	data *Data
	log  *log.Helper
}

func NewDifyRepo(data *Data, logger log.Logger) biz.DifyRepo {
	return &difyRepo{data: data, log: log.NewHelper(logger)}
}

// Create implements biz.DifyRepo.
func (dr *difyRepo) Create(ctx context.Context, d *domain.Dify) (string, error) {
	collection := dr.data.mongodb.Collection(DIFY_COLLECTION)
	insertOneRet, err := collection.InsertOne(ctx, d)
	if err != nil {
		dr.log.Errorf("failed insert: %v", err)
		return "", err
	}
	return insertOneRet.InsertedID.(primitive.ObjectID).Hex(), nil
}

// List implements biz.DifyRepo.
func (dr *difyRepo) List(ctx context.Context, searchReq *domain.DifySearchReq) (*[]domain.Dify, int64, error) {
	collection := dr.data.mongodb.Collection(DIFY_COLLECTION)
	offset, limit := paginate(searchReq.PageNum, 6)
	// 查询条件
	var condBm bson.M = bson.M{}
	if searchReq.AppType != "" {
		condBm["appType"] = searchReq.AppType
	}
	if searchReq.Category != "" {
		condBm["category"] = searchReq.Category
	}
	if searchReq.Name != "" {
		condBm["name"] = bson.M{
			"$regex": searchReq.Name,
		}
	}
	// 排序条件
	sortRule := bson.M{}
	switch searchReq.Sort {
	case domain.Like:
		sortRule["likeCnt"] = -1
	case domain.Download:
		sortRule["downloadCnt"] = -1
	case domain.CreateTime:
		sortRule["createTime"] = -1
	default:
		sortRule["likeCnt"] = -1
	}
	projection := bson.M{"likeIps": 0, "downloadIps": 0}
	findOpts := options.Find().SetProjection(projection).SetSkip(offset).SetLimit(limit).SetSort(sortRule)
	findCursor, err := collection.Find(ctx, condBm, findOpts)
	if err != nil {
		dr.log.Errorf("find err: %v", err)
		return nil, 0, err
	}
	var results []domain.Dify
	var total int64
	if err = findCursor.All(ctx, &results); err != nil {
		dr.log.Errorf("decode err: %v", err)
		return nil, 0, err
	}
	if total, err = collection.CountDocuments(ctx, condBm); err != nil {
		dr.log.Errorf("count err: %v", err)
		return nil, 0, err
	}
	// mr.log.Infof("result: %v", results)
	return &results, total, nil
}

func (dr *difyRepo) FindOne(ctx context.Context, id string) (*domain.Dify, error) {
	var result domain.Dify
	collection := dr.data.mongodb.Collection(DIFY_COLLECTION)
	xid, _ := primitive.ObjectIDFromHex(id)
	singleResult := collection.FindOne(ctx, bson.D{{Key: "_id", Value: xid}})
	if err := singleResult.Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (dr *difyRepo) Update(ctx context.Context, dify *domain.Dify) error {
	collection := dr.data.mongodb.Collection(DIFY_COLLECTION)
	xid, _ := primitive.ObjectIDFromHex(dify.Id)
	filter := bson.M{
		"_id": xid,
	}
	updateField := bson.M{}
	if dify.DownloadCnt > 0 {
		updateField["downloadCnt"] = dify.DownloadCnt
	}
	if len(dify.DownloadIps) > 0 {
		updateField["downloadIps"] = dify.DownloadIps
	}
	if dify.LikeCnt > 0 {
		updateField["likeCnt"] = dify.LikeCnt
	}
	if len(dify.LikeIps) > 0 {
		updateField["likeIps"] = dify.LikeIps
	}
	if dify.DislikeCnt > 0 {
		updateField["dislikeCnt"] = dify.DislikeCnt
	}
	if len(dify.DislikeIps) > 0 {
		updateField["dislikeIps"] = dify.DislikeIps
	}
	update := bson.M{
		"$set": updateField,
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	dr.log.Infof("id = %s, result = %+v", dify.Id, *result)
	return err
}

func (dr *difyRepo) Del(ctx context.Context, id string) error {
	return nil
}
