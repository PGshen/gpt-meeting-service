/*
 * @Descripttion:
 * @version:
 * @Date: 2023-05-03 15:20:01
 * @LastEditTime: 2023-07-08 00:11:59
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

const MEETING_TEMPLATE_COLLECTION = "meeting_template"

type meetingTemplateRepo struct {
	data *Data
	log  *log.Helper
}

func NewMeetingTemplateRepo(data *Data, logger log.Logger) biz.MeetingTemplateRepo {
	return &meetingTemplateRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (mr *meetingTemplateRepo) CreateMeetingTemplate(ctx context.Context, mt *domain.MeetingTemplate) (bool, error) {
	collection := mr.data.mongodb.Collection(MEETING_TEMPLATE_COLLECTION)
	insertOneRet, err := collection.InsertOne(ctx, mt)
	if err != nil {
		mr.log.Errorf("failed insert: %v", err)
		return false, err
	}
	mr.log.Infof("success insert: %v", insertOneRet)
	return true, nil
}

func (mr *meetingTemplateRepo) UpdateMeetingTemplate(ctx context.Context, meeting *domain.MeetingTemplate) (bool, error) {
	collection := mr.data.mongodb.Collection(MEETING_TEMPLATE_COLLECTION)
	filter := bson.M{"_id": meeting.Id}
	update := bson.M{"$set": meeting}
	result, err := collection.UpdateOne(ctx, filter, update)
	mr.log.Infof("id = %s, result = %v", meeting.Id, result)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (mr *meetingTemplateRepo) DeleteMeetingTemplate(ctx context.Context, id string) (bool, error) {
	collection := mr.data.mongodb.Collection(MEETING_TEMPLATE_COLLECTION)
	xid, _ := primitive.ObjectIDFromHex(id)
	_, err := collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: xid}})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (mr *meetingTemplateRepo) List(ctx context.Context, condition *domain.ListMeetingTemplateCondition) (*domain.MeetingTemplateList, int64, error) {
	collection := mr.data.mongodb.Collection(MEETING_TEMPLATE_COLLECTION)
	offset, pageSize := paginate(condition.PageNum, condition.PageSize)
	var condBm bson.M = bson.M{}
	if condition.Id != "" {
		condBm["_id"] = condition.Id
	}
	if condition.Name != "" {
		searchRegx := bson.M{"$regex": primitive.Regex{Pattern: condition.Name, Options: "i"}}
		condBm["name"] = searchRegx
	}
	findOpts := options.Find().SetSkip(offset).SetLimit(pageSize)
	findCursor, err := collection.Find(ctx, condBm, findOpts)
	if err != nil {
		mr.log.Errorf("find err: %v", err)
		return nil, 0, err
	}
	var results domain.MeetingTemplateList
	var total int64
	if err = findCursor.All(ctx, &results); err != nil {
		mr.log.Errorf("decode err: %v", err)
		return nil, 0, err
	}
	if total, err = collection.CountDocuments(ctx, condBm); err != nil {
		mr.log.Errorf("count err: %v", err)
		return nil, 0, err
	}
	// mr.log.Infof("result: %v", results)
	return &results, total, nil
}

func (mr *meetingTemplateRepo) FindOne(ctx context.Context, id string) (*domain.MeetingTemplate, error) {
	var result domain.MeetingTemplate
	collection := mr.data.mongodb.Collection(MEETING_TEMPLATE_COLLECTION)
	xid, _ := primitive.ObjectIDFromHex(id)
	singleResult := collection.FindOne(ctx, bson.D{{Key: "_id", Value: xid}})
	if err := singleResult.Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
