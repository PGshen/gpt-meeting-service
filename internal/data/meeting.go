/*
 * @Descripttion:
 * @version:
 * @Date: 2023-05-23 19:59:49
 * @LastEditTime: 2023-07-02 11:35:21
 */
package data

import (
	"context"
	"fmt"
	"gpt-meeting-service/internal/biz"
	"gpt-meeting-service/internal/domain"
	"gpt-meeting-service/internal/lib/datastruct"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const MEETING_COLLECTION = "meeting"

type meetingRepo struct {
	data *Data
	log  *log.Helper
}

func NewMeetingRepo(data *Data, logger log.Logger) biz.MeetingRepo {
	return &meetingRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Create implements biz.MeetingRepo
func (mr *meetingRepo) Create(ctx context.Context, m *domain.Meeting) (string, error) {
	collection := mr.data.mongodb.Collection(MEETING_COLLECTION)
	insertOneRet, err := collection.InsertOne(ctx, m)
	if err != nil {
		mr.log.Errorf("failed insert: %v", err)
		return "", err
	}
	mr.log.Infof("success insert: %v", insertOneRet)
	return insertOneRet.InsertedID.(primitive.ObjectID).Hex(), nil
}

// FindOne implements biz.MeetingRepo
func (mr *meetingRepo) FindOne(ctx context.Context, id string) (*domain.Meeting, error) {
	var result domain.Meeting
	collection := mr.data.mongodb.Collection(MEETING_COLLECTION)
	xid, _ := primitive.ObjectIDFromHex(id)
	singleResult := collection.FindOne(ctx, bson.D{{Key: "_id", Value: xid}})
	if err := singleResult.Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// List implements biz.MeetingRepo
func (mr *meetingRepo) List(ctx context.Context, pageNum, pageSize int64, createdBy string) (*[]domain.Meeting, int64, error) {
	collection := mr.data.mongodb.Collection(MEETING_COLLECTION)
	offset, limit := paginate(pageNum, pageSize)
	var condBm bson.M = bson.M{}
	if createdBy != "" {
		condBm["createdBy"] = createdBy
	}
	projection := bson.M{"meetingFlow": 0, "meetingData": 0}
	sortRule := bson.M{"createdTime": -1}
	findOpts := options.Find().SetProjection(projection).SetSkip(offset).SetLimit(limit).SetSort(sortRule)
	findCursor, err := collection.Find(ctx, condBm, findOpts)
	if err != nil {
		mr.log.Errorf("find err: %v", err)
		return nil, 0, err
	}
	var results []domain.Meeting
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

// CountByUser implements biz.MeetingRepo
func (mr *meetingRepo) CountByUser(ctx context.Context, createdBy string) (int64, error) {
	collection := mr.data.mongodb.Collection(MEETING_COLLECTION)
	condBm := bson.M{
		"createdBy": createdBy,
	}
	return collection.CountDocuments(ctx, condBm)
}

// UpdateMeeting implements biz.MeetingRepo
func (mr *meetingRepo) UpdateMeeting(ctx context.Context, meeting *domain.Meeting) error {
	collection := mr.data.mongodb.Collection(MEETING_COLLECTION)
	xid, _ := primitive.ObjectIDFromHex(meeting.Id)
	filter := bson.M{
		"_id": xid,
	}
	updateField := bson.M{}
	if meeting.Name != "" {
		updateField["name"] = meeting.Name
	}
	if meeting.TopicGoal != "" {
		updateField["topicGoal"] = meeting.TopicGoal
	}
	if meeting.OverView != "" {
		updateField["overview"] = meeting.OverView
	}
	if meeting.Conclusion != "" {
		updateField["conclusion"] = meeting.Conclusion
	}
	if meeting.Status != "" {
		updateField["status"] = meeting.Status
	}
	update := bson.M{
		"$set": updateField,
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	mr.log.Infof("id = %s, result = %+v", meeting.Id, *result)
	return err
}

// UpdateStatus implements biz.MeetingRepo
func (mr *meetingRepo) UpdateStatus(ctx context.Context, meetingId string, flowItemId string, status domain.Status) error {
	collection := mr.data.mongodb.Collection(MEETING_COLLECTION)
	xid, _ := primitive.ObjectIDFromHex(meetingId)
	filter := bson.M{
		"_id": xid,
	}
	update := bson.M{
		"$set": bson.M{
			"meetingFlow.$[elem].status": status,
		},
	}
	arrayFilters := options.ArrayFilters{
		Filters: []interface{}{bson.M{"elem.nodeInfo._id": flowItemId}},
	}
	result, err := collection.UpdateOne(ctx, filter, update, &options.UpdateOptions{ArrayFilters: &arrayFilters})
	mr.log.Infof("meetingId = %s, flowItemId = %s, result = %+v", meetingId, flowItemId, *result)
	return err
}

// UpdateStatus implements biz.MeetingRepo
func (mr *meetingRepo) UpdateMeetingFlow(ctx context.Context, meetingId string, flowItemId string, flowItem domain.FlowItem) error {
	collection := mr.data.mongodb.Collection(MEETING_COLLECTION)
	xid, _ := primitive.ObjectIDFromHex(meetingId)
	filter := bson.M{
		"_id": xid,
	}
	update := bson.M{
		"$set": bson.M{
			"meetingFlow.$[elem]": flowItem,
		},
	}
	arrayFilters := options.ArrayFilters{
		Filters: []interface{}{bson.M{"elem.nodeInfo._id": flowItemId}},
	}
	result, err := collection.UpdateOne(ctx, filter, update, &options.UpdateOptions{ArrayFilters: &arrayFilters})
	mr.log.Infof("meetingId = %s, flowItemId = %s, result = %+v", meetingId, flowItemId, *result)
	return err
}

// UpdateMeetingData implements biz.MeetingRepo
func (mr *meetingRepo) UpdateMeetingData(ctx context.Context, meetingId string, meetingDataItemId string, meetingDataItem domain.MeetingDataItem) error {
	collection := mr.data.mongodb.Collection(MEETING_COLLECTION)
	xid, _ := primitive.ObjectIDFromHex(meetingId)
	filter := bson.M{
		"_id": xid,
	}
	update := bson.M{
		"$set": bson.M{
			"meetingData." + meetingDataItemId: meetingDataItem,
		},
	}
	upsert := true
	options := &options.UpdateOptions{
		Upsert: &upsert,
	}
	result, err := collection.UpdateOne(ctx, filter, update, options)
	mr.log.Infof("meetingId = %s, meetingDataItemId = %s, result = %+v", meetingId, meetingDataItemId, *result)
	if result.ModifiedCount < 1 {
		mr.log.Errorf("modifiedCount < 1")
	}
	return err
}

// UpdateMeetingChatTreeNodeData implements biz.MeetingRepo
func (mr *meetingRepo) UpdateMeetingChatTreeNode(ctx context.Context, meetingId string, meetingDataItemId string, chatNodeIdPath []string, treeNode datastruct.TreeNode[domain.Conversation]) error {
	collection := mr.data.mongodb.Collection(MEETING_COLLECTION)
	filter, updatePath, filters := getCondition(meetingId, meetingDataItemId, chatNodeIdPath)
	updatePath += ".children"
	update := bson.M{
		"$push": bson.M{
			updatePath: treeNode,
		},
	}
	var result *mongo.UpdateResult
	var err error
	if len(filters) > 0 {
		arrayFilters := options.ArrayFilters{
			Filters: filters,
		}
		result, err = collection.UpdateOne(ctx, filter, update, &options.UpdateOptions{ArrayFilters: &arrayFilters})
	} else {
		result, err = collection.UpdateOne(ctx, filter, update)
	}
	mr.log.Infof("meetingId = %s, meetingDataItemeId = %s, result = %+v", meetingId, meetingDataItemId, *result)
	return err
}

func (mr *meetingRepo) UpdateMeetingChatTreeNodeData(ctx context.Context, meetingId string, meetingDataItemId string, chatNodeIdPath []string, conversation domain.Conversation) error {
	collection := mr.data.mongodb.Collection(MEETING_COLLECTION)
	filter, updatePath, filters := getCondition(meetingId, meetingDataItemId, chatNodeIdPath)
	updatePath += ".data"
	update := bson.M{
		"$set": bson.M{
			updatePath: conversation,
		},
	}
	arrayFilters := options.ArrayFilters{
		Filters: filters,
	}
	result, err := collection.UpdateOne(ctx, filter, update, &options.UpdateOptions{ArrayFilters: &arrayFilters})
	mr.log.Infof("meetingId = %s, meetingDataItemeId = %s, result = %+v", meetingId, meetingDataItemId, *result)
	return err
}

func getCondition(meetingId string, meetingDataItemId string, chatNodeIdPath []string) (filter bson.M, updatePath string, filters []interface{}) {
	xid, _ := primitive.ObjectIDFromHex(meetingId)
	filter = bson.M{
		"_id": xid,
	}
	// 根据chatNodeIdPath找到需要更新的节点
	// 根节点特殊处理
	updatePath = "meetingData." + meetingDataItemId + ".process.chatTree.root"
	if len(chatNodeIdPath) > 1 {
		chatNodeIdPath = chatNodeIdPath[1:]

		for index, chatNodeId := range chatNodeIdPath {
			itemName := "item" + strconv.Itoa(index)
			updatePath += fmt.Sprintf(".children.$[%s]", itemName)
			filters = append(filters, bson.M{fmt.Sprintf("%s.data.id", itemName): chatNodeId})
		}
	}
	return filter, updatePath, filters
}
