package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kataras/go-sessions/v3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var err error

func signup(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "views/signup.html")
		return
	}

	username := req.FormValue("username")
	password := req.FormValue("password")

	var user string

	err := db.QueryRow("SELECT username FROM users WHERE username=?", username).Scan(&user)

	switch {
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}

		_, err = db.Exec("INSERT INTO users(username, password) VALUES(?, ?)", username, hashedPassword)
		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}

		res.Write([]byte("User created!"))
		return
	case err != nil:
		http.Error(res, "Server error, unable to create your account.", 500)
		return
	default:
		http.Redirect(res, req, "/", 200)
	}
}

func login(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "views/login.html")
		return
	}

	username := req.FormValue("username")
	password := req.FormValue("password")

	var databaseUsername string
	var databasePassword string

	err := db.QueryRow("SELECT username, password FROM users WHERE username=?", username).Scan(&databaseUsername, &databasePassword)

	if err != nil {
		http.Redirect(res, req, "/login", 200)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(databasePassword), []byte(password))
	if err != nil {
		http.Redirect(res, req, "/login", 200)
		return
	} else {

		session := sessions.Start(res, req)
		session.Set("username", databaseUsername)
		http.Redirect(res, req, "/home", 200)

		res.Write([]byte("Hello " + databaseUsername))
	}

}

func home(res http.ResponseWriter, req *http.Request) {
	session := sessions.Start(res, req)
	if len(session.GetString("username")) == 0 {
		http.Redirect(res, req, "/login", 200)
	}

	var data = map[string]string{
		"username": session.GetString("username"),
		"message":  "Welcome to the Go !",
	}
	var t, err = template.ParseFiles("views/home.html")
	if err != nil {
		fmt.Println(err.Error())
		return
	} else {
		t.Execute(res, data)
		return
	}

}

func logout(res http.ResponseWriter, req *http.Request) {
	session := sessions.Start(res, req)
	session.Clear()
	sessions.Destroy(res, req)
	http.Redirect(res, req, "/", 200)
}

func index(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "views/index.html")
}

func main() {
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1)/tes_golang")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/", index)
	http.HandleFunc("/home", home)
	http.HandleFunc("/logout", logout)
	fmt.Println("Server running on port :8000")
	http.ListenAndServe(":8000", nil)
}
