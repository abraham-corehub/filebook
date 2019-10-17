package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/admin"
	"github.com/qor/auth"
	"github.com/qor/auth/auth_identity"
	"github.com/qor/auth/providers/password"
)

// User Create a GORM-backend model
type User struct {
	gorm.Model
	Name string
}

// Department Create another GORM-backend model
type Department struct {
	gorm.Model
	Name string
}

// Inward Create another GORM-backend model
type Inward struct {
	gorm.Model
	Title string
	Type  string
	Date  time.Time
}

func main() {
	dB, _ := gorm.Open("sqlite3", "dbpbqor.db")
	dB.AutoMigrate(&User{}, &Department{}, &Inward{})
	dB.AutoMigrate(&auth_identity.AuthIdentity{})

	Admin := admin.New(&admin.AdminConfig{DB: dB})
	Auth := auth.New(&auth.Config{DB: dB})

	Auth.RegisterProvider(password.New(&password.Config{}))
	Admin.AddResource(&User{})
	Admin.AddResource(&Department{})
	Admin.AddResource(&Inward{})

	mux := http.NewServeMux()

	Admin.MountTo("/admin", mux)

	fmt.Println("Listening on: http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
