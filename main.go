// ref: https://elijahomolo.medium.com/user-creation-and-authentication-in-golang-part-1-f6b6cc08d9fe
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// define a user model
type User struct {
	Id       int
	Username string
	City     string
	Email    string
	Password string
}

// load .env file
func goDotEnvVariable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}

// connect to the database and return it as an object
func dbConn() (db *sql.DB) {
	// pass the db credentials into variables
	host := goDotEnvVariable("DBHOST")
	port := goDotEnvVariable("DBPORT")
	dbUser := goDotEnvVariable("DBUSER")
	dbPass := goDotEnvVariable("DBPASS")
	dbname := goDotEnvVariable("DBNAME")
	// create a connection string
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, dbUser, dbPass, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	return db
}

//define a template
var tmpl = template.Must(template.ParseGlob("forms/*"))

//define an Index function that includes the http write and read parameters
func Index(w http.ResponseWriter, r *http.Request) {
	//connect to the database
	db := dbConn()
	// Query the database and return all rows from the user table
	rows, err := db.Query(`SELECT "user_id", "username", "city", "email" FROM public."users"`)
	//Handle any errors
	CheckError(err)
	//Define and populate a User struct from the returned data from the query
	usr := User{}
	//The list of Users that will be passed to the html template
	res := []User{}
	//Loop through each row and populate a User
	for rows.Next() {
		var id int
		var username, city, email string
		err = rows.Scan(&id, &username, &city, &email)
		CheckError(err)
		usr.Id = id
		usr.Email = email
		usr.Username = username
		usr.City = city
		res = append(res, usr)
	}
	//write to the Index template that will range through the User struct displaying a field for the returned data
	tmpl.ExecuteTemplate(w, "Index", res)
	//close the database connection
	defer db.Close()
}

func New(w http.ResponseWriter, r *http.Request) {
	//Execute the template New which will pass the input to /insert
	tmpl.ExecuteTemplate(w, "New", nil)
}

func Edit(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	nId := r.URL.Query().Get("id")
	rows, err := db.Query(`SELECT * FROM public."users" WHERE "user_id"=$1`, nId)
	CheckError(err)
	usr := User{}
	for rows.Next() {
		var id int
		var username, city, email, password string
		err = rows.Scan(&id, &username, &city, &email, &password)
		CheckError(err)
		usr.Id = id
		usr.Username = username
		usr.Password = password
		usr.Email = email
		usr.City = city
	}
	tmpl.ExecuteTemplate(w, "Edit", usr)
	//defer db.Close()
}

func Show(w http.ResponseWriter, r *http.Request) {
	//connect to the db
	db := dbConn()
	//assign a variable to the id passed in the URL when view is clicked
	nId := r.URL.Query().Get("id")
	//run a query against the db filtering the user_id table using the passed id
	rows, err := db.Query(`SELECT * FROM public."users" WHERE "user_id"=$1`, nId)
	//handle error
	CheckError(err)
	//construct a User
	usr := User{}
	for rows.Next() {
		var id int
		var username, city, email, password string
		err = rows.Scan(&id, &username, &city, &password, &email)
		CheckError(err)
		usr.Id = id
		usr.Username = username
		usr.Email = email
		usr.City = city
	}
	//Execute the Show template using the Users data
	tmpl.ExecuteTemplate(w, "Show", usr)
	defer db.Close()
}

func Insert(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	//If it's a post request, assign a variable to the value returned in each field of the New page.
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		city := r.FormValue("city")
		email := r.FormValue("email")
		//prepare a query to insert the data into the database
		insForm, err := db.Prepare(`INSERT INTO public.users(username,password, city, email) VALUES ($1,$2, $3, $4)`)
		//check for  and handle any errors
		CheckError(err)
		//execute the query using the form data
		insForm.Exec(username, password, city, email)
		//print out added data in terminal
		log.Println("INSERT: Username: " + username + " | City: " + city + " | Email: " + email)
	}
	defer db.Close()
	//redirect to the index page
	http.Redirect(w, r, "/", 301)
}

func Update(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		username := r.FormValue("username")
		city := r.FormValue("city")
		email := r.FormValue("email")
		password := r.FormValue("password")
		id := r.FormValue("uid")
		insForm, err := db.Prepare(`UPDATE public."users" SET username=$1, city=$2, email=$3, password=$4 WHERE "user_id"=$5`)
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(username, city, email, password, id)
		log.Println("UPDATE: Username: " + username)
	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	usr := r.URL.Query().Get("id")
	delForm, err := db.Prepare(`DELETE FROM public."users" WHERE user_id=$1`)
	if err != nil {
		panic(err.Error())
	}
	delForm.Exec(usr)
	log.Println("DELETE")
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	//Provide address server will be provided on
	log.Println("Server started on: http://localhost:8080")
	//Create and serve a route for the Index function
	http.HandleFunc("/", Index)
	http.HandleFunc("/show", Show)
	http.HandleFunc("/edit", Edit)
	http.HandleFunc("/new", New)
	http.HandleFunc("/insert", Insert)
	http.HandleFunc("/update", Update)
	http.HandleFunc("/delete", Delete)
	http.ListenAndServe(":8080", nil)
}
