package server

import (
	"log"
	"net/http"

	"github.com/pirateunclejack/go-practice/golang-csrf-project/server/middleware"
)

func StartServer(hostname string, port string) error {
    host := hostname + ":" + port
    log.Println("listening on: ", host)

    handler := middleware.NewHandler()

    http.Handle("/", handler)
    return http.ListenAndServe(host, nil)
}
