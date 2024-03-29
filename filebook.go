package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
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
	"github.com/qor/oss/filesystem"
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
	BranchID       uint
	OrganizationID uint
}

// Seat is gorm model
type Seat struct {
	gorm.Model
	DepartmentID uint
	Name         string
	Code         string
	/*
		UserID          uint
		DelegatedUserID uint
	*/
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
	Seat    []Seat
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

var fbA *admin.Admin

const dirData = "data"
const nameFileDB = "dbfb.db"

func main() {
	initLog()
	dB, _ := gorm.Open("sqlite3", dirData+"/"+nameFileDB)
	//dB.LogMode(true)

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
	//loadResUser()
	//loadResInward()
	//loadResSeat()
	loadResDept()
	//loadResBranch()
	//loadResOrg()

	mux := http.NewServeMux()

	router := fbA.GetRouter()

	router.Post("/ajax", func(c *admin.Context) {
		handlerAjax(c)
	})

	fbA.MountTo("/admin", mux)

	// Data Storage
	media.RegisterCallbacks(dB)
	mux.Handle("/data/", utils.FileServer(http.Dir(".")))
	oss.URLTemplate = "/data/{{class}}/{{primary_key}}/{{column}}/{{filename_with_hash}}"
	oss.Storage = filesystem.New(".")

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
	fbA.AddResource(&Sender{}, &admin.Config{Invisible: true})
	fbA.AddResource(&User{}, &confAdmn)
	fbA.AddResource(&Seat{}, &admin.Config{Invisible: true})
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

	userRes.Meta(&admin.Meta{
		Name:      "Seat",
		FieldName: "Seat",
		Type:      "string",
		Valuer: func(r interface{}, c *qor.Context) interface{} {
			seat := &Seat{}
			c.DB.Where("id = ?", r.(*User).SeatID).First(&seat)
			return seat.Name
		},
	})

	userRes.Meta(&admin.Meta{
		Name:      "Dept.",
		FieldName: "Department",
		Type:      "string",
		Valuer: func(r interface{}, c *qor.Context) interface{} {
			dept := &Department{}
			c.DB.Where("id = ?", r.(*User).DepartmentID).First(&dept)
			return dept.Name
		},
	})

	userRes.Meta(&admin.Meta{
		Name:      "Branch",
		FieldName: "Branch",
		Type:      "string",
		Valuer: func(r interface{}, c *qor.Context) interface{} {
			branch := &Branch{}
			c.DB.Where("id = ?", r.(*User).BranchID).First(&branch)
			return branch.Name
		},
	})

	userRes.Meta(&admin.Meta{
		Name:      "Org.",
		FieldName: "Organization",
		Type:      "string",
		Valuer: func(r interface{}, c *qor.Context) interface{} {
			org := &Organization{}
			c.DB.Where("id = ?", r.(*User).OrganizationID).First(&org)
			return org.Name
		},
	})

	rowsUserDetailsNE := [][]string{
		{"Name", "Phone"},
		{"Email", "Password"},
		{"Dob", "Gender"},
		{"Addresses"},
	}

	attrsUserDetailsI := []string{
		"Name",
		"Email",
		"Phone",
		"Dob",
		"Gender",
		"Seat",
		"Branch",
		"Dept.",
		"Org.",
	}

	sectionUserDetailsNE := &admin.Section{
		Rows: rowsUserDetailsNE,
	}

	userRes.NewAttrs(
		sectionUserDetailsNE,
	)

	userRes.EditAttrs(
		sectionUserDetailsNE,
	)

	userRes.IndexAttrs(
		attrsUserDetailsI,
	)
	userRes.SearchAttrs(
		"ID",
		"Name",
		"Email",
		"Phone",
		"Addresses.Address",
		"Addresses.Pincode",
	)

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

	rowsMenuInwardDetails := [][]string{
		{"Type", "Mode"},
		{"Date"},
		{"Remarks"},
	}

	sectionInwardDetails := &admin.Section{
		Title: "Inward Details",
		Rows:  rowsMenuInwardDetails,
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

	attrsInward := []string{
		"Title",
		"Sender",
		"Status",
		"Files",
	}

	resInward.EditAttrs(
		attrsInward[0],
		sectionInwardDetails,
		attrsInward[1],
		attrsInward[3],
	)

	resInward.NewAttrs(
		attrsInward[0],
		sectionInwardDetails,
		attrsInward[1],
		attrsInward[2],
		attrsInward[3],
	)

	resInward.SearchAttrs(
		"ID",
		"Title",
		"Sender.Name",
	)

	/*
		metaSender := resInward.Meta(&admin.Meta{
			Name: "Sender",
		})

		typesSender := []string{
			"Individual",
			"Department",
			"Organization",
		}

		rowsSectionSenderDetails := [][]string{
			{"Type", "Name"},
			{"Email", "Phone"},
			{"Address"},
		}

		resSender := metaSender.Resource
		sectionSenderDetails := &admin.Section{
			Rows: rowsSectionSenderDetails,
		}
		resSender.EditAttrs(
			sectionSenderDetails,
		)

		resSender.NewAttrs(
			sectionSenderDetails,
		)

		resSender.Meta(&admin.Meta{
			Name: "Type",
			Config: &admin.SelectOneConfig{
				Collection: typesSender,
			},
		})

		resSender.Meta(&admin.Meta{
			Name:      "Address",
			FieldName: "address",
			Type:      "text",
			Setter: func(r interface{}, mV *resource.MetaValue, c *qor.Context) {
				if nV := utils.ToString(mV.Value); nV != "" {
					r.(*Sender).Address = string(nV)
				}
			},
		})
	*/

	resSender := fbA.GetResource("Sender")

	resInward.Meta(&admin.Meta{
		Name:      "Sender",
		FieldName: "sender",
		Type:      "string",
		Config: &admin.SelectOneConfig{
			SelectMode:         "bottom_sheet",
			RemoteDataResource: resSender,
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
			handlerS := func(n string, f string) func(*gorm.DB, *qor.Context) *gorm.DB {
				return func(db *gorm.DB, context *qor.Context) *gorm.DB {
					return db.Where(n+" = ?", f)
				}
			}
			resInward.Scope(&admin.Scope{
				Name:    f,
				Group:   n,
				Handler: handlerS(n, f),
			})
		}
	}
}

func loadResSeat() {
	resSeat := fbA.GetResource("Seat")

	attrsSeat := []string{
		"Name",
		//"User",
		//"Delegated User",
		"Organization",
		"Branch",
		"Department",
	}

	sectionSeatDetails := &admin.Section{
		Rows: [][]string{
			attrsSeat[1:],
		},
	}

	resSeat.IndexAttrs(
		attrsSeat[0],
		sectionSeatDetails,
	)
	resSeat.NewAttrs(
		attrsSeat[0],
		sectionSeatDetails,
	)
	resSeat.EditAttrs(
		attrsSeat[0],
		sectionSeatDetails,
	)

	resSeat.SearchAttrs(
		attrsSeat...,
	)

	oldSearchHandler := resSeat.SearchHandler
	resSeat.SearchHandler = func(keyword string, context *qor.Context) *gorm.DB {
		return oldSearchHandler(keyword, context)
	}

	orgs := []Organization{}
	fbA.DB.Find(&orgs)
	nO := make([]string, 0)
	for _, org := range orgs {
		nO = append(nO, org.Name)
	}

	branches := []Branch{}
	fbA.DB.Find(&branches)
	nB := make([]string, 0)
	for _, branches := range branches {
		nB = append(nB, branches.Name)
	}

	depts := []Department{}
	fbA.DB.Find(&depts)
	nD := make([]string, 0)
	for _, dept := range depts {
		nD = append(nD, dept.Name)
	}
	/*
		metaOrg := admin.Meta{
			Name:      "Organization",
			FieldName: "Organization",
			Type:      "string",
			Config: &admin.SelectOneConfig{
				Collection: nO,
			},
			Valuer: func(r interface{}, c *qor.Context) interface{} {
				org := &Organization{}
				c.DB.Where("id = ?", r.(*Seat).OrganizationID).First(&org)
				return org.Name
			},
			Setter: func(r interface{}, mV *resource.MetaValue, c *qor.Context) {
				org := &Organization{}
				if v := utils.ToString(mV.Value); v != "" {
					c.DB.Where("name = ?", v).First(&org)
					r.(*Seat).OrganizationID = org.ID
				}
			},
		}
	*/

	//resBranch := fbA.GetResource("Branch")
	/*
		metaBranch := admin.Meta{
			Name:      "Branch",
			FieldName: "Branch",
			Type:      "string",
			Config: &admin.SelectOneConfig{
				Collection: nB,
			},
			Valuer: func(r interface{}, c *qor.Context) interface{} {
				branch := &Branch{}
				c.DB.Where("id = ?", r.(*Seat).BranchID).First(&branch)
				return branch.Name
			},
			Setter: func(r interface{}, mV *resource.MetaValue, c *qor.Context) {
				branch := &Branch{}
				if v := utils.ToString(mV.Value); v != "" {
					c.DB.Where("name = ?", v).First(&branch)
					r.(*Seat).BranchID = branch.ID
				}
			},
		}
	*/

	//resDept := fbA.GetResource("Department")
	metaDept := admin.Meta{
		Name:      "Department",
		FieldName: "Department",
		Type:      "string",
		/*
			Config: &admin.SelectOneConfig{
				SelectMode:         "select_async",
				RemoteDataResource: resDept,
			},
		*/
		Config: &admin.SelectOneConfig{
			Collection: nD,
		},
		Valuer: func(r interface{}, c *qor.Context) interface{} {
			dept := &Department{}
			c.DB.Where("id = ?", r.(*Seat).DepartmentID).First(&dept)
			return dept.Name
		},
		Setter: func(r interface{}, mV *resource.MetaValue, c *qor.Context) {
			dept := &Department{}
			if v := utils.ToString(mV.Value); v != "" {
				// v is "ID" when admin.SelectOneConfig is configured as RemoteDataSource
				c.DB.Where("name = ?", v).First(&dept)
				r.(*Seat).DepartmentID = dept.ID
			}
		},
	}

	//resSeat.Meta(&metaOrg)
	resSeat.Meta(&metaDept)
	//resSeat.Meta(&metaBranch)
}

func loadResDept() {
	resDept := fbA.GetResource("Department")
	resSeat := fbA.GetResource("Seat")

	resDept.Meta(&admin.Meta{
		Name:      "Seat",
		FieldName: "seat",
		Type:      "string",
		Config: &admin.SelectOneConfig{
			SelectMode:         "bottom_sheet",
			RemoteDataResource: resSeat,
		},
	})

}

func loadResBranch() {

}

func loadResOrg() {

}

func handlerAjax(c *admin.Context) {
	// do somehing here
	c.Request.ParseForm()
	res := c.Request.Form.Get("res")
	id := c.Request.Form.Get("id")
	field := c.Request.Form.Get("field")
	value := c.Request.Form.Get("value")
	fmt.Println("Ajax Request!, data:", res, id, field, value)

	type Response struct {
		Name string
	}
	r := Response{}
	r.Name = "Abraham"
	w := c.Writer
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(r)
}

func strToSHA256(str string) []byte {
	strSHA := sha1.New()
	strSHA.Write([]byte(str))
	strSHAHexStr := hex.EncodeToString(strSHA.Sum(nil))
	return []byte(strSHAHexStr)
}
