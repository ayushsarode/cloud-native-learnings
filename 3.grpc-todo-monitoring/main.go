package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "todo-grpc/pb" // Replace with your module path
)

// Prometheus metrics
var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	todoCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "todos_total",
			Help: "Total number of todos",
		},
	)

	completedTodoCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "todos_completed_total",
			Help: "Total number of completed todos",
		},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(todoCount)
	prometheus.MustRegister(completedTodoCount)
}

// Helper function to convert bool to string for protobuf
func convertBoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// Todo represents a todo item
type Todo struct {
	ID          string
	Title       string
	Description string
	Completed   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TodoServer implements the TodoService
type TodoServer struct {
	pb.UnimplementedTodoServiceServer
	todos map[string]*Todo
	mu    sync.RWMutex
}

func NewTodoServer() *TodoServer {
	return &TodoServer{
		todos: make(map[string]*Todo),
	}
}

func (s *TodoServer) updateMetrics() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.todos)
	completed := 0

	for _, todo := range s.todos {
		if todo.Completed {
			completed++
		}
	}

	todoCount.Set(float64(total))
	completedTodoCount.Set(float64(completed))
}

func (s *TodoServer) CreateTodo(ctx context.Context, req *pb.CreateTodoRequest) (*pb.CreateTodoResponse, error) {
	start := time.Now()
	method := "CreateTodo"

	defer func() {
		requestDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
	}()

	if req.Title == "" {
		requestsTotal.WithLabelValues(method, "error").Inc()
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	todo := &Todo{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.todos[todo.ID] = todo
	s.updateMetrics()

	requestsTotal.WithLabelValues(method, "success").Inc()

	return &pb.CreateTodoResponse{
		Todo: &pb.Todo{
			Id:          todo.ID,
			Title:       todo.Title,
			Description: todo.Description,
			Completed:   convertBoolToString(todo.Completed),
			CreatedAt:   todo.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   todo.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (s *TodoServer) GetTodo(ctx context.Context, req *pb.GetTodoRequest) (*pb.CreateTodoResponse, error) {
	start := time.Now()
	method := "GetTodo"

	defer func() {
		requestDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
	}()

	s.mu.RLock()
	defer s.mu.RUnlock()

	todo, exists := s.todos[req.Id]
	if !exists {
		requestsTotal.WithLabelValues(method, "not_found").Inc()
		return nil, status.Error(codes.NotFound, "todo not found")
	}

	requestsTotal.WithLabelValues(method, "success").Inc()

	return &pb.CreateTodoResponse{ // Changed from GetTodoResponse to CreateTodoResponse as per proto definition
		Todo: &pb.Todo{
			Id:          todo.ID,
			Title:       todo.Title,
			Description: todo.Description,
			Completed:   convertBoolToString(todo.Completed),
			CreatedAt:   todo.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   todo.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (s *TodoServer) UpdateTodo(ctx context.Context, req *pb.UpdateTodoRequest) (*pb.UpdateTodoResponse, error) {
	start := time.Now()
	method := "UpdateTodo"

	defer func() {
		requestDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
	}()

	s.mu.Lock()

	defer s.mu.Unlock()

	todo, exists := s.todos[req.Id]

	if !exists {
		requestsTotal.WithLabelValues(method, "not_found").Inc()

		return nil, status.Error(codes.NotFound, "todo not found")
	}

	todo.Title = req.Title
	todo.Description = req.Description
	todo.Completed = req.Completed == "true" // Convert string to bool
	todo.UpdatedAt = time.Now()

	s.updateMetrics()

	requestsTotal.WithLabelValues(method, "success").Inc()

	return &pb.UpdateTodoResponse{
		Todo: &pb.Todo{
			Id:          todo.ID,
			Title:       todo.Title,
			Description: todo.Description,
			Completed:   convertBoolToString(todo.Completed),
			CreatedAt:   todo.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   todo.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (s *TodoServer) DeleteTodo(ctx context.Context, req *pb.DeleteTodoRequest) (*pb.DeleteTodoResponse, error) {
	start := time.Now()
	method := "DeleteTodo"

	defer func() {
		requestDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
	}()

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.todos[req.Id]

	if !exists {
		requestsTotal.WithLabelValues(method, "not_found").Inc()

		return nil, status.Error(codes.NotFound, "todo not found")
	}

	delete(s.todos, req.Id)
	s.updateMetrics()

	requestsTotal.WithLabelValues(method, "success").Inc()

	return &pb.DeleteTodoResponse{Success: true}, nil

}

func (s *TodoServer) ListTodos(ctx context.Context, req *pb.ListTodosRequest) (*pb.ListTodosResponse, error) {
	start := time.Now()
	method := "ListTodos"

	defer func() {
		requestDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
	}()

	s.mu.RLock()
	defer s.mu.RUnlock()

	var todos []*pb.Todo

	for _, todo := range s.todos {
		todos = append(todos, &pb.Todo{
			Id:          todo.ID,
			Title:       todo.Title,
			Description: todo.Description,
			Completed:   convertBoolToString(todo.Completed),
			CreatedAt:   todo.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   todo.UpdatedAt.Format(time.RFC3339),
		})
	}

	requestsTotal.WithLabelValues(method, "success").Inc()

	return &pb.ListTodosResponse{
		Todo:  todos, // Changed from Todos to Todo as per proto definition
		Total: int32(len(todos)),
	}, nil
}

func main() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Metrics server listening on :8080/metrics")

		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)

	}

	s := grpc.NewServer()

	todoServer := NewTodoServer()

	pb.RegisterTodoServiceServer(s, todoServer)

	log.Printf("gRPC server listening on :50051")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
