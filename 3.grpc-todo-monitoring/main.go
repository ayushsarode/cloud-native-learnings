package main

import (
	"context"

	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	
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
			Completed:   todo.Completed,
			CreatedAt:   todo.CreatedAt.Unix(),
			UpdatedAt:   todo.UpdatedAt.Unix(),
		},
	}, nil
}

func (s *TodoServer) GetTodo(ctx context.Context, req *pb.GetTodoRequest) (*pb.GetTodoResponse, error) {
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

	return &pb.GetTodoResponse{
		Todo: &pb.Todo{
			Id:          todo.ID,
			Title:       todo.Title,
			Description: todo.Description,
			Completed:   todo.Completed,
			CreatedAt:   todo.CreatedAt.Unix(),
			UpdatedAt:   todo.UpdatedAt.Unix(),
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
	todo.Completed = req.Completed
	todo.UpdatedAt = time.Now()

	s.updateMetrics()

	requestsTotal.WithLabelValues(method, "success").Inc()

	return &pb.UpdateTodoResponse{
		Todo: &pb.Todo{
			Id: todo.ID,
			Title: todo.Title,
			Description: todo.Description,
			Completed: todo.Completed,
			CreatedAt: todo.CreatedAt.Unix(),
			UpdatedAt: todo.UpdatedAt.Unix(),
		},
	}, nil
}

func (s *TodoServer) DeleteTodo (ctx context.Context, req *pb.DeleteTodoRequest) (*pb.DeleteTodoResponse, error) {
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

	return &pb.dele

}