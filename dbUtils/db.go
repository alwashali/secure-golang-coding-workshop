package dbUtils

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var database *gorm.DB
var err error

// TokenBlackList contains blacklisted tokens due to logout before expiry time
type TokenBlackList struct {
	Token     string
	ExpiredAt string
}

// Message data structure for contact us form
type Message struct {
	gorm.Model
	Name      string `json:"name"`
	Email     string `json:"email"`
	DOB       string `json:"dob"`
	SSN       string `json:"ssn"`
	EnteredBy string
}

// User data structure for loging
type User struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// ConnectDB returns a database connector object
//Credentials if exist should be put in .env file
func ConnectDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "database.sqlite3")
	if err != nil {
		panic("Database Connection Error")
	}
	return db
}

// InitialMigration setup the tables
func InitialMigration() {
	database, err = gorm.Open("sqlite3", "database.sqlite3")
	if err != nil {
		fmt.Println(err.Error())
		panic("Opps, Failed to connect to database")

	}
	defer database.Close()
	database.AutoMigrate(&Message{}, &User{}, &TokenBlackList{})
}
