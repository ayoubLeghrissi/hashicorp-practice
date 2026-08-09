package service

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cartpb "github.com/services/proto/cart"
	"github.com/services/cart-service/repository"
	inventorypb "github.com/services/proto/inventory"
	"github.com/services/pkg/logger"
	productpb "github.com/services/proto/product"
)

// CartService implements the gRPC CartServiceServer interface.
type CartService struct {
	cartpb.UnimplementedCartServiceServer
	repo            *repository.CartRepository
	productClient   productpb.ProductServiceClient
	inventoryClient inventorypb.InventoryServiceClient
	log             *logger.Logger
}

// NewCartService creates a new CartService.
func NewCartService(
	repo *repository.CartRepository,
	productClient productpb.ProductServiceClient,
	inventoryClient inventorypb.InventoryServiceClient,
	log *logger.Logger,
) *CartService {
	return &CartService{
		repo:            repo,
		productClient:   productClient,
		inventoryClient: inventoryClient,
		log:             log,
	}
}

func (s *CartService) GetCart(ctx context.Context, req *cartpb.GetCartRequest) (*cartpb.GetCartResponse, error) {
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	cart, err := s.repo.GetOrCreateActiveCart(req.UserId)
	if err != nil {
		s.log.Error("Failed to get cart: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get cart")
	}

	protoCart, err := s.buildProtoCart(ctx, cart)
	if err != nil {
		return nil, err
	}

	return &cartpb.GetCartResponse{Cart: protoCart}, nil
}

func (s *CartService) AddToCart(ctx context.Context, req *cartpb.AddToCartRequest) (*cartpb.AddToCartResponse, error) {
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}
	if req.ProductId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "product_id is required")
	}
	if req.Quantity <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "quantity must be positive")
	}

	// 1. Verify product exists and get price
	productResp, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{Id: req.ProductId})
	if err != nil {
		s.log.Error("Product not found: %v", err)
		return nil, status.Errorf(codes.NotFound, "product not found")
	}

	// 2. Check stock availability
	stockResp, err := s.inventoryClient.CheckStock(ctx, &inventorypb.CheckStockRequest{
		ProductId:         req.ProductId,
		RequestedQuantity: req.Quantity,
	})
	if err != nil {
		s.log.Error("Failed to check stock: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to check stock availability")
	}
	if !stockResp.InStock {
		return &cartpb.AddToCartResponse{
			Success: false,
			Message: fmt.Sprintf("Insufficient stock. Available: %d", stockResp.AvailableQuantity),
		}, nil
	}

	// 3. Get or create cart
	cart, err := s.repo.GetOrCreateActiveCart(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cart")
	}

	// 4. Add item to cart
	_, err = s.repo.AddItem(cart.ID, req.ProductId, req.Quantity, productResp.Product.Price)
	if err != nil {
		s.log.Error("Failed to add item to cart: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to add item to cart")
	}

	// 5. Return updated cart
	protoCart, err := s.buildProtoCart(ctx, cart)
	if err != nil {
		return nil, err
	}

	return &cartpb.AddToCartResponse{
		Cart:    protoCart,
		Success: true,
		Message: "Item added to cart",
	}, nil
}

func (s *CartService) RemoveFromCart(ctx context.Context, req *cartpb.RemoveFromCartRequest) (*cartpb.RemoveFromCartResponse, error) {
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	cart, err := s.repo.GetOrCreateActiveCart(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cart")
	}

	err = s.repo.RemoveItem(cart.ID, req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "item not found in cart")
	}

	protoCart, err := s.buildProtoCart(ctx, cart)
	if err != nil {
		return nil, err
	}

	return &cartpb.RemoveFromCartResponse{
		Cart:    protoCart,
		Success: true,
	}, nil
}

func (s *CartService) UpdateCartItem(ctx context.Context, req *cartpb.UpdateCartItemRequest) (*cartpb.UpdateCartItemResponse, error) {
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	// Check stock for new quantity
	if req.Quantity > 0 {
		stockResp, err := s.inventoryClient.CheckStock(ctx, &inventorypb.CheckStockRequest{
			ProductId:         req.ProductId,
			RequestedQuantity: req.Quantity,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to check stock")
		}
		if !stockResp.InStock {
			return &cartpb.UpdateCartItemResponse{
				Success: false,
				Message: fmt.Sprintf("Insufficient stock. Available: %d", stockResp.AvailableQuantity),
			}, nil
		}
	}

	cart, err := s.repo.GetOrCreateActiveCart(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cart")
	}

	_, err = s.repo.UpdateItemQuantity(cart.ID, req.ProductId, req.Quantity)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "item not found in cart")
	}

	protoCart, err := s.buildProtoCart(ctx, cart)
	if err != nil {
		return nil, err
	}

	return &cartpb.UpdateCartItemResponse{
		Cart:    protoCart,
		Success: true,
		Message: "Cart item updated",
	}, nil
}

func (s *CartService) Checkout(ctx context.Context, req *cartpb.CheckoutRequest) (*cartpb.CheckoutResponse, error) {
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	cart, err := s.repo.GetOrCreateActiveCart(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cart")
	}

	items, err := s.repo.GetCartItems(cart.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cart items")
	}

	if len(items) == 0 {
		return nil, status.Errorf(codes.FailedPrecondition, "cart is empty")
	}

	// Deduct inventory for each item
	for _, item := range items {
		_, err := s.inventoryClient.DeductInventory(ctx, &inventorypb.DeductInventoryRequest{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
		if err != nil {
			s.log.Error("Failed to deduct inventory for product %s: %v", item.ProductID, err)
			return nil, status.Errorf(codes.FailedPrecondition,
				"insufficient stock for product %s", item.ProductID)
		}
	}

	// Calculate total
	total, err := s.repo.CalculateTotal(cart.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate total")
	}

	// Mark cart as checked out
	orderID, err := s.repo.CheckoutCart(cart.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to checkout cart")
	}

	s.log.Info("Checkout completed for user %s: order=%s, total=%.2f", req.UserId, orderID, total)

	return &cartpb.CheckoutResponse{
		Success:    true,
		OrderId:    orderID,
		Message:    "Checkout successful",
		TotalPrice: total,
	}, nil
}

func (s *CartService) ClearCart(ctx context.Context, req *cartpb.ClearCartRequest) (*cartpb.ClearCartResponse, error) {
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	cart, err := s.repo.GetOrCreateActiveCart(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cart")
	}

	err = s.repo.ClearCart(cart.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to clear cart")
	}

	return &cartpb.ClearCartResponse{Success: true}, nil
}

// buildProtoCart builds a Cart proto message with enriched product data.
func (s *CartService) buildProtoCart(ctx context.Context, cart *repository.Cart) (*cartpb.Cart, error) {
	items, err := s.repo.GetCartItems(cart.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cart items")
	}

	var protoItems []*cartpb.CartItem
	var totalPrice float64

	for _, item := range items {
		protoItem := &cartpb.CartItem{
			Id:        item.ID,
			CartId:    item.CartID,
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}

		// Enrich with product data
		productResp, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{Id: item.ProductID})
		if err == nil {
			protoItem.Product = productResp.Product
		}

		protoItems = append(protoItems, protoItem)
		totalPrice += float64(item.Quantity) * item.UnitPrice
	}

	return &cartpb.Cart{
		Id:         cart.ID,
		UserId:     cart.UserID,
		Items:      protoItems,
		TotalPrice: totalPrice,
		Status:     cart.Status,
		CreatedAt:  cart.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  cart.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
