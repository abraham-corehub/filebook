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

var fbA *admin.Admin

func main() {
	initLog()
	dB, _ := gorm.Open("sqlite3", "dbfb.db")
	media.RegisterCallbacks(dB)
	dB.LogMode(true)

	dB.AutoMigrate(
		&User{},
		&Seat{},
		&Department{},
		&Organization{},
		&Branch{},
		&Inward{},
		&Document{},
		&Address{},
		&Sender{},
	)

	fbA = admin.New(&admin.AdminConfig{DB: dB, SiteName: "File Book"})
	fbA.AssetFS.PrependPath(filepath.Join(utils.AppRoot, "views"))

	loadMasters()
	loadResUser()
	loadResInward()
	loadResSeat()
	loadResDept()
	loadResBranch()
	loadResOrg()

	mux := http.NewServeMux()
	fbA.MountTo("/admin", mux)
	for _, path := range []string{"system", "javascripts", "stylesheets", "images"} {
		mux.Handle(fmt.Sprintf("/%s/", path), utils.FileServer(http.Dir("public")))
	}

	log.Println("FileBook Started!")
	log.Println("Listening on: http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func initLog() {
	log.SetPrefix("Log: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

func loadMasters() {
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

	fbA.AddResource(&Inward{}, &confFMgt)
	fbA.AddResource(&User{}, &confAdmn)
	fbA.AddResource(&Seat{}, &confAdmn)
	fbA.AddResource(&Department{}, &confAdmn)
	fbA.AddResource(&Organization{}, &confAdmn)
	fbA.AddResource(&Branch{}, &confAdmn)
}

func loadResUser() {
	userRes := fbA.GetResource("User")
	genderTypes := []string{
		"Male",
		"Female",
		"Other",
		"Unfilled",
	}

	configGenderTypes := &admin.SelectOneConfig{
		Collection: genderTypes,
	}

	userRes.IndexAttrs("-Password")

	userRes.Meta(&admin.Meta{
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

	userRes.Meta(&admin.Meta{
		Name:   "Gender",
		Config: configGenderTypes,
	})

	userRes.Meta(&admin.Meta{
		Name: "Dob",
		Type: "date",
	})

	attrsUser := [][]string{
		{"Seat", "seat"},
		{"Dept.", "department"},
		{"Org.", "organization"},
		{"Department", "department"},
		{"Organization", "organization"},
	}

	valuer := func(interface{}, *qor.Context) interface{} { return "" }
	setter := func(interface{}, *resource.MetaValue, *qor.Context) {}

	for _, aU := range attrsUser {
		switch aU[1] {

		case "seat":
			seat := Seat{}
			valuer = func(r interface{}, c *qor.Context) interface{} {
				fbA.DB.Where("id = ?", r.(*User).SeatID).First(&seat)
				return seat.Name
			}
			setter = func(r interface{}, mV *resource.MetaValue, c *qor.Context) {
				if newValue := utils.ToString(mV.Value); newValue != "" {
					fbA.DB.Where("name = ?", newValue).First(&seat)
					r.(*User).SeatID = seat.ID
				}
			}

		case "department":
			dept := Department{}
			valuer = func(r interface{}, c *qor.Context) interface{} {
				fbA.DB.Where("id = ?", r.(*User).DepartmentID).First(&dept)
				return dept.Name
			}
			setter = func(r interface{}, mV *resource.MetaValue, c *qor.Context) {
				if newValue := utils.ToString(mV.Value); newValue != "" {
					fbA.DB.Where("name = ?", newValue).First(&dept)
					r.(*User).DepartmentID = dept.ID
				}
			}

		case "organization":
			org := Organization{}
			valuer = func(r interface{}, c *qor.Context) interface{} {
				fbA.DB.Where("id = ?", r.(*User).OrganizationID).First(&org)
				return org.Name
			}
			setter = func(r interface{}, mV *resource.MetaValue, c *qor.Context) {
				if newValue := utils.ToString(mV.Value); newValue != "" {
					fbA.DB.Where("name = ?", newValue).First(&org)
					r.(*User).OrganizationID = org.ID
				}
			}
		}

		userRes.Meta(&admin.Meta{
			Name:      aU[0],
			FieldName: aU[1],
			Type:      "string",
			Valuer:    valuer,
			Setter:    setter,
		})
	}

	rowsUserDetailsNE := [][]string{
		{"Name", "Phone"},
		{"Email", "Password"},
		{"Dob", "Gender"},
		{"Addresses"},
		{"Seat"},
		{"Department"},
		{"Organization"},
	}

	rowsUserDetailsI := [][]string{
		{"Name"},
		{"Email", "Phone"},
		{"Dob", "Gender"},
		{"Addresses"},
		{"Seat"},
		{"Dept."},
		{"Org."},
	}

	sectionUserDetailsNE := &admin.Section{
		Rows: rowsUserDetailsNE,
	}

	sectionUserDetailsI := &admin.Section{
		Rows: rowsUserDetailsI,
	}

	userRes.NewAttrs(
		sectionUserDetailsNE,
	)

	userRes.EditAttrs(
		sectionUserDetailsNE,
	)

	userRes.IndexAttrs(
		sectionUserDetailsI,
	)
	userRes.SearchAttrs(
		"ID",
		"Name",
		"Email",
		"Phone",
		"Addresses.Address",
		"Addresses.Pincode",
	)
	/*
		for _, gT := range genderTypes {
			user.Scope(&admin.Scope{
				Name:    gT,
				Group:   "Gender",
				Handler: handlerScope("Gender", gT),
			})
		}
	*/
}

func loadResInward() {
	resInward := fbA.GetResource("Inward")

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
		{"Date"},
		{"Remarks"},
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

	resInward.Meta(&admin.Meta{
		Name: "Type",
		Config: &admin.SelectOneConfig{
			Collection: typesInward,
		},
	})

	resInward.Meta(&admin.Meta{
		Name: "Mode",
		Config: &admin.SelectOneConfig{
			Collection: modesInward,
		},
	})

	resInward.Meta(&admin.Meta{
		Name: "Status",
		Config: &admin.SelectOneConfig{
			Collection: statusesInward,
		},
	})

	resInward.EditAttrs(
		"Title",
		"Sender",
		sectionInwardDetails,
		"Files",
	)

	resInward.NewAttrs(
		"Title",
		"Sender",
		sectionInwardDetails,
		"Status",
		"Files",
	)

	resInward.SearchAttrs(
		"ID",
		"Title",
		"Sender.Name",
	)

	metaSender := resInward.Meta(&admin.Meta{
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
		Setter: func(r interface{}, mV *resource.MetaValue, c *qor.Context) {
			if nV := utils.ToString(mV.Value); nV != "" {
				r.(*Sender).Address = string(nV)
			}
		},
	})

	resInward.Meta(&admin.Meta{
		Name:      "Remarks",
		FieldName: "remarks",
		Type:      "text",
		Setter: func(r interface{}, mV *resource.MetaValue, c *qor.Context) {
			if nV := utils.ToString(mV.Value); nV != "" {
				r.(*Inward).Remarks = string(nV)
			}
		},
	})

	resInward.IndexAttrs(
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
			resInward.Scope(&admin.Scope{
				Name:    f,
				Group:   n,
				Handler: handlerScope(n, f),
			})
		}
	}
}

func handlerScope(n string, f string) func(*gorm.DB, *qor.Context) *gorm.DB {
	return func(db *gorm.DB, context *qor.Context) *gorm.DB {
		return db.Where(n+" = ?", f)
	}
}

func loadResSeat() {
	//resSeat := fbA.GetResource("Seat")

}

func loadResDept() {

}

func loadResBranch() {

}

func loadResOrg() {

}

func strToSHA256(str string) []byte {
	strSHA := sha1.New()
	strSHA.Write([]byte(str))
	strSHAHexStr := hex.EncodeToString(strSHA.Sum(nil))
	return []byte(strSHAHexStr)
}
