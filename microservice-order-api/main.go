package main

import (
	"fmt"
	"net/http"
)



func main() {
	server := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(basicHandler),
	}

	err := server.ListenAndServe()
	fmt.Println("Server is running at 8080")

	if err != nil {
		fmt.Println("failed to listen to server", err)
	}


	
}
	func basicHandler(w http.ResponseWriter, r *http.Request) {
		r.Method == http.MethodGet {
			if r.URL.Path == "/foo" {

			}
		}

		
	}



