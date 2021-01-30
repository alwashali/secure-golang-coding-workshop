package controllers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"moneylion_assignmnet/dbUtils"
	"moneylion_assignmnet/loger"
	"moneylion_assignmnet/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/csrf"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

// routes functions

func MainPage(w http.ResponseWriter, r *http.Request) {
	data := ""
	t2, err := template.ParseFiles("templates/Index.html")
	if err != nil {
		fmt.Println(err)
	}

	err = t2.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}
}

// LoginPage form (GET)
func LoginPage(w http.ResponseWriter, r *http.Request) {

	// what if he is already logged in
	// redirect to contactform page
	// fmt.Println(r.Header.Get("Cookie"))
	w.Header().Set("X-CSRF-Token", csrf.Token(r))
	w.Header().Set("cache-control", "no-cache, no-store, must-revalidate")
	if ok := strings.Contains(r.Header.Get("Cookie"), "Authorization"); ok {
		authorizationCookie, err := r.Cookie("Authorization")
		if err != nil {
			fmt.Println("Error getting the authorizationCookie", err)
			return
		}
		if authorizationCookie.Value != "" {
			bearerToken := strings.Split(authorizationCookie.Value, " ")
			if len(bearerToken) == 2 {

				valid := utils.VerifyToken(bearerToken[1])
				if valid {
					http.Redirect(w, r, "https://"+r.Host+"/contactform", http.StatusTemporaryRedirect)
					return
				}
			}

		}

	}

	// Not logged in, render normal login form
	t, err := template.ParseFiles("templates/login.html")
	if err != nil {
		fmt.Println(err)
	}
	err = t.Execute(w, map[string]interface{}{csrf.TemplateTag: csrf.TemplateField(r)})
	if err != nil {
		fmt.Println(err)

	}
}

// SignUp API POST
func SignUp(w http.ResponseWriter, r *http.Request) {

	var u dbUtils.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		fmt.Fprintln(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if utils.Validate("email", u.Email) && utils.Validate("username", u.Username) && utils.Validate("password", u.Password) {

		db := dbUtils.ConnectDB()
		defer db.Close()

		// if user does not exist
		if db.Where("username= ?", strings.ToLower(u.Username)).First(&u).RecordNotFound() {
			HashedPassword := utils.HashPasswored([]byte(u.Password))
			user := &dbUtils.User{
				Username: strings.ToLower(u.Username),
				Password: HashedPassword,
				Email:    u.Email,
			}
			db.Create(&user)

			// logging
			loger.Logit(fmt.Sprintf("%s %s", "New User Registered: ", user.Username))

			fmt.Fprintf(w, "Success, You can login now")
		} else { // end of if db.where
			fmt.Fprintln(w, "User Already exist")
		}
	} else { // end of if Validate
		message := `
		Fill all the fields correctly, double check the 
		password, it must contains at least one
		Upper case, 
		Lower case, 
		Number, 
		Special characters
		`
		fmt.Fprintln(w, message)
	}

}

// SignUpPage GET shows the signup form
func SignUpPage(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("X-CSRF-Token", csrf.Token(r))
	t2, err := template.ParseFiles("templates/signup.html")
	if err != nil {
		fmt.Println(err)
	}

	err = t2.Execute(w, map[string]interface{}{csrf.TemplateTag: csrf.TemplateField(r)})
	if err != nil {
		fmt.Println(err)
	}
}

// Logout deletes cookie from client and redirect to home page
func Logout(w http.ResponseWriter, r *http.Request) {

	c := http.Cookie{
		Name:     "Authorization",
		Value:    " ",
		HttpOnly: true,
		Domain:   "localhost",
		// set cookie in the past to be deleted from browser
		Expires: time.Now().Add(time.Hour * -24),
	}
	http.SetCookie(w, &c)

	// logging
	user, err := utils.GetTokenOwner(r)
	if err == nil {
		loger.Logit(fmt.Sprintf("%s %s", user, " logged out"))
	}

	if ok := strings.Contains(r.Header.Get("Cookie"), "Authorization"); ok {
		authorizationCookie, err := r.Cookie("Authorization")
		if err != nil {
			fmt.Println("Error getting the JWT cookie", err)
		}
		if authorizationCookie.Value != "" {
			bearerToken := strings.Split(authorizationCookie.Value, " ")

			if len(bearerToken) == 2 {
				blacklisted := utils.BlackListToken(bearerToken[1])

				if !blacklisted {
					fmt.Println("Error blacklisting token", err)
				}

			}
		}

	}
	http.Redirect(w, r, "https://"+r.Host, http.StatusSeeOther)
	return

}

var limiter = rate.NewLimiter(rate.Every(10*time.Second), 10)

// SignIn api (Post)
func SignIn(w http.ResponseWriter, r *http.Request) {

	if limiter.Allow() == false {
		http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
		return
	}

	var u dbUtils.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		fmt.Fprintln(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	db := dbUtils.ConnectDB()
	defer db.Close()

	if utils.Validate("username", u.Username) {
		user := dbUtils.User{}
		if db.Where("username= ?", strings.ToLower(u.Username)).First(&user).RecordNotFound() {
			// logging
			loger.Logit("Failed Login, user does not exist")
			http.Error(w, "User Does not exist", http.StatusUnauthorized)
			return
		}

		//verify password
		ComparedPassword := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password))

		if ComparedPassword == bcrypt.ErrMismatchedHashAndPassword && ComparedPassword != nil {
			w.WriteHeader(http.StatusUnauthorized)

			// logging
			loger.Logit(fmt.Sprintf("%s %s", user.Username, "failed to login"))

			fmt.Fprintf(w, "Failed Login")
			return
		}

		token, err := utils.GenerateToken(user.Username)
		if err != nil {
			fmt.Println("Error Generating token")
			panic(err)
		}

		c := http.Cookie{
			Name:     "Authorization",
			Value:    "Bearer " + token,
			HttpOnly: true,
			Domain:   "localhost",
			Expires:  time.Now().Add(time.Minute * 10),
			Secure:   true,
			Path:     "/",
		}
		http.SetCookie(w, &c)

		// logging
		loger.Logit(fmt.Sprintf("%s %s", u.Username, ": Successful Login"))

		fmt.Fprintf(w, "Success")
		return
	}
	w.WriteHeader(http.StatusUnauthorized)

	// logging
	loger.Logit(fmt.Sprintf("%s %s", u.Username, " Failed Login"))

	fmt.Fprintf(w, "Failed Login")

}

// SendMessage Actual POST API
func SendMessage(w http.ResponseWriter, r *http.Request) {

	var msg dbUtils.Message
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&msg); err != nil {
		fmt.Fprintln(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if utils.Validate("email", msg.Email) && utils.Validate("username", msg.Name) && utils.Validate("ssn", msg.SSN) && utils.Validate("dob", msg.DOB) {

		user, err := utils.GetTokenOwner(r)
		if err != nil {
			fmt.Fprintf(w, "Error")
			return
		}
		db := dbUtils.ConnectDB()
		defer db.Close()
		db.Create(&dbUtils.Message{Name: msg.Name, DOB: msg.DOB, Email: msg.Email, SSN: msg.SSN, EnteredBy: user})

		//logging
		loger.Logit(fmt.Sprintf("%s %s", user, "added new message"))

		//send response
		fmt.Fprintf(w, "Sent Successfully")
		return

	}

	fmt.Fprintf(w, "Invalid Input Format")
	return
}

// Contactus page
func Contactus(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("X-CSRF-Token", csrf.Token(r))
	w.Header().Set("cache-control", "no-cache, no-store, must-revalidate")

	t, err := template.ParseFiles("templates/contactform.html")
	if err != nil {
		fmt.Println(err)
	}
	err = t.Execute(w, map[string]interface{}{csrf.TemplateTag: csrf.TemplateField(r)})
	if err != nil {
		fmt.Println(err)
	}
}

// VerifyAccess checks token is valid and not expired
func VerifyAccess(next http.HandlerFunc) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check if cookie Authorization exists
		if ok := strings.Contains(r.Header.Get("Cookie"), "Authorization"); ok {
			authorizationCookie, err := r.Cookie("Authorization")
			if err != nil {
				fmt.Println("Error getting the authcookie", err)
			}
			if authorizationCookie.Value != "" {
				bearerToken := strings.Split(authorizationCookie.Value, " ")
				if len(bearerToken) == 2 {

					if valid := utils.VerifyToken(bearerToken[1]); valid {
						next(w, r)
					} else {
						http.Redirect(w, r, "https://"+r.Host+"/login", http.StatusTemporaryRedirect)
					}
				}
			}

		} else {
			// Auth is not found
			// Redirect to login for authentication
			http.Redirect(w, r, "https://"+r.Host+"/login", http.StatusTemporaryRedirect)
		}

	}) // (  closing of HandlerFunc

}
