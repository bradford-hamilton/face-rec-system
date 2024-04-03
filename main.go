package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.mlbam.net/blamson/face-rec-system/internal/server"
	"github.mlbam.net/blamson/face-rec-system/internal/storage"

	// postgres driver
	_ "github.com/lib/pq"
)

func main() {
	db, err := storage.NewDB()
	if err != nil {
		log.Fatal("failed to initialize database")
	}

	if err := db.ExportEmbeddings(); err != nil {
		log.Fatalf("failed export embeddings to npy: %+v", err)
	}

	s := server.New(db)
	port := os.Getenv("FACE_REC_SYSTEM_SERVER_PORT")
	if port == "" {
		port = "4000"
	}

	fmt.Printf("Serving application on port %s\n", port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, s.Mux))
}
