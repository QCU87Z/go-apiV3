// Shit half complete tutorial
// https://www.linkedin.com/pulse/building-your-first-crud-app-go-hands-on-tutorial-zackaria-slimane-
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var dbHost string
var dbUser string
var dbPwd string
var dbName string

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	dbHost = os.Getenv("dbHost")
	dbUser = os.Getenv("dbUser")
	dbPwd = os.Getenv("dbPwd")
	dbName = os.Getenv("dbName")

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/users", handleGetAll)
	router.Post("/user", handleCreate)
	router.Get("/user/{id}", handleGet)
	router.Put("/user/{id}", handleUpdate)
	router.Delete("/user/{id}", handleDelete)

	log.Fatal(http.ListenAndServe(":8000", router))
}

// create a new user struct
type User struct {
	ID    int
	Name  string
	Email string
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	// Establish a database connection
	db, err := sql.Open(dbHost, dbUser+":"+dbPwd+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Parse JSON data from the request body
	var user User
	json.NewDecoder(r.Body).Decode(&user)

	// Invoke the CreateUser function to execute the database operation
	err = CreateUser(db, user.Name, user.Email)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Respond with a success status and message
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "User created successfully")
}

func CreateUser(db *sql.DB, name, email string) error {
	query := "INSERT INTO users (name, email) VALUES (?, ?)"
	_, err := db.Exec(query, name, email)
	if err != nil {
		return err
	}
	return nil
}

func handleGetAll(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open(dbHost, dbUser+":"+dbPwd+"@/"+dbName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	users := GetUsers(db)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func GetUsers(db *sql.DB) []User {
	query := "SELECT * FROM users"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}

	var users []User

	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	return users
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open(dbHost, dbUser+":"+dbPwd+"@/"+dbName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	//retrieve the id parameter
	idParam := chi.URLParam(r, "id")

	userID, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := GetUser(db, userID)
	if err != nil {
		log.Print("user not found with id", userID)
		http.Error(w, "User not found ", http.StatusNotFound)
		return
	}
	// Convert the user object to JSON and send it
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func GetUser(db *sql.DB, id int) (*User, error) {
	query := "SELECT * FROM users WHERE id = ?"
	row := db.QueryRow(query, id)

	user := &User{}
	err := row.Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	// Establish a database connection
	db, err := sql.Open(dbHost, dbUser+":"+dbPwd+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Get the 'id' parameter from the URL
	idParam := chi.URLParam(r, "id")

	// Convert 'id' to an integer
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user User
	err = json.NewDecoder(r.Body).Decode(&user)

	// Invoke the UpdateUser function to update the user data in the database
	UpdateUser(db, userID, user.Name, user.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	fmt.Fprintln(w, "User updated successfully")
}

func UpdateUser(db *sql.DB, id int, name, email string) error {
	query := "UPDATE users SET name = ?, email = ? WHERE id = ?"
	_, err := db.Exec(query, name, email, id)
	if err != nil {
		return err
	}
	return nil
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	// Establish a database connection
	db, err := sql.Open(dbHost, dbUser+":"+dbPwd+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Get the 'id' parameter from the URL
	idParam := chi.URLParam(r, "id")

	// Convert 'id' to an integer
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	user := DeleteUser(db, userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	fmt.Fprintln(w, "User deleted successfully")
	// Convert the user object to JSON and send it
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func DeleteUser(db *sql.DB, id int) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}
