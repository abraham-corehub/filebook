package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/admin"
	"github.com/qor/media"
	"github.com/qor/media/oss"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
)

// User Create a GORM-backend model
type User struct {
	gorm.Model
	Email     string
	Password  string
	Name      sql.NullString
	Gender    string
	Role      string
	Height    int
	Addresses []Address
}

// Address Create a GORM-backend model
type Address struct {
	gorm.Model
	UserID      uint
	ContactName string `form:"contact-name"`
	Phone       string `form:"phone"`
	City        string `form:"city"`
	Address1    string `form:"address1"`
	Address2    string `form:"address2"`
}

// Department Create another GORM-backend model
type Department struct {
	gorm.Model
	Name string
}

// Inward Create another GORM-backend model
type Inward struct {
	gorm.Model
	Title     string
	From      string
	As        string
	Date      time.Time
	Category  string
	Remarks   string
	Documents []Document
	Status    string
}

// Document is a gorm Model
type Document struct {
	gorm.Model
	InwardID uint
	Document oss.OSS
}

// Range slider struct
type Range struct {
	Min    int
	Max    int
	Value  int
	Step   int
	Width  string
	Height string
}

func main() {
	initLog()
	dB, _ := gorm.Open("sqlite3", "dbp.db")
	dB.LogMode(true)

	media.RegisterCallbacks(dB)

	mux := http.NewServeMux()
	for _, path := range []string{"system", "javascripts", "stylesheets", "images"} {
		mux.Handle(fmt.Sprintf("/%s/", path), utils.FileServer(http.Dir("public")))
	}
	dB.AutoMigrate(&User{}, &Department{}, &Inward{}, &Address{}, &Document{})

	ppdA := admin.New(&admin.AdminConfig{DB: dB})
	inward := ppdA.AddResource(&Inward{})
	ppdA.AddResource(&Document{})
	ppdA.AddResource(&Department{})

	user := ppdA.AddResource(&User{}, &admin.Config{Menu: []string{"User Management"}})

	user.Meta(&admin.Meta{
		Name:      "Volume",
		FieldName: "Volume",
		Type:      "range",
		Valuer:    func(interface{}, *qor.Context) interface{} { return "" },
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			record.(*Range).Min = 0
			record.(*Range).Max = 100
			record.(*Range).Value = 20
			record.(*Range).Step = 1
			record.(*Range).Width = "100%"
			record.(*Range).Height = "50px"
		},
	})

	user.Meta(&admin.Meta{
		Name:      "Password",
		FieldName: "Password",
		Type:      "password",
		Valuer:    func(interface{}, *qor.Context) interface{} { return "" },
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if newPassword := utils.ToString(metaValue.Value); newPassword != "" {
				pWSHA := strToSHA256(newPassword)
				record.(*User).Password = string(pWSHA)
			}
		},
	})
	user.IndexAttrs("-Password")
	user.Meta(&admin.Meta{Name: "Role", Config: &admin.SelectOneConfig{Collection: []string{"Admin", "Inward Admin", "Inward User", "Root"}}})
	//user.Meta(&admin.Meta{Name: "Volume", Type: "range"})
	inward.Meta(&admin.Meta{Name: "Category", Config: &admin.SelectOneConfig{Collection: []string{"Tender", "Notice", "Record"}}})

	inward.NewAttrs(
		"Title",
		&admin.Section{
			Title: "Received",
			Rows: [][]string{
				{"From", "As"},
				{"Date"},
			},
		},
		"Category",
		"Remarks",
		"Documents",
	)

	inward.EditAttrs(
		"Title",
		&admin.Section{
			Title: "Received",
			Rows: [][]string{
				{"From", "As"},
				{"Date"},
			},
		},
		"Category",
		"Remarks",
		"Documents",
	)

	inward.IndexAttrs("-Documents", "-Status")
	inward.SearchAttrs("Title", "From", "Date", "Status")

	inward.Meta(&admin.Meta{Name: "As", Config: &admin.SelectOneConfig{Collection: []string{"Physical Document", "Telephone Enquiry"}}})
	inward.Meta(&admin.Meta{
		Name:      "Remarks",
		FieldName: "Remarks",
		Type:      "text",
	})

	ppdA.MountTo("/admin", mux)
	log.Println("ZOD Started!")
	log.Println("Listening on: http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func initLog() {
	log.SetPrefix("Log: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

func strToSHA256(str string) []byte {
	strSHA := sha1.New()
	strSHA.Write([]byte(str))
	strSHAHexStr := hex.EncodeToString(strSHA.Sum(nil))
	return []byte(strSHAHexStr)
}
