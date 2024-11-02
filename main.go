package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {
	s, err := newServer("docdb.data", "8081")

	if err != nil {
		log.Fatal(err)
	}
	defer s.db.Close()

	router := httprouter.New()

	router.POST("/docs", s.addDocument)
	router.GET("/docs", s.searchDocuments)
	router.GET("/docs/:id", s.getDocument)

	log.Println("Server started on port " + s.port)
	log.Fatal(http.ListenAndServe(":"+s.port, router))
}
