package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

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
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/update", updateHandler)

	// Uruchomienie serwera
	log.Fatal(http.ListenAndServe(":8000", nil))
}

// Tworzenie tabeli w SQLite
func createTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);`)
	if err != nil {
		log.Fatal(err)
	}
}

// Strona główna - lista użytkowników
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	rows, err := db.Query("SELECT id, name FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var users []struct {
		ID   int
		Name string
	}

	for rows.Next() {
		var user struct {
			ID   int
			Name string
		}
		rows.Scan(&user.ID, &user.Name)
		users = append(users, user)
	}

	tmpl.Execute(w, users)
}

// Dodawanie użytkownika
func addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		name := r.FormValue("name")
		_, err := db.Exec("INSERT INTO users (name) VALUES (?)", name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, `<li id="user-%s">%s
			<button hx-get="/edit?id=%s" hx-target="#user-%s" hx-swap="outerHTML">Edytuj</button>
			<button hx-delete="/delete?id=%s" hx-target="#user-%s" hx-swap="delete">Usuń</button>
		</li>`, name, name, name, name, name, name)
	}
}

// Usuwanie użytkownika
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		id := r.URL.Query().Get("id")
		_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Pobieranie formularza edycji
func editHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	var name string
	err := db.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&name)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
        <li id="user-%s">
            <form hx-put="/update" hx-target="#user-%s" hx-swap="outerHTML">
                <input type="hidden" name="id" value="%s">
                <input type="text" name="name" value="%s">
                <button type="submit">Zapisz</button>
            </form>
        </li>
    `, id, id, id, name)
}

// Aktualizacja użytkownika
func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		id := r.FormValue("id")
		name := r.FormValue("name")

		_, err := db.Exec("UPDATE users SET name = ? WHERE id = ?", name, id)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprintf(w, `
            <li id="user-%s">
                <span id="name-%s">%s</span>
                <button hx-get="/edit?id=%s" hx-target="#user-%s" hx-swap="outerHTML">Edytuj</button>
                <button hx-delete="/delete?id=%s" hx-target="#user-%s" hx-swap="delete">Usuń</button>
            </li>
        `, id, id, name, id, id, id, id)
	}
}
