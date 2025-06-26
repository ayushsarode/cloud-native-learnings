package handler

import (
	"fmt"
	"net/http"
)

type Order struct {
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create an order")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "Order created successfully")
}

func (o *Order) ListAll(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List all orders")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "List of all orders")
}

func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get an order by ID")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Order details")
}

func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update an order by ID")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Order updated successfully")
}

func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete an order by ID")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Order deleted successfully")
}
