package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

const getQuery string = `SELECT * FROM todos;`
const createQuery string = `INSERT INTO todos (title, done) VALUES (?, ?);`
const dbFile string = "todos.db"

type TodoData struct {
	Title string
	Done bool
}

type Todo struct {
	Id int64 `json:"id"`
	Title string `json:"title"`
	Done bool `json:"done"`
}

var db *sql.DB


func readAllTodos() ([]Todo, error) {
	rows, err := db.Query(getQuery)
	if err != nil {
		return nil, errors.New("Failed to query database: " + err.Error())
	}
	defer rows.Close()

	var todos []Todo

	for rows.Next() {
		var todo Todo

		err := rows.Scan(&todo.Id, &todo.Title, &todo.Done)
		if err != nil {
			return nil, errors.New("Failed to query database: " + err.Error())
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

func createTodo(todoData TodoData) (*Todo, error) {
	stmt, err := db.Prepare(createQuery)
	if err != nil {
		return nil, errors.New("Failed to prepare query: " + err.Error())
	}
	defer stmt.Close()

	res, err := stmt.Exec(todoData.Title, todoData.Done)
	if err != nil {
		return nil, errors.New("Failed to insert row: " + err.Error())
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, errors.New("Failed to get last id: " + err.Error())
	}

	return 	&Todo{
		Id: id,
		Title: todoData.Title,
		Done: todoData.Done,
	}, nil
}

func handleGetTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := readAllTodos()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func handlePostTodos(w http.ResponseWriter, r *http.Request) {
	var todoData TodoData
	err := json.NewDecoder(r.Body).Decode(&todoData)
	if err != nil {
		log.Println("Failed to decode post body", err)
		w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte("Bad Request"))
		return
	}

	todo, err := createTodo(todoData)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func main() {
	log.Println("Opening connection to database...")
	var err error
	
	if db == nil {
		db, err = sql.Open("sqlite3", dbFile)
		if err != nil {
			log.Fatal("Failed to connect to database", err)
		}
	}
	defer db.Close()
	log.Println("Connected to database")


	router := http.NewServeMux()

	router.HandleFunc("GET /todos", handleGetTodos)
	router.HandleFunc("POST /todos", handlePostTodos)

	server := http.Server{
		Addr: ":8080",
		Handler: router,
	}

	log.Println("Starting server on port 8080")
	server.ListenAndServe()
}