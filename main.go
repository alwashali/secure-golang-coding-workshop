package main

import (
	"fmt"
	"log"
	"moneylion_assignmnet/controllers"
	"moneylion_assignmnet/dbUtils"
	"moneylion_assignmnet/loger"
	"moneylion_assignmnet/utils"
	"net/http"
	"os"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/subosito/gotenv"
)

func init() {
	err := gotenv.Load()
	if err != nil {
		fmt.Printf(".env not found\n\n")
		os.Exit(1)
	}

}

func runServer() {

	// rate limit

	route := mux.NewRouter().StrictSlash(true)

	// static route for css
	filesystem := http.FileServer(http.Dir("./templates/css"))
	route.PathPrefix("/templates/css").Handler(http.StripPrefix("/templates/css", filesystem))

	// All app routes
	route.HandleFunc("/", controllers.MainPage).Methods("GET")
	// Login page, view login form page
	route.HandleFunc("/login", controllers.LoginPage).Methods("GET")
	// Login POST for actual login
	route.HandleFunc("/login", controllers.SignIn).Methods("POST")
	route.HandleFunc("/signup", controllers.SignUpPage).Methods("GET")
	route.HandleFunc("/signup", controllers.SignUp).Methods("POST")
	route.HandleFunc("/contact", controllers.VerifyAccess(controllers.SendMessage)).Methods("POST")
	route.HandleFunc("/contactform", controllers.VerifyAccess(controllers.Contactus)).Methods("GET")
	route.HandleFunc("/logout", controllers.Logout).Methods("GET")

	fmt.Println("Up and Running => https://localhost:8080")

	// logging
	loger.Logit("Server started")

	log.Fatal(http.ListenAndServeTLS(":8080", "keys/server.crt", "keys/server.key", csrf.Protect([]byte(utils.CSRFKey))(route)))
}

func main() {
	dbUtils.InitialMigration()
	runServer()

}
