package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var db *sql.DB

func main() {
	fmt.Println("Serwer startuje na http://localhost:8000")

	// Połączenie z bazą danych
	var err error
	db, err = sql.Open("sqlite", "database.db")
	if err != nil {
		log.Fatal(err)
	}
	createTable()
	defer db.Close()

	// Obsługa stron
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/register", registerHandler)

	// Uruchomienie serwera
	log.Fatal(http.ListenAndServe(":8000", nil))
}

// Tworzenie tabeli w SQLite
func createTable() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		);
	`)
	if err != nil {
		log.Fatal(err)
	}
}

// Strona główna - formularz rejestracji
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

// Funkcja do hashowania hasła
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Rejestracja użytkownika
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Sprawdzenie czy użytkownik już istnieje
		var exists int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email=?", email).Scan(&exists)
		if err != nil {
			log.Fatal(err)
		}
		if exists > 0 {
			fmt.Fprintln(w, "Użytkownik już istnieje!")
			return
		}

		// Hashowanie hasła
		hashedPassword, err := hashPassword(password)
		if err != nil {
			log.Fatal(err)
		}

		// Dodanie użytkownika do bazy danych
		_, err = db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", email, hashedPassword)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprintln(w, "Rejestracja zakończona sukcesem!")
	}
}
