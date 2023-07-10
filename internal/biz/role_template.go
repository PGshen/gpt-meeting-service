/*
 * @Descripttion:
 * @version:
 * @Date: 2023-04-29 23:29:27
 * @LastEditTime: 2023-06-22 16:16:55
 */
package biz

import (
	"context"
	"gpt-meeting-service/internal/domain"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type RoleTemplateRepo interface {
	CreateRoleTemplate(context.Context, *domain.RoleTemplate) (bool, error)
	UpdateRoleTemplate(context.Context, *domain.RoleTemplate) (bool, error)
	DeleteRoleTemplate(context.Context, string) (bool, error)
	List(context.Context, domain.ListCondition) (*domain.RoleTemplateList, int64, error)
	FindOne(context.Context, string) (*domain.RoleTemplate, error)
}

type RoleTemplateUsecase struct {
	repo RoleTemplateRepo
	log  *log.Helper
}

func NewRoleTemplateUsecase(repo RoleTemplateRepo, logger log.Logger) *RoleTemplateUsecase {
	return &RoleTemplateUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (rc *RoleTemplateUsecase) Create(ctx context.Context, role *domain.RoleTemplate) (bool, error) {
	role.CreatedTime = time.Now().Unix()
	return rc.repo.CreateRoleTemplate(ctx, role)
}

func (rc *RoleTemplateUsecase) Update(ctx context.Context, role *domain.RoleTemplate) (bool, error) {
	return rc.repo.UpdateRoleTemplate(ctx, role)
}

func (rc *RoleTemplateUsecase) Delete(ctx context.Context, id string) (bool, error) {
	return rc.repo.DeleteRoleTemplate(ctx, id)
}

func (rc *RoleTemplateUsecase) Find(ctx context.Context, condition *domain.ListCondition) (*domain.RoleTemplateList, int64, error) {
	return rc.repo.List(ctx, *condition)
}
