/*
 * @Descripttion:
 * @version:
 * @Date: 2023-05-03 14:14:00
 * @LastEditTime: 2023-07-08 00:14:17
 */
package service

import (
	"context"

	pb "gpt-meeting-service/api/template/v1"
	"gpt-meeting-service/internal/biz"
	"gpt-meeting-service/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MeetingTemplateService struct {
	pb.UnimplementedMeetingServer

	mu  *biz.MeetingTemplateUsecase
	log *log.Helper
}

func NewMeetingTemplateService(mu *biz.MeetingTemplateUsecase, logger log.Logger) *MeetingTemplateService {
	return &MeetingTemplateService{
		mu:  mu,
		log: log.NewHelper(logger),
	}
}

func (s *MeetingTemplateService) CreateMeeting(ctx context.Context, req *pb.CreateMeetingRequest) (*pb.BoolReply, error) {
	ok, err := s.mu.Create(ctx, &domain.MeetingTemplate{
		Name:         req.Name,
		Avatar:       req.Avatar,
		Description:  req.Description,
		Example:      req.Example,
		TemplateFlow: req.TemplateFlow,
		TemplateData: req.TemplateData,
		CreatedBy:    req.CreatedBy,
	})
	return RespBool(ok, err)
}

func (s *MeetingTemplateService) UpdateMeeting(ctx context.Context, req *pb.UpdateMeetingRequest) (*pb.BoolReply, error) {
	xid, _ := primitive.ObjectIDFromHex(req.Id)
	ok, err := s.mu.Update(ctx, &domain.MeetingTemplate{
		Id:           xid,
		Name:         req.Name,
		Avatar:       req.Avatar,
		Description:  req.Description,
		Example:      req.Example,
		TemplateFlow: req.TemplateFlow,
		TemplateData: req.TemplateData,
	})
	return RespBool(ok, err)
}

func (s *MeetingTemplateService) DeleteMeeting(ctx context.Context, req *pb.DeleteMeetingRequest) (*pb.BoolReply, error) {
	ok, err := s.mu.Delete(ctx, req.Id)
	return RespBool(ok, err)
}

func (s *MeetingTemplateService) GetMeeting(ctx context.Context, req *pb.GetMeetingRequest) (*pb.GetMeetingReply, error) {
	meetingTemplate, err := s.mu.FindById(ctx, req.Id)
	if err != nil {
		return &pb.GetMeetingReply{
			Code: 400,
			Msg:  err.Error(),
		}, err
	} else {
		return &pb.GetMeetingReply{
			Code: 200,
			Msg:  "success",
			Data: meetingTemplate.MeetingTemplateToPb(),
		}, nil
	}
}

func (s *MeetingTemplateService) ListMeeting(ctx context.Context, req *pb.ListMeetingRequest) (*pb.ListMeetingReply, error) {
	meetingList, total, err := s.mu.Find(ctx, &domain.ListMeetingTemplateCondition{
		Id:       req.Id,
		Name:     req.Name,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return &pb.ListMeetingReply{
			Code: 400,
			Msg:  err.Error(),
		}, err
	} else {
		var data []*pb.MeetingInfo
		for _, meeting := range *meetingList {
			data = append(data, meeting.MeetingTemplateToPb())
		}
		return &pb.ListMeetingReply{
			Code: 200,
			Msg:  "success",
			Data: &pb.ListMeetingReply_Data{
				Total: int64(total),
				Data:  data,
			},
		}, nil
	}
}
