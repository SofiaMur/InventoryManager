package main

import (
	"InventoryManagement/proto"
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

// Структура сервера с картой продукта
type server struct {
	proto.UnimplementedInventoryServiceServer
	products map[string]*proto.Product
}

// Добавление продукта
func (s *server) AddProduct(ctx context.Context, req *proto.AddProductRequest) (*proto.AddProductResponse, error) {
	product := req.GetProduct()
	s.products[product.Id] = product

	return &proto.AddProductResponse{Id: product.Id}, nil
}

// Получение продукта по ID
func (s *server) GetProduct(ctx context.Context, req *proto.GetProductRequest) (*proto.GetProductResponse, error) {
	product, exists := s.products[req.Id]
	if !exists {
		return nil, fmt.Errorf("Продукт не найден...")
	}

	return &proto.GetProductResponse{Product: product}, nil
}

// Изменение количества товара
func (s *server) UpdateStock(ctx context.Context, req *proto.UpdateStockRequest) (*proto.UpdateStockResponse, error) {
	product, exists := s.products[req.Change.ProductId]
	if !exists {
		return nil, fmt.Errorf("Продукт не найден...")
	}
	product.Quantity += req.Change.Delta

	return &proto.UpdateStockResponse{Product: product}, nil
}

// Удаление товара
func (s *server) RemoveProduct(ctx context.Context, req *proto.RemoveProductRequest) (*proto.RemoveProductResponse, error) {
	delete(s.products, req.Id)

	return &proto.RemoveProductResponse{Message: "Продукт успешно удален"}, nil
}

// Получение списка всех продуктов
func (s *server) ListProducts(ctx context.Context, req *proto.ListProductsRequest) (*proto.ListProductsResponse, error) {
	var productList []*proto.Product
	for _, p := range s.products {
		productList = append(productList, p)
	}

	return &proto.ListProductsResponse{Products: productList}, nil
}

// Поток оповещений о низком остатке
func (s *server) StreamStockAlerts(req *proto.ListProductsRequest, stream proto.InventoryService_StreamStockAlertsServer) error {
	for {
		for _, product := range s.products {
			if product.Quantity < 15 {
				alert := &proto.StockAlert{
					ProductId: product.Id,
					Message:   fmt.Sprintf("СРОЧНО! Пополните %s, осталось всего %d шт.", product.Name, product.Quantity),
				}

				if err := stream.Send(alert); err != nil {
					return err
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func main() {
	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Ошибка запуска: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterInventoryServiceServer(grpcServer, &server{
		products: make(map[string]*proto.Product),
	})

	fmt.Println("Сервер успешно запущен на порту :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
