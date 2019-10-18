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
	Title string
	Type  string
	Date  time.Time
}

func main() {
	dB, _ := gorm.Open("sqlite3", "dbp.db")
	dB.AutoMigrate(&User{}, &Department{}, &Inward{}, &Address{})

	ppdA := admin.New(&admin.AdminConfig{DB: dB})

	ppdA.AddResource(&Department{})
	ppdA.AddResource(&Inward{})
	user := ppdA.AddResource(&User{}, &admin.Config{Menu: []string{"User Management"}})
	user.Meta(&admin.Meta{
		Name:      "Password",
		FieldName: "Password",
		Type:      "password",
		Valuer:    func(interface{}, *qor.Context) interface{} { return "" },
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if newPassword := utils.ToString(metaValue.Value); newPassword != "" {
				shaPW := strToSHA256(newPassword)
				record.(*User).Password = string(shaPW)
			}
		},
	})
	user.IndexAttrs("-Password")
	mux := http.NewServeMux()

	ppdA.MountTo("/admin", mux)

	fmt.Println("Listening on: http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func strToSHA256(str string) []byte {
	pW := str
	pWH := sha1.New()
	pWH.Write([]byte(pW))
	pWHS := hex.EncodeToString(pWH.Sum(nil))
	return []byte(pWHS)
}
