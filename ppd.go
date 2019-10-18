package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/admin"
	"github.com/qor/media"
	"github.com/qor/media/filesystem"
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
	Title    string
	From     string
	As       string
	Date     time.Time
	Category string
	Remarks  string
	Document filesystem.FileSystem
	Status   string
}

func main() {
	dB, _ := gorm.Open("sqlite3", "dbp.db")
	media.RegisterCallbacks(dB)

	mux := http.NewServeMux()
	for _, path := range []string{"system", "javascripts", "stylesheets", "images"} {
		mux.Handle(fmt.Sprintf("/%s/", path), utils.FileServer(http.Dir("public")))
	}
	dB.AutoMigrate(&User{}, &Department{}, &Inward{}, &Address{})

	ppdA := admin.New(&admin.AdminConfig{DB: dB})
	inward := ppdA.AddResource(&Inward{})
	ppdA.AddResource(&Department{})
	user := ppdA.AddResource(&User{}, &admin.Config{Menu: []string{"User Management"}})

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
		"Document",
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
		"Document",
	)

	inward.Meta(&admin.Meta{Name: "As", Config: &admin.SelectOneConfig{Collection: []string{"Physical Document", "Telephone Enquiry"}}})
	inward.Meta(&admin.Meta{
		Name:      "Remarks",
		FieldName: "Remarks",
		Type:      "text",
	})

	ppdA.MountTo("/admin", mux)

	fmt.Println("Listening on: http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func strToSHA256(str string) []byte {
	strSHA := sha1.New()
	strSHA.Write([]byte(str))
	strSHAHexStr := hex.EncodeToString(strSHA.Sum(nil))
	return []byte(strSHAHexStr)
}
