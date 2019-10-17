package main

import (
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/admin"
)

// User is GORM backend model
type User struct {
	gorm.Model
	Name     string
	Role     int
	Username string `gorm:"not null;unique"`
	Password []byte
	Albums   []Album
}

// Image is GORM backend model
type Image struct {
	gorm.Model
	Name string
	User User
}

// Album is GORM backend model
type Album struct {
	gorm.Model
	Name   string
	Images []Image
	User   User
}

// Session is GORM backend model
type Session struct {
	gorm.Model
	Token  string
	TimeIn string
	User   User
}

const dirDB = `db/`
const fNDB = `dbpbqor.db`

func main() {
	dB, _ := gorm.Open(`sqlite3`, fNDB)
	dB.AutoMigrate(&User{}, &Image{}, &Album{}, &Session{})

	// Initalize
	Admin := admin.New(&admin.AdminConfig{DB: dB})
	/*
		album := Admin.AddResource(&models.Product{}, &admin.Config{Menu: []string{"Product Management"}})
		album.Action(&admin.Action{
			Name: "View On Site",
			URL: func(record interface{}, context *admin.Context) string {
				if product, ok := record.(*models.Product); ok {
					return fmt.Sprintf("/products/%v", product.Code)
				}
				return "#"
			},
			Modes: []string{"menu_item", "edit"},
		})
	*/
	// Allow to use Admin to manage User, Product
	Admin.AddResource(&User{})
	Admin.AddResource(&Album{})
	Admin.AddResource(&Image{})

	// initalize an HTTP request multiplexer
	mux := http.NewServeMux()

	// Mount admin interface to mux
	Admin.MountTo("/admin", mux)

	fmt.Println("Listening on: http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
