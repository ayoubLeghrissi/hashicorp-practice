package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	inventorypb "github.com/services/proto/inventory"
	"github.com/services/inventory-service/repository"
	"github.com/services/pkg/logger"
	productpb "github.com/services/proto/product"
)

// InventoryService implements the gRPC InventoryServiceServer interface.
type InventoryService struct {
	inventorypb.UnimplementedInventoryServiceServer
	repo          *repository.InventoryRepository
	productClient productpb.ProductServiceClient
	log           *logger.Logger
}

// NewInventoryService creates a new InventoryService.
func NewInventoryService(repo *repository.InventoryRepository, productClient productpb.ProductServiceClient, log *logger.Logger) *InventoryService {
	return &InventoryService{repo: repo, productClient: productClient, log: log}
}

func (s *InventoryService) AddInventory(ctx context.Context, req *inventorypb.AddInventoryRequest) (*inventorypb.AddInventoryResponse, error) {
	if req.ProductId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "product_id is required")
	}
	if req.Quantity <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "quantity must be positive")
	}

	// Verify product exists via Product Service
	_, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{Id: req.ProductId})
	if err != nil {
		s.log.Error("Product not found in Product Service: %v", err)
		return nil, status.Errorf(codes.NotFound, "product not found: %s", req.ProductId)
	}

	item := &repository.InventoryItem{
		ProductID:         req.ProductId,
		Quantity:          req.Quantity,
		WarehouseLocation: req.WarehouseLocation,
		ReorderLevel:      req.ReorderLevel,
	}

	created, err := s.repo.Add(item)
	if err != nil {
		s.log.Error("Failed to add inventory: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to add inventory")
	}

	return &inventorypb.AddInventoryResponse{
		Item: toProtoInventoryItem(created, nil),
	}, nil
}

func (s *InventoryService) GetInventory(ctx context.Context, req *inventorypb.GetInventoryRequest) (*inventorypb.GetInventoryResponse, error) {
	if req.ProductId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "product_id is required")
	}

	item, err := s.repo.GetByProductID(req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "inventory not found for product")
	}

	// Enrich with product data
	var product *productpb.Product
	productResp, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{Id: req.ProductId})
	if err == nil {
		product = productResp.Product
	}

	return &inventorypb.GetInventoryResponse{
		Item: toProtoInventoryItem(item, product),
	}, nil
}

func (s *InventoryService) ListInventory(ctx context.Context, req *inventorypb.ListInventoryRequest) (*inventorypb.ListInventoryResponse, error) {
	items, totalCount, err := s.repo.List(int(req.Page), int(req.PageSize))
	if err != nil {
		s.log.Error("Failed to list inventory: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list inventory")
	}

	var protoItems []*inventorypb.InventoryItem
	for _, item := range items {
		var product *productpb.Product
		productResp, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{Id: item.ProductID})
		if err == nil {
			product = productResp.Product
		}
		protoItems = append(protoItems, toProtoInventoryItem(item, product))
	}

	return &inventorypb.ListInventoryResponse{
		Items:      protoItems,
		TotalCount: int32(totalCount),
	}, nil
}

func (s *InventoryService) UpdateQuantity(ctx context.Context, req *inventorypb.UpdateQuantityRequest) (*inventorypb.UpdateQuantityResponse, error) {
	if req.ProductId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "product_id is required")
	}

	item, err := s.repo.UpdateQuantity(req.ProductId, req.QuantityChange)
	if err != nil {
		s.log.Error("Failed to update quantity: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update quantity: %v", err)
	}

	return &inventorypb.UpdateQuantityResponse{
		Item:    toProtoInventoryItem(item, nil),
		Success: true,
	}, nil
}

func (s *InventoryService) DeductInventory(ctx context.Context, req *inventorypb.DeductInventoryRequest) (*inventorypb.DeductInventoryResponse, error) {
	if req.ProductId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "product_id is required")
	}
	if req.Quantity <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "quantity must be positive")
	}

	remaining, err := s.repo.Deduct(req.ProductId, req.Quantity)
	if err != nil {
		s.log.Error("Failed to deduct inventory: %v", err)
		return nil, status.Errorf(codes.FailedPrecondition, "insufficient stock")
	}

	return &inventorypb.DeductInventoryResponse{
		Success:           true,
		RemainingQuantity: remaining,
	}, nil
}

func (s *InventoryService) CheckStock(ctx context.Context, req *inventorypb.CheckStockRequest) (*inventorypb.CheckStockResponse, error) {
	if req.ProductId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "product_id is required")
	}

	inStock, available, err := s.repo.CheckStock(req.ProductId, req.RequestedQuantity)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found in inventory")
	}

	return &inventorypb.CheckStockResponse{
		InStock:           inStock,
		AvailableQuantity: available,
	}, nil
}

func toProtoInventoryItem(item *repository.InventoryItem, product *productpb.Product) *inventorypb.InventoryItem {
	protoItem := &inventorypb.InventoryItem{
		Id:                item.ID,
		ProductId:         item.ProductID,
		Quantity:          item.Quantity,
		WarehouseLocation: item.WarehouseLocation,
		ReorderLevel:      item.ReorderLevel,
		UpdatedAt:         item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if product != nil {
		protoItem.Product = product
	}
	return protoItem
}
