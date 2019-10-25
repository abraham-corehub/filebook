package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/ppd/config/bindatafs"

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
	Name      string
	Email     string
	Addresses []Address
}

// Inward Create another GORM-backend model
type Inward struct {
	gorm.Model
	Title     string
	Sender    SenderType
	Mode      string
	Type      string
	Date      time.Time
	Remarks   string
	Documents []Document
	//Docs      []FileType
	Status string
}

// SenderType of Inwards
type SenderType struct {
	gorm.Model
	InwardID  uint
	Name      string
	Type      string
	Email     string
	Addresses []Address
}

// Document is a gorm Model
type Document struct {
	gorm.Model
	InwardID   uint
	Name       string
	Attachment oss.OSS
}

// FileType struct for File Types
type FileType struct {
	gorm.Model
	InwardID uint
	Name     string
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
	dB.AutoMigrate(&User{}, &Department{}, &Inward{}, &Address{})

	ppdA := admin.New(&admin.AdminConfig{DB: dB})
	inward := ppdA.AddResource(&Inward{})
	ppdA.AddResource(&Department{})

	user := ppdA.AddResource(&User{}, &admin.Config{Menu: []string{"User Management"}})
	ppdA.AssetFS.PrependPath(filepath.Join(utils.AppRoot, "views"))
	/*
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
	*/
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

	inward.Meta(&admin.Meta{Name: "Type", Config: &admin.SelectOneConfig{Collection: []string{"Letter", "Application", "Tender", "Invitation"}}})
	inward.Meta(&admin.Meta{Name: "Sender", Config: &admin.SelectOneConfig{Collection: []string{"Individual", "Department", "Organisation"}}})
	inward.Meta(&admin.Meta{Name: "Mode", Config: &admin.SelectOneConfig{Collection: []string{"By Hand", "Tele Call", "Email", "Web Enquiry"}}})
	inward.Meta(&admin.Meta{
		Name:      "Remarks",
		FieldName: "Remarks",
		Type:      "text",
	})

	/*
			inward.Meta(&admin.Meta{
				Name: "Docs",
				Type: "file_upload",
				Valuer: func(record interface{}, context *qor.Context) interface{} {
					l := len(record.(*Inward).Docs)
					if l > 0 {
						return l
					}
					return "Nil"
				},
				Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
					record.(*FileType).Name = utils.ToString(metaValue.Value)
				},
		 })
	*/
	inward.EditAttrs(
		"Title",
		&admin.Section{
			Title: "Received From",
			Rows: [][]string{
				{"Sender", "Mode"},
				{"Type", "Date"},
			},
		},
		"Remarks",
		&admin.Section{
			Rows: [][]string{
				{"Documents"},
			},
		},
	)
	inward.SearchAttrs(
		"Title",
		"Source",
	)

	inward.NewAttrs(
		"Title",
		&admin.Section{
			Title: "Received From",
			Rows: [][]string{
				{"Sender", "Mode"},
				{"Type", "Date"},
			},
		},
		"Remarks",
		&admin.Section{
			Rows: [][]string{
				{"Documents"},
			},
		},
	)

	inward.IndexAttrs("-Status")

	ppdA.MountTo("/admin", mux)
	log.Println("ZOD Started!")
	log.Println("Listening on: http://localhost:8080")
	//log.Println(ppdA.AssetFS.Glob(""))
	http.ListenAndServe(":8080", mux)
}

func main01() {
	assetFS := bindatafs.AssetFS
	// Register view paths into AssetFS
	assetFS.RegisterPath("public")
	assetFS.RegisterPath("public/static")
	//assetFS.RegisterPath("<your_project>/vender/plugin/views")

	// Compile templates under registered view paths into binary
	//assetFS.Compile()

	// Get file content with name
	//fileContent, ok := assetFS.Asset("home/index.tmpl")
	fileContent, ok := assetFS.Asset("testStatic.txt")
	fmt.Println(ok, string(fileContent))
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
