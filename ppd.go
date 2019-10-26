package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
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
	Name           string
	Phone          string
	Email          string
	Password       string
	Dob            time.Time
	Gender         string
	Addresses      []Address
	SeatID         uint
	DepartmentID   uint
	OrganizationID uint
}

// Seat is gorm model
type Seat struct {
	gorm.Model
	Name            string
	UserID          uint
	DelegatedUserID uint
	DepartmentID    uint
	BranchID        uint
	OrganizationID  uint
}

// Address Create a GORM-backend model
type Address struct {
	gorm.Model
	UserID      uint
	ContactName string
	Building    string
	Street      string
	Locality    string
	City        string
	Pincode     string
}

// Department Create another GORM-backend model
type Department struct {
	gorm.Model
	Name    string
	Email   string
	Address Address
}

// Inward Create another GORM-backend model
type Inward struct {
	gorm.Model
	Title   string
	Sender  Sender
	Mode    string
	Type    string
	Date    time.Time
	Remarks string
	Files   []Document
	Status  string
}

// Sender of Inwards
type Sender struct {
	gorm.Model
	InwardID uint
	Name     string
	Type     string
	Email    string
	Address  Address
}

// Organization model
type Organization struct {
	gorm.Model
	Name      string
	Address   Address
	Contact   string
	Website   string
	PRContact string
}

// Branch model
type Branch struct {
	gorm.Model
	Name    string
	Address Address
	Contact string
	Website string
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
	media.RegisterCallbacks(dB)
	dB.LogMode(true)

	dB.AutoMigrate(
		&User{},
		&Seat{},
		&Department{},
		&Organization{},
		&Branch{},
		&Inward{},
		&Address{},
		&Sender{},
	)

	ppdA := admin.New(&admin.AdminConfig{DB: dB})
	ppdA.AssetFS.PrependPath(filepath.Join(utils.AppRoot, "views"))

	loadMasters(ppdA)
	loadRes("User", ppdA)
	loadRes("Inward", ppdA)

	mux := http.NewServeMux()
	ppdA.MountTo("/admin", mux)
	for _, path := range []string{"system", "javascripts", "stylesheets", "images"} {
		mux.Handle(fmt.Sprintf("/%s/", path), utils.FileServer(http.Dir("public")))
	}

	log.Println("ZOD Started!")
	log.Println("Listening on: http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func initLog() {
	log.SetPrefix("Log: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

func loadMasters(ppdA *admin.Admin) {
	fileMConf := admin.Config{
		Menu: []string{
			"File Management",
		},
	}
	admtrnConf := admin.Config{
		Menu: []string{
			"Administration",
		},
	}

	ppdA.AddResource(&Inward{}, &fileMConf)
	ppdA.AddResource(&User{}, &admtrnConf)
	ppdA.AddResource(&Seat{}, &admtrnConf)
	ppdA.AddResource(&Department{}, &admtrnConf)
	ppdA.AddResource(&Organization{}, &admtrnConf)
	ppdA.AddResource(&Branch{}, &admtrnConf)
}

func loadRes(nR string, ppdA *admin.Admin) {
	switch nR {
	case "User":
		user := ppdA.GetResource(nR)
		user.IndexAttrs("-Password")
		user.Meta(&admin.Meta{
			Name:      "Password",
			FieldName: "password",
			Type:      "password",
			Valuer:    func(interface{}, *qor.Context) interface{} { return "" },
			Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
				if newPassword := utils.ToString(metaValue.Value); newPassword != "" {
					//pWSHA := strToSHA256(newPassword)
					record.(*User).Password = string(newPassword)
				}
			},
		})
		user.Meta(&admin.Meta{
			Name: "Gender",
			Config: &admin.SelectOneConfig{
				Collection: []string{
					"Male",
					"Female",
					"Other",
				},
			},
		})

		user.Meta(&admin.Meta{
			Name: "Dob",
			Type: "date",
		})

	case "Inward":
		inward := ppdA.GetResource(nR)
		inward.Meta(&admin.Meta{
			Name: "Type",
			Config: &admin.SelectOneConfig{
				Collection: []string{
					"Letter",
					"Application",
					"Tender",
					"Invitation",
				},
			},
		})

		inward.Meta(&admin.Meta{
			Name: "Mode",
			Config: &admin.SelectOneConfig{
				Collection: []string{
					"By Hand",
					"Tele Call",
					"Email",
					"Web Enquiry",
				},
			},
		})

		inward.Meta(&admin.Meta{
			Name: "Status",
			Config: &admin.SelectOneConfig{
				Collection: []string{
					"Received",
					"Opened",
					"Processed",
					"Rejected",
				},
			},
		})

		inward.EditAttrs(
			"Title",
			"Sender",
			&admin.Section{
				Title: "Inward Details",
				Rows: [][]string{
					{"Type", "Mode"},
					{"Date", "Remarks"},
					{"Status"},
				},
			},
			"Files",
		)

		inward.ShowAttrs(
			"Title",
			"Sender",
			&admin.Section{
				Title: "Inward Details",
				Rows: [][]string{
					{"Type", "Mode"},
					{"Date", "Remarks"},
					{"Status"},
				},
			},
			"Files",
		)
		senderMeta := inward.Meta(&admin.Meta{
			Name: "Sender",
		})

		senderResource := senderMeta.Resource
		senderResource.EditAttrs(
			&admin.Section{
				Rows: [][]string{
					{"Type", "Name"},
					{"Email"},
					{"Address"},
				},
			})

		senderResource.ShowAttrs(
			"Type",
			"Name",
			"Email",
			"Address",
		)

		senderResource.Meta(&admin.Meta{
			Name: "Type",
			Config: &admin.SelectOneConfig{
				Collection: []string{
					"Individual",
					"Department",
					"Organization",
				},
			},
		})
		inward.Meta(&admin.Meta{
			Name:      "Remarks",
			FieldName: "Remarks",
			Type:      "text",
		})

		inward.IndexAttrs(
			"-Files",
			"-Remarks",
		)
	}
}

func strToSHA256(str string) []byte {
	strSHA := sha1.New()
	strSHA.Write([]byte(str))
	strSHAHexStr := hex.EncodeToString(strSHA.Sum(nil))
	return []byte(strSHAHexStr)
}
