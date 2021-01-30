package utils

import (
	"errors"
	"fmt"
	"moneylion_assignmnet/dbUtils"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// Key hold environment Variable KEY.
// Values are loaded to system using gotenv package from .env
var Key = os.Getenv("KEY")

// CSRFKey used by gorilla/csrf
var CSRFKey = os.Getenv("CSRFKEY")

// Generate JWT token for authentication and authorization
// verify https://jwt.io/

func GenerateToken(username string) (string, error) {

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Minute * 10).Unix()

	tokenString, err := token.SignedString([]byte(Key))

	if err != nil {
		fmt.Errorf("Error %s", err.Error())
		return "", err
	}

	return tokenString, nil
}

func HashPasswored(pass []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pass, bcrypt.MinCost)
	if err != nil {
		panic(err)
	}

	return string(hash)
}

// Validate is Manually Validating input using regular expressions
// Inputtype tells the type of string to be validated
func Validate(InputType string, value string) bool {

	if len(value) > 0 && len(value) < 100 {

		// golang does not support lookahead regex
		// Regex returns false if empty
		// All input validation here
		usernameRegex := regexp.MustCompile("^[A-Za-z0-9]+$")
		emailRegex := regexp.MustCompile("^[A-Za-z0-9.]+[@]+[A-Za-z0-9]+[.]+[A-Za-z0-9]+$")
		dateRegex := regexp.MustCompile("^[0-9][0-9][/][0-9][0-9][/][0-9][0-9][0-9][0-9]$")
		ssnRegex := regexp.MustCompile("^[0-9][0-9][0-9][-][0-9][0-9][-][0-9][0-9][0-9][0-9]$")
		//password check
		num := `[0-9]{1}`
		az := `[a-z]{1}`
		AZ := `[A-Z]{1}`
		symbol := `[!@#~$%^&*()+|_]{1}`

		switch InputType {
		case "username":
			if usernameRegex.MatchString(value) {
				return true
			}
			return false
		case "email":
			if emailRegex.MatchString(value) {
				return true
			}
			return false
		case "password":
			// password max size already verified above
			if len(value) >= 8 {
				if b, _ := regexp.MatchString(num, value); b {
					if b, _ := regexp.MatchString(az, value); b {
						if b, _ := regexp.MatchString(AZ, value); b {
							if b, _ := regexp.MatchString(symbol, value); b {
								return true
							}
						}

					}

				}

				return false
			}
			return false
		case "ssn":
			if ssnRegex.MatchString(value) {
				return true
			}
			return false
		case "dob":
			if dateRegex.MatchString(value) {
				date := strings.Split(value, "/")
				days, _ := strconv.Atoi(date[0])
				months, _ := strconv.Atoi(date[1])
				years, _ := strconv.Atoi(date[2])
				if days > 0 && days <= 31 && months <= 12 && months >= 1 && years >= 1900 {

					return true
				}

			}
			return false
		}
	} else {
		return false
	}

	return true
}

// VerifyToken send by protecting middleware function
func VerifyToken(bearerToken string) bool {
	token, error := jwt.Parse(bearerToken, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error, unexpected signing method")
		}

		return []byte(Key), nil
	})

	if error != nil {
		return false
	}
	db := dbUtils.ConnectDB()
	defer db.Close()
	blacktoken := dbUtils.TokenBlackList{}

	if token.Valid && db.Where("token= ?", token.Raw).First(&blacktoken).RecordNotFound() {
		return true
	}
	return false
}

// GetTokenOwner extract token from request and return the username
func GetTokenOwner(r *http.Request) (string, error) {

	if ok := strings.Contains(r.Header.Get("Cookie"), "Authorization"); ok {
		authorizationCookie, err := r.Cookie("Authorization")
		if err != nil {
			fmt.Println("Error getting the authcookie", err)
		}
		if authorizationCookie.Value != "" {
			bearerToken := strings.Split(authorizationCookie.Value, " ")
			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(bearerToken[1], claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(Key), nil
			})

			if err != nil {
				return "", errors.New("Error parsing the token")
			}
			if token.Valid {

				// do something with decoded claims
				for key, val := range claims {
					if key == "username" {
						return val.(string), nil
					}
				}
			}

		}
	}
	return "", errors.New("Error parsing the token")
}

// BlackListToken invalidate token when logout
// Useful when logout before expiry time
func BlackListToken(btoken string) bool {
	expiryTime := ""
	db := dbUtils.ConnectDB()
	defer db.Close()
	claims := jwt.MapClaims{}
	parsedToken, _ := jwt.ParseWithClaims(btoken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(Key), nil
	})

	if parsedToken.Valid {
		for key, val := range claims {
			if key == "exp" {

				expTimeInt64, _ := val.(float64)
				expiryTime = time.Unix(int64(expTimeInt64), 0).String()

			}
		}
	}
	db.Create(&dbUtils.TokenBlackList{Token: btoken, ExpiredAt: expiryTime})

	return true
}
