package main

import (
	"context"
	"log"
	"google.golang.org/grpc"
	pb "InventoryManagement/proto"
)

// Вывод алертов при низком остатке
func streamAlerts(client pb.InventoryServiceClient) {
	stream, err := client.StreamStockAlerts(context.Background(), &pb.ListProductsRequest{})
	if err != nil {
		log.Fatalf("Ошибка открытия стрима: %v", err)
	}

	for {
		alert, err := stream.Recv()
		if err != nil {
			log.Fatalf("Ошибка стрима: %v", err)
		}
		log.Printf("Внимание пользователь! ID товара:%s — %s", alert.ProductId, alert.Message)
	}
}

func main() {
	// Подключение к серверу
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	defer conn.Close()
	client := pb.NewInventoryServiceClient(conn)

	// Товары
	products := []*pb.Product{
		{Id: "1", Name: "Тетрадь", Quantity: 160},
		{Id: "2", Name: "Ручка", Quantity: 120},
		{Id: "3", Name: "Ластик", Quantity: 70},
		{Id: "4", Name: "Маркер", Quantity: 60},
		{Id: "5", Name: "Линейка", Quantity: 12},
		{Id: "6", Name: "Папка", Quantity: 10},
	}

	// Добавление товаров
	for _, p := range products {
		resp, err := client.AddProduct(context.Background(), &pb.AddProductRequest{Product: p})
		if err != nil {
			log.Fatalf("Ошибка добавления продукта %s: %v", p.Name, err)
		}
		log.Printf("Добавлен: %s (ID: %s)", p.Name, resp.Id)
	}

	// Получение списка товаров
	listResp, err := client.ListProducts(context.Background(), &pb.ListProductsRequest{})
	if err != nil {
		log.Fatalf("Ошибка получения списка: %v", err)
	}
	log.Println("Список продуктов:")
	for _, p := range listResp.Products {
		log.Printf("%s — %d шт.", p.Name, p.Quantity)
	}

	// Получение одного товара
	getResp, err := client.GetProduct(context.Background(), &pb.GetProductRequest{Id: "2"})
	if err != nil {
		log.Fatalf("Ошибка получения продукта: %v", err)
	}
	log.Printf("Получен продукт: %s — %d шт.", getResp.Product.Name, getResp.Product.Quantity)

	// Изменение остатка
	_, err = client.UpdateStock(context.Background(), &pb.UpdateStockRequest{
		Change: &pb.StockChange{ProductId: "2", Delta: -3},
	})
	if err != nil {
		log.Fatalf("Ошибка обновления остатка: %v", err)
	}
	log.Println("Остаток по Ручке уменьшен на 3")

	// Удаление товара
	_, err = client.RemoveProduct(context.Background(), &pb.RemoveProductRequest{Id: "6"})
	if err != nil {
		log.Fatalf("Ошибка удаления: %v", err)
	}
	log.Println("Папка удалена из списка")

	// Финальный вывод списка товаров, чтобы увидеть изменения
	finalListResp, err := client.ListProducts(context.Background(), &pb.ListProductsRequest{})
	if err != nil {
		log.Fatalf("Ошибка получения списка товаров: %v", err)
	}
	log.Println("Финальный список товаров:")
	for _, p := range finalListResp.Products {
		log.Printf("%s — %d шт.", p.Name, p.Quantity)
	}

	// Оповещение о низком остатке товаров
	go streamAlerts(client)
	select {}
}