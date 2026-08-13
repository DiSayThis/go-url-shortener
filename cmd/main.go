package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-api/configs"
	"go-api/internal/auth"
	"go-api/internal/database"
	"go-api/internal/link"
	"go-api/pkg/db"
	"go-api/pkg/logging"
	"go-api/pkg/middleware"
)

func main() {
	conf := configs.LoadConfig()
	appLogger := logging.NewLogger(conf.Environment)
	slog.SetDefault(appLogger)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := run(ctx, conf); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, conf *configs.Config) error {

	pool, err := db.NewDbPool(ctx, conf.Db.DbUrl)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()
	queries := database.New(pool)

	router := http.NewServeMux()

	//Repositories
	linkRepository := link.NewLinkRepository(queries)

	//Services
	linkService := link.NewLinkService(linkRepository)

	//Handle routes
	auth.NewAuthHandler(router, auth.AuthHandlerDeps{Config: conf})
	link.NewLinkHandler(router, link.LinkHandlerDeps{
		Service: linkService,
		Logger:  slog.Default(),
	})

	//Middlewares
	corsMiddleware := middleware.CORS(conf.CORS.AllowedOrigins)
	stack := middleware.Chain(
		middleware.RequestLogger,
		corsMiddleware,
	)(router)

	server := &http.Server{
		Addr:              net.JoinHostPort(conf.Http.Host, conf.Http.Port),
		Handler:           stack,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverError := make(chan error, 1)

	go func() {
		slog.Info("starting HTTP server", "address", server.Addr)
		serverError <- server.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("run HTTP server: %w", err)
		}
		return nil

	case <-ctx.Done():
		slog.Info("shutting down application")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	err = <-serverError
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("stop HTTP server: %w", err)
	}

	return nil
}
