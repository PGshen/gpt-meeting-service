/*
 * @Descripttion:
 * @version:
 * @Date: 2023-04-29 22:30:30
 * @LastEditTime: 2023-04-29 23:28:30
 */
package server

import (
	v1 "gpt-meeting-service/api/template/v1"
	"gpt-meeting-service/internal/conf"
	"gpt-meeting-service/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, role *service.RoleTemplateService, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterRoleServer(srv, role)
	return srv
}
