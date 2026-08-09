package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/services/pkg/logger"
	pb "github.com/services/proto/product"
	"github.com/services/product-service/repository"
)

// ProductService implements the gRPC ProductServiceServer interface.
type ProductService struct {
	pb.UnimplementedProductServiceServer
	repo *repository.ProductRepository
	log  *logger.Logger
}

// NewProductService creates a new ProductService.
func NewProductService(repo *repository.ProductRepository, log *logger.Logger) *ProductService {
	return &ProductService{repo: repo, log: log}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "product name is required")
	}
	if req.Price < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "product price must be non-negative")
	}

	p := &repository.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageUrl,
		Category:    req.Category,
	}

	created, err := s.repo.Create(p)
	if err != nil {
		s.log.Error("Failed to create product: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create product")
	}

	return &pb.CreateProductResponse{
		Product: toProtoProduct(created),
	}, nil
}

func (s *ProductService) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "product ID is required")
	}

	p, err := s.repo.GetByID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found")
	}

	return &pb.GetProductResponse{
		Product: toProtoProduct(p),
	}, nil
}

func (s *ProductService) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)

	products, totalCount, err := s.repo.List(page, pageSize, req.Category)
	if err != nil {
		s.log.Error("Failed to list products: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list products")
	}

	var protoProducts []*pb.Product
	for _, p := range products {
		protoProducts = append(protoProducts, toProtoProduct(p))
	}

	return &pb.ListProductsResponse{
		Products:   protoProducts,
		TotalCount: int32(totalCount),
	}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "product ID is required")
	}

	p := &repository.Product{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageUrl,
		Category:    req.Category,
	}

	updated, err := s.repo.Update(p)
	if err != nil {
		s.log.Error("Failed to update product: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update product")
	}

	return &pb.UpdateProductResponse{
		Product: toProtoProduct(updated),
	}, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "product ID is required")
	}

	err := s.repo.Delete(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found")
	}

	return &pb.DeleteProductResponse{Success: true}, nil
}

func toProtoProduct(p *repository.Product) *pb.Product {
	return &pb.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		ImageUrl:    p.ImageURL,
		Category:    p.Category,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
