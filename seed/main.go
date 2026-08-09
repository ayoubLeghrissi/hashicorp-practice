package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	inventorypb "github.com/services/proto/inventory"
	productpb "github.com/services/proto/product"
)

func main() {
	// Connect to Product Service
	productConn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to product service: %v", err)
	}
	defer productConn.Close()
	productClient := productpb.NewProductServiceClient(productConn)

	// Connect to Inventory Service
	inventoryConn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to inventory service: %v", err)
	}
	defer inventoryConn.Close()
	inventoryClient := inventorypb.NewInventoryServiceClient(inventoryConn)

	products := []struct {
		Name        string
		Description string
		Price       float64
		Category    string
		Qty         int32
	}{
		{"Neural Interface V2", "Brain-computer interface for seamless device control.", 1299.99, "Cybernetics", 15},
		{"Quantum Laptop Pro", "Compute at the speed of light. 1TB Qubit storage.", 4999.00, "Computers", 5},
		{"Holographic Display", "True 3D holographic monitor for designers.", 899.50, "Displays", 30},
		{"Gravity Boots", "Anti-gravity footwear for extreme comfort.", 399.99, "Apparel", 50},
		{"Fusion Core Battery", "Never charge your phone again. Lasts 100 years.", 149.00, "Accessories", 200},
	}

	for _, p := range products {
		pres, err := productClient.CreateProduct(context.Background(), &productpb.CreateProductRequest{
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Category:    p.Category,
		})
		if err != nil {
			log.Printf("Failed to create product %s: %v", p.Name, err)
			continue
		}

		product := pres.Product
		fmt.Printf("Created Product: %s (ID: %s)\n", product.Name, product.Id)

		_, err = inventoryClient.AddInventory(context.Background(), &inventorypb.AddInventoryRequest{
			ProductId:         product.Id,
			Quantity:          p.Qty,
			WarehouseLocation: "Main Hub",
			ReorderLevel:      5,
		})
		if err != nil {
			log.Printf("Failed to add inventory for %s: %v", product.Name, err)
		} else {
			fmt.Printf("Added %d units to inventory for %s\n", p.Qty, product.Name)
		}
	}
	fmt.Println("Seeding complete!")
}
