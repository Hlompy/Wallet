package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/Hlompy/Wallet/internal/config"
	"github.com/Hlompy/Wallet/internal/db"
	"github.com/Hlompy/Wallet/internal/handler"
	"github.com/Hlompy/Wallet/internal/repository"
	"github.com/Hlompy/Wallet/internal/service"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("config.env"); err != nil {
		log.Println("config.env not found, using system env")
	}

	cfg := config.Load()

	var database *sql.DB
	var err error

	for i := 1; i <= 10; i++ {
		database, err = db.New(cfg.DBDsn)
		if err == nil {
			log.Println("connected to database")
			break
		}

		log.Printf("db not ready (attempt %d/10), retrying...\n", i)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("could not connect to database:", err)
	}

	if err := db.Migrate(database); err != nil {
		log.Fatal(err)
	}

	repo := repository.New(database)
	svc := service.New(repo)
	h := handler.New(svc)

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/wallet", h.PostWallet).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/wallets/{id}", h.GetBalance).Methods(http.MethodGet)

	log.Println("server started on :" + cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, r))
}
