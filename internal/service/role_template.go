/*
 * @Descripttion:
 * @version:
 * @Date: 2023-04-29 23:26:09
 * @LastEditTime: 2023-07-09 10:28:15
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

type RoleTemplateService struct {
	pb.UnimplementedRoleServer

	rc  *biz.RoleTemplateUsecase
	log *log.Helper
}

func NewRoleTemplateService(rc *biz.RoleTemplateUsecase, logger log.Logger) *RoleTemplateService {
	return &RoleTemplateService{
		rc:  rc,
		log: log.NewHelper(logger),
	}
}

func (s *RoleTemplateService) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.BoolReply, error) {
	ok, err := s.rc.Create(ctx, &domain.RoleTemplate{
		Avatar:      req.Avatar,
		Summary:     req.Summary,
		Description: req.Description,
		Example:     req.Example,
		CreatedBy:   req.CreatedBy,
	})
	return RespBool(ok, err)
}

func (s *RoleTemplateService) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.BoolReply, error) {
	xid, _ := primitive.ObjectIDFromHex(req.Id)
	ok, err := s.rc.Update(ctx, &domain.RoleTemplate{
		Id:          xid,
		Avatar:      req.Avatar,
		Summary:     req.Summary,
		Description: req.Description,
		Example:     req.Example,
	})
	return RespBool(ok, err)
}

func (s *RoleTemplateService) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*pb.BoolReply, error) {
	ok, err := s.rc.Delete(ctx, req.Id)
	return RespBool(ok, err)
}

func (s *RoleTemplateService) ListRole(ctx context.Context, req *pb.ListRoleRequest) (*pb.ListRoleReply, error) {
	roleList, total, err := s.rc.Find(ctx, &domain.ListCondition{
		Id:       req.Id,
		Summary:  req.Summary,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return &pb.ListRoleReply{
			Code: 400,
			Msg:  err.Error(),
		}, err
	} else {
		var data []*pb.RoleInfo
		for _, role := range *roleList {
			roleResp := pb.RoleInfo{
				Id:          role.Id.Hex(),
				Avatar:      role.Avatar,
				Summary:     role.Summary,
				Description: role.Description,
				Example:     role.Example,
				StarCount:   role.StarCount,
				CreatedBy:   role.CreatedBy,
				CreatedTime: role.CreatedTime,
			}
			data = append(data, &roleResp)
		}
		return &pb.ListRoleReply{
			Code: 200,
			Msg:  "success",
			Data: &pb.ListRoleReply_Data{
				Total: total,
				Data:  data,
			},
		}, nil
	}
}
