package main

import (
	"context"
	"fmt"
	"go-api/configs"
	"go-api/internal/auth"
	"go-api/pkg/dbgorm"
	"go-api/pkg/dbsqlc"
	"net/http"
)

func main() {
	conf := configs.LoadConfig()
	addr := conf.Http.Host + ":" + conf.Http.Port

	_, err := dbgorm.NewDb(conf)
	if err != nil {
		fmt.Print("connect to database: %w", err)
	}

	pool, err := dbsqlc.NewDbPool(
		context.Background(),
		conf.Db.Dbstring,
	)
	if err != nil {
		fmt.Print("connect to database: %w", err)
	}
	defer pool.Close()

	router := http.NewServeMux()
	auth.NewAuthHandler(router, auth.AuthHandlerDeps{Config: conf})

	server := http.Server{
		Addr:    addr,
		Handler: router,
	}

	fmt.Println("Server start:", addr)
	server.ListenAndServe()
}
