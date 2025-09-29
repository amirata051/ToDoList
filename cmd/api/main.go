package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/amirata051/todo-list/internal/handlers"
	"github.com/amirata051/todo-list/internal/store"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
)

const (
	port string = ":9000"
)

func main() {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, os.Kill)

	db := store.Connect()
	defer func() {
		if err := db.Client().Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v\n", err)
		}
	}()

	rnd := renderer.New()

	// Create the handler instance with its dependencies
	todoHandler := &handlers.TodoHandler{
		DB:  db,
		Rnd: rnd,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", todoHandler.Home)

	r.Mount("/todo", todoRoutes(todoHandler))

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Listening on port ", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel()
	log.Println("Server gracefully stopped!")
}

// todoRoutes sets up the routes for the todo resource.
func todoRoutes(h *handlers.TodoHandler) http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/", h.FetchTodos)
		r.Post("/", h.CreateTodo)
		r.Put("/{id}", h.UpdateTodo)
		r.Delete("/{id}", h.DeleteTodo)
	})
	return rg
}
