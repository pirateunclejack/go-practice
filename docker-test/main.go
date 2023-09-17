package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
)

func helloFunc(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func hiFunc(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "Hi")
}

func main() {
	http.HandleFunc("/", helloFunc)
	http.HandleFunc("/hi", hiFunc)

	log.Fatal(http.ListenAndServe(":8081", nil))
}
