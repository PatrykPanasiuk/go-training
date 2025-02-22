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
	fmt.Println("Serwer startuje...")

	// Połączenie z bazą danych
	var err error
	db, err = sql.Open("sqlite", "database.db")
	if err != nil {
		log.Fatal(err)
	}

	createTable()

	defer db.Close()

	// Obsługa strony głównej
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/add", addHandler)

	// Uruchomienie serwera
	http.ListenAndServe(":8000", nil)
}

func createTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);`)
	if err != nil {
		log.Fatal(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	// Pobieramy użytkowników z bazy
	rows, err := db.Query("SELECT id, name FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var users []struct {
		ID   int
		Name string
	}

	// Wczytujemy użytkowników do tablicy
	for rows.Next() {
		var user struct {
			ID   int
			Name string
		}
		rows.Scan(&user.ID, &user.Name)
		users = append(users, user)
	}

	// Renderujemy HTML i przekazujemy użytkowników
	tmpl.Execute(w, users)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		name := r.FormValue("name") // Pobieramy wartość z formularza
		_, err := db.Exec("INSERT INTO users (name) VALUES (?)", name)
		if err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther) // Przekierowanie na stronę główną
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	var name string
	err := db.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&name)
	if err != nil {
		log.Fatal(err)
	}

	// Zwracamy formularz edycji użytkownika
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
