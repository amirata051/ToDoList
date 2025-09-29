package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/amirata051/todo-list/internal/models"
	"github.com/go-chi/chi"
	"github.com/thedevsaddam/renderer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	collectionName string = "todo"
)

// TodoHandler holds the dependencies for the todo handlers.
type TodoHandler struct {
	DB  *mongo.Database
	Rnd *renderer.Render
}

// Home is the handler for the home page.
func (h *TodoHandler) Home(w http.ResponseWriter, r *http.Request) {
	err := h.Rnd.Template(w, http.StatusOK, []string{"static/home.tpl"}, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// FetchTodos fetches all todos.
func (h *TodoHandler) FetchTodos(w http.ResponseWriter, r *http.Request) {
	var todos []models.TodoModel

	cursor, err := h.DB.Collection(collectionName).Find(r.Context(), bson.M{})
	if err != nil {
		h.Rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to fetch todo",
			"error":   err.Error(),
		})
		return
	}
	defer cursor.Close(r.Context())

	if err = cursor.All(r.Context(), &todos); err != nil {
		h.Rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to fetch todo",
			"error":   err.Error(),
		})
		return
	}

	todoList := []models.Todo{}
	for _, t := range todos {
		todoList = append(todoList, models.Todo{
			ID:        t.ID.Hex(),
			Title:     t.Title,
			Completed: t.Completed,
			CreatedAt: t.CreatedAt,
		})
	}

	h.Rnd.JSON(w, http.StatusOK, renderer.M{
		"data": todoList,
	})
}

// CreateTodo creates a new todo.
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var t models.Todo

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.Rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Invalid request"})
		return
	}

	if t.Title == "" {
		h.Rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title field is requried",
		})
		return
	}

	tm := models.TodoModel{
		ID:        primitive.NewObjectID(),
		Title:     t.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}
	_, err := h.DB.Collection(collectionName).InsertOne(r.Context(), tm)
	if err != nil {
		h.Rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to save todo",
			"error":   err.Error(),
		})
		return
	}

	h.Rnd.JSON(w, http.StatusCreated, renderer.M{
		"message": "Todo created successfully",
		"todo_id": tm.ID.Hex(),
	})
}

// UpdateTodo updates an existing todo.
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		h.Rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
		})
		return
	}

	var t models.Todo

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.Rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Invalid request"})
		return
	}

	if t.Title == "" {
		h.Rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title field is requried",
		})
		return
	}

	_, err = h.DB.Collection(collectionName).UpdateOne(
		r.Context(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"title": t.Title, "completed": t.Completed}},
	)
	if err != nil {
		h.Rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to update todo",
			"error":   err.Error(),
		})
		return
	}

	h.Rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo updated successfully",
	})
}

// DeleteTodo deletes a todo.
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		h.Rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
		})
		return
	}

	_, err = h.DB.Collection(collectionName).DeleteOne(r.Context(), bson.M{"_id": objID})
	if err != nil {
		h.Rnd.JSON(w, http.StatusInternalServerError, renderer.M{
			"message": "Failed to delete todo",
			"error":   err.Error(),
		})
		return
	}

	h.Rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo deleted successfully",
	})
}
