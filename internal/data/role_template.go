/*
 * @Descripttion:
 * @version:
 * @Date: 2023-04-29 23:45:36
 * @LastEditTime: 2023-07-08 00:10:32
 */
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

const ROLE_TEMPLATE_COLLECTION = "role_template"

type roleTemplateRepo struct {
	data *Data
	log  *log.Helper
}

// FindOne implements biz.RoleTemplateRepo.
func (mr *roleTemplateRepo) FindOne(ctx context.Context, id string) (*domain.RoleTemplate, error) {
	var result domain.RoleTemplate
	collection := mr.data.mongodb.Collection(ROLE_TEMPLATE_COLLECTION)
	xid, _ := primitive.ObjectIDFromHex(id)
	singleResult := collection.FindOne(ctx, bson.D{{Key: "_id", Value: xid}})
	if err := singleResult.Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func NewRoleTemplateRepo(data *Data, logger log.Logger) biz.RoleTemplateRepo {
	return &roleTemplateRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (rr *roleTemplateRepo) CreateRoleTemplate(ctx context.Context, role *domain.RoleTemplate) (bool, error) {
	collection := rr.data.mongodb.Collection(ROLE_TEMPLATE_COLLECTION)
	insertOneRet, err := collection.InsertOne(ctx, role)
	if err != nil {
		rr.log.Errorf("failed insert: %v", err)
		return false, err
	}
	rr.log.Infof("success insert: %v", insertOneRet)
	return true, nil
}

func (rr *roleTemplateRepo) UpdateRoleTemplate(ctx context.Context, role *domain.RoleTemplate) (bool, error) {
	collection := rr.data.mongodb.Collection(ROLE_TEMPLATE_COLLECTION)
	filter := bson.M{"_id": role.Id}
	update := bson.M{"$set": role}
	result, err := collection.UpdateOne(ctx, filter, update)
	rr.log.Infof("id = %s, result = %v", role.Id, result)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (rr *roleTemplateRepo) DeleteRoleTemplate(ctx context.Context, id string) (bool, error) {
	collection := rr.data.mongodb.Collection(ROLE_TEMPLATE_COLLECTION)
	xid, _ := primitive.ObjectIDFromHex(id)
	_, err := collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: xid}})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (rr *roleTemplateRepo) List(ctx context.Context, condition domain.ListCondition) (*domain.RoleTemplateList, int64, error) {
	collection := rr.data.mongodb.Collection(ROLE_TEMPLATE_COLLECTION)
	offset, pageSize := paginate(condition.PageNum, condition.PageSize)
	var condBm bson.M = bson.M{}
	if condition.Id != "" {
		condBm["_id"] = condition.Id
	}
	if condition.Summary != "" {
		searchRegx := bson.M{"$regex": primitive.Regex{Pattern: condition.Summary, Options: "i"}}
		condBm["summary"] = searchRegx
	}
	findOpts := options.Find().SetSkip(offset).SetLimit(pageSize)
	findCursor, err := collection.Find(ctx, condBm, findOpts)
	if err != nil {
		rr.log.Errorf("find err: %v", err)
		return nil, 0, err
	}
	var results domain.RoleTemplateList
	var total int64
	if err = findCursor.All(ctx, &results); err != nil {
		rr.log.Errorf("decode err: %v", err)
		return nil, 0, err
	}
	if total, err = collection.CountDocuments(ctx, condBm); err != nil {
		rr.log.Errorf("count err: %v", err)
		return nil, 0, err
	}
	// rr.log.Infof("result: %v", results)
	return &results, total, nil
}
