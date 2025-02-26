package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/sessions"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var db *sql.DB
var store = sessions.NewCookieStore([]byte("super-secret-key"))

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
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/account", accountHandler)
	http.HandleFunc("/logout", logoutHandler)

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

// Main page - registration form
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

// Password hashing
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// User registration
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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		email := r.FormValue("email")
		password := r.FormValue("password")

		log.Println("Próba logowania dla e-maila:", email)

		// Pobranie użytkownika z bazy
		var id int
		var hashedPassword string
		err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&id, &hashedPassword)
		if err != nil {
			log.Println("Błąd: Użytkownik nie znaleziony:", email)
			fmt.Fprintln(w, "Nieprawidłowy e-mail lub hasło")
			return
		}

		// Sprawdzenie hasła
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		if err != nil {
			log.Println("Błąd: Niepoprawne hasło dla e-maila:", email)
			fmt.Fprintln(w, "Nieprawidłowy e-mail lub hasło")
			return
		}

		// Tworzenie sesji
		session, _ := store.Get(r, "session")
		session.Values["userID"] = id
		session.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   3600, // Sesja ważna 1 godzinę
			HttpOnly: true, // Nie pozwala na dostęp JavaScript
		}
		err = session.Save(r, w)
		if err != nil {
			log.Println("Błąd zapisywania sesji:", err)
			fmt.Fprintln(w, "Błąd serwera")
			return
		}

		log.Println("Zalogowano! userID:", id)
		fmt.Fprintln(w, "Zalogowano pomyślnie!")
	}
}

func accountHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")

	// Sprawdzenie, czy użytkownik jest zalogowany
	userID, ok := session.Values["userID"]
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Pobranie e-maila użytkownika z bazy
	var email string
	err := db.QueryRow("SELECT email FROM users WHERE id = ?", userID).Scan(&email)
	if err != nil {
		http.Error(w, "Błąd serwera", http.StatusInternalServerError)
		return
	}

	// Wyświetlenie strony konta
	fmt.Fprintf(w, "<h1>Moje konto</h1><p>E-mail: %s</p><a href='/logout'>Wyloguj</a>", email)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Values["userID"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
