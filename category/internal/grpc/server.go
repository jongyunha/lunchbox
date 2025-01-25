package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/jongyunha/lunchbox/category/categorypb"
	"github.com/jongyunha/lunchbox/category/internal/application"
	"google.golang.org/grpc"
)

type server struct {
	app application.App
	categorypb.UnimplementedCategoryServiceServer
}

var _ categorypb.CategoryServiceServer = (*server)(nil)

func RegisterServer(app application.App, registrar grpc.ServiceRegistrar) error {
	categorypb.RegisterCategoryServiceServer(registrar, server{app: app})
	return nil
}

func (s server) RegisterCategory(ctx context.Context, request *categorypb.RegisterCategoryRequest,
) (*categorypb.RegisterCategoryResponse, error) {
	id := uuid.New().String()
	err := s.app.RegisterCategory(ctx, application.RegisterCategory{
		ID:   id,
		Name: request.GetName(),
	})
	return &categorypb.RegisterCategoryResponse{Id: id}, err
}
