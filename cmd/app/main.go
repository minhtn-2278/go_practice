package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"employee-management/internal/handlers"
	"employee-management/internal/middleware"
	"employee-management/internal/repositories/postgres"
	"employee-management/internal/services"

	_ "github.com/lib/pq"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	databaseURL := os.Getenv("DATABASE_URL")

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingContext, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()
	if err := db.PingContext(pingContext); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	mux, err := newHTTPMux(db)
	if err != nil {
		return err
	}

	httpHandler := middleware.Logging(middleware.Recover(mux))

	server := &http.Server{
		Addr:              getEnv("HTTP_ADDR", ":8080"),
		Handler:           httpHandler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Server is listening at %s", getEnv("HTTP_ADDR", ":8080"))

	serverError := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverError <- err
		}
	}()

	select {
	case err := <-serverError:
		return fmt.Errorf("run HTTP server: %w", err)
	case <-ctx.Done():
	}

	log.Println("Shutting down HTTP server")
	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}
	log.Println("HTTP server stopped gracefully")

	return nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func newHTTPMux(db *sql.DB) (*http.ServeMux, error) {
	departmentRepository, err := postgres.NewDepartmentRepository(db)
	if err != nil {
		return nil, fmt.Errorf("create department repository: %w", err)
	}

	employeeRepository, err := postgres.NewEmployeeRepository(db, departmentRepository)
	if err != nil {
		return nil, fmt.Errorf("create employee repository: %w", err)
	}

	employeeService, err := services.NewEmployeeService(employeeRepository)
	if err != nil {
		return nil, fmt.Errorf("create employee service: %w", err)
	}

	employeeHandler, err := handlers.NewEmployeeHandler(employeeService)
	if err != nil {
		return nil, fmt.Errorf("create employee handler: %w", err)
	}

	departmentService, err := services.NewDepartmentService(departmentRepository)
	if err != nil {
		return nil, fmt.Errorf("create department service: %w", err)
	}

	departmentHandler, err := handlers.NewDepartmentHandler(departmentService)
	if err != nil {
		return nil, fmt.Errorf("create department handler: %w", err)
	}

	mux := http.NewServeMux()
	employeeHandler.Register(mux)
	departmentHandler.Register(mux)

	return mux, nil
}
