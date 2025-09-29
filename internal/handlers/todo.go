package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/thedevsaddam/renderer"
	"github.com/akhilsharma/todo/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var rnd *renderer.Render

func init() {
	rnd = renderer.New()
}

type TodoHandler struct {
	DB             *mongo.Database
	CollectionName string
}

func (h *TodoHandler) Home(w http.ResponseWriter, r *http.Request) {
	err := rnd.Template(w, http.StatusOK, []string{"static/home.tpl"}, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TodoHandler) FetchTodos(w http.ResponseWriter, r *http.Request) {
	var todos []models.TodoModel

	cursor, err := h.DB.Collection(h.CollectionName).Find(context.Background(), bson.M{})
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{"message": "Failed to fetch todo", "error": err})
		return
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &todos); err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{"message": "Failed to fetch todo", "error": err})
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

	rnd.JSON(w, http.StatusOK, renderer.M{"data": todoList})
}

func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var t models.Todo

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusProcessing, err)
		return
	}

	if t.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "The title field is requried"})
		return
	}

	tm := models.TodoModel{
		ID:        primitive.NewObjectID(),
		Title:     t.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}

	_, err := h.DB.Collection(h.CollectionName).InsertOne(context.Background(), tm)
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{"message": "Failed to save todo", "error": err})
		return
	}

	rnd.JSON(w, http.StatusCreated, renderer.M{"message": "Todo created successfully", "todo_id": tm.ID.Hex()})
}

func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "The id is invalid"})
		return
	}

	var t models.Todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusProcessing, err)
		return
	}

	if t.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "The title field is requried"})
		return
	}

	_, err = h.DB.Collection(h.CollectionName).UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"title": t.Title, "completed": t.Completed}},
	)
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{"message": "Failed to update todo", "error": err})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{"message": "Todo updated successfully"})
}

func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "The id is invalid"})
		return
	}

	_, err = h.DB.Collection(h.CollectionName).DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{"message": "Failed to delete todo", "error": err})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{"message": "Todo deleted successfully"})
}

func (h *TodoHandler) Routes() http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/", h.FetchTodos)
		r.Post("/", h.CreateTodo)
		r.Put("/{id}", h.UpdateTodo)
		r.Delete("/{id}", h.DeleteTodo)
	})
	return rg
}