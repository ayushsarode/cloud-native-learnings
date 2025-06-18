package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "todo-grpc/pb"
)

func main() {
	// Use grpc.Dial instead of grpc.NewClient (which doesn't exist)
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	defer conn.Close()

	client := pb.NewTodoServiceClient(conn)

	// Increase timeout to 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	//create a todo

	createResp, err := client.CreateTodo(ctx, &pb.CreateTodoRequest{
		Title:       "Learn gRPC",
		Description: "Build a todo app with gRPC and monitoring",
	})

	if err != nil {
		log.Fatalf("CreateTodo failed: %v", err)
	}

	log.Printf("Created Todo: %v", createResp.Todo)

	//get todo

	getResp, err := client.GetTodo(ctx, &pb.GetTodoRequest{
		Id: createResp.Todo.Id,
	})

	if err != nil {
		log.Fatalf("GetTodo failed: %v", err)
	}

	log.Printf("Retrieved todo: %v", getResp.Todo)

	// update the todo

	updateResp, err := client.UpdateTodo(ctx, &pb.UpdateTodoRequest{
		Id:          createResp.Todo.Id,
		Title:       "Learn gRPC - Updated",
		Description: "Build a todo app with gRPC and monitoring - Updated",
		Completed:   "true",
	})

	if err != nil {
		log.Fatalf("UpdateTodo failed: %v", err)
	}
	log.Printf("Updated Todo: %v", updateResp.Todo)

	listResp, err := client.ListTodos(ctx, &pb.ListTodosRequest{})
	if err != nil {
		log.Fatalf("ListTodos failed: %v", err)
	}

	log.Printf("Listed %d Todos", listResp.Total)

	for _, todo := range listResp.Todo {
		log.Printf("Todo: %v", todo)
	}

}
