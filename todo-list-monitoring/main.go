package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	// "github.com/prometheus/client_golang/prometheus"
	// "github.com/prometheus/client_golang/prometheus/promhttp"
)



type TODO struct {
	ID int `json:"id"`
	Task string	`json:"task"`
	Done string	`json:"done"`

}

var (
	todos	= []TODO{}
	mutex   = sync.Mutex{}
	todoID	= 1

	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "todo_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"endpoint"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "todo_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)


	todoCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "todo_count",
			Help: "Current number of todos",
		},
	)

)

func init() {
	prometheus.MustRegister(requestsTotal,requestDuration,todoCount)
}

func observe(endpoint string, handler http.HandlerFunc) http.HandlerFunc {
	return func( w http.ResponseWriter, r *http.Request) {
		timer := prometheus.NewTimer(requestDuration.WithLabelValues(endpoint))

		defer timer.ObserveDuration()

		requestsTotal.WithLabelValues(endpoint).Inc()


		handler(w,r)                                                                                                                
	}
}

func getTodos (w http.ResponseWriter, r * http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	json.NewEncoder(w).Encode(todos)
}

func addTodos (w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	var todo TODO

	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	todo.ID = todoID
	todoID ++

	todos = append(todos, todo)
	todoCount.Set(float64(len(todos)))
	json.NewEncoder(w).Encode(todo)




}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/todos", observe("get_todos", getTodos))
	http.HandleFunc("/add", observe("add_todo", addTodos))

	fmt.Println("Server running on :8000")

	http.ListenAndServe(":8000", nil)
}


