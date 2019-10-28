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
	UserID uint
	/*
		ContactName string
		Building    string
		Street      string
		Locality    string
		City        string
	*/
	Address string
	Pincode string
}

// Department Create another GORM-backend model
type Department struct {
	gorm.Model
	Name    string
	Email   string
	Phone   string
	Address string
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
	Phone    string
	Address  string
}

// Organization model
type Organization struct {
	gorm.Model
	Name      string
	Address   string
	Contact   string
	Website   string
	PRContact string
}

// Branch model
type Branch struct {
	gorm.Model
	Name    string
	Address string
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
	confFMgt := admin.Config{
		Menu: []string{
			"File Management",
		},
	}

	confAdmn := admin.Config{
		Menu: []string{
			"Administration",
		},
	}

	ppdA.AddResource(&Inward{}, &confFMgt)
	ppdA.AddResource(&User{}, &confAdmn)
	ppdA.AddResource(&Seat{}, &confAdmn)
	ppdA.AddResource(&Department{}, &confAdmn)
	ppdA.AddResource(&Organization{}, &confAdmn)
	ppdA.AddResource(&Branch{}, &confAdmn)
}

func loadRes(nR string, ppdA *admin.Admin) {
	switch nR {
	case "User":
		user := ppdA.GetResource(nR)
		genderTypes := []string{
			"Male",
			"Female",
			"Other",
			"Unfilled",
		}

		configGenderTypes := &admin.SelectOneConfig{
			Collection: genderTypes,
		}

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
			Name:   "Gender",
			Config: configGenderTypes,
		})

		user.Meta(&admin.Meta{
			Name: "Dob",
			Type: "date",
		})

		// Filter users by gender
		user.Filter(&admin.Filter{
			Name:   "Gender",
			Config: configGenderTypes,
		})

	case "Inward":
		inward := ppdA.GetResource(nR)

		typesInward := []string{
			"Letter",
			"Application",
			"Tender",
			"Invitation",
		}

		modesInward := []string{
			"By Hand",
			"Tele Call",
			"Email",
			"Web Enquiry",
		}

		statusesInward := []string{
			"Received",
			"Opened",
			"Processed",
			"Rejected",
		}

		typesSender := []string{
			"Individual",
			"Department",
			"Organization",
		}

		rowsMenuInwardDetails := [][]string{
			{"Type", "Mode"},
			{"Date", "Remarks"},
		}

		rowsMenuSenderDetails := [][]string{
			{"Type", "Name"},
			{"Email", "Phone"},
			{"Address"},
		}

		sectionInwardDetails := &admin.Section{
			Title: "Inward Details",
			Rows:  rowsMenuInwardDetails,
		}

		sectionSenderDetails := &admin.Section{
			Rows: rowsMenuSenderDetails,
		}

		inward.Meta(&admin.Meta{
			Name: "Type",
			Config: &admin.SelectOneConfig{
				Collection: typesInward,
			},
		})

		inward.Meta(&admin.Meta{
			Name: "Mode",
			Config: &admin.SelectOneConfig{
				Collection: modesInward,
			},
		})

		inward.Meta(&admin.Meta{
			Name: "Status",
			Config: &admin.SelectOneConfig{
				Collection: statusesInward,
			},
		})

		inward.EditAttrs(
			"Title",
			"Sender",
			sectionInwardDetails,
			"Files",
		)

		inward.NewAttrs(
			"Title",
			"Sender",
			sectionInwardDetails,
			"Files",
		)

		metaSender := inward.Meta(&admin.Meta{
			Name: "Sender",
		})

		sndrRes := metaSender.Resource

		sndrRes.EditAttrs(
			sectionSenderDetails,
		)

		sndrRes.NewAttrs(
			sectionSenderDetails,
		)

		sndrRes.Meta(&admin.Meta{
			Name: "Type",
			Config: &admin.SelectOneConfig{
				Collection: typesSender,
			},
		})

		sndrRes.Meta(&admin.Meta{
			Name:      "Address",
			FieldName: "address",
			Type:      "text",
			Setter:    fnSetter("Address"),
		})

		inward.Meta(&admin.Meta{
			Name:      "Remarks",
			FieldName: "remarks",
			Type:      "text",
			Setter:    fnSetter("Remarks"),
		})

		inward.IndexAttrs(
			"-Files",
			"-Remarks",
		)

		fieldsScope := [][]string{
			typesInward,
			modesInward,
			statusesInward,
		}

		namesScope := []string{
			"Type",
			"Mode",
			"Status",
		}

		for i, n := range namesScope {
			for _, f := range fieldsScope[i] {
				inward.Scope(&admin.Scope{
					Name:    f,
					Group:   n,
					Handler: handlerScope(n, f),
				})
			}
		}
	}
}

func fnSetter(s string) func(interface{}, *resource.MetaValue, *qor.Context) {
	return func(r interface{}, mV *resource.MetaValue, c *qor.Context) {
		if nV := utils.ToString(mV.Value); nV != "" {
			switch s {
			case "Address":
				r.(*Sender).Address = string(nV)
			case "Remarks":
				r.(*Inward).Remarks = string(nV)
			}
		}
	}
}

func handlerScope(n string, f string) func(*gorm.DB, *qor.Context) *gorm.DB {
	return func(db *gorm.DB, context *qor.Context) *gorm.DB {
		return db.Where(n+" = ?", f)
	}
}

func strToSHA256(str string) []byte {
	strSHA := sha1.New()
	strSHA.Write([]byte(str))
	strSHAHexStr := hex.EncodeToString(strSHA.Sum(nil))
	return []byte(strSHAHexStr)
}
