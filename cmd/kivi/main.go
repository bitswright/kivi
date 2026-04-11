package main

import (
	"fmt"
	"log"
	"net/http"

	kivihttp "github.com/bitswright/kivi/internal/http"
	"github.com/bitswright/kivi/internal/store"
)

func main() {
	// s := store.NewMemStore()
	// s, err := store.NewLogStore("kivi.log")
	s, err := store.NewWALStore("wal.log")
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	h := kivihttp.NewHandler(s)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	fmt.Println("Kivi listening on 5001")
	log.Fatal(http.ListenAndServe(":5001", mux))
}
