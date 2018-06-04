package onesite

import (
	"reflect"
	"rentroll/importers/core"
	"strings"
)

// CSVFieldMap is struct which contains several categories
// used to store the data from onesite to rentroll system
type CSVFieldMap struct {
	RentableTypeCSV    core.RentableTypeCSV
	PeopleCSV          core.PeopleCSV
	RentableCSV        core.RentableCSV
	RentalAgreementCSV core.RentalAgreementCSV
	CustomAttributeCSV core.CustomAttributeCSV
}

// csvColumnFieldMap contains internal OneSite Structure fields
// to csv columns, used to refer columns from struct fields
var csvColumnFieldMap = map[string]string{
	"bldgunit":        "Unit",
	"floorplan":       "FloorPlan",
	"unitdesignation": "UnitDesignation",
	"sqft":            "SQFT",
	"unitleasestatus": "UnitLeaseStatus",
	"name":            "Name",
	"phonenumber":     "PhoneNumber",
	"email":           "Email",
	"movein":          "MoveIn",
	"moveout":         "MoveOut",
	"leasestart":      "LeaseStart",
	"leaseend":        "LeaseEnd",
	"marketaddl":      "MarketAddl",
	"rent":            "Rent",
	// "tax":              "TAX",
}

var marketAddl = "marketaddl"
var marketRent = "marketrent"

// CSVRow contains fields which represents value
// exactly to the each raw of onesite input csv file
type CSVRow struct {
	Unit            string
	FloorPlan       string
	UnitDesignation string
	SQFT            string
	UnitLeaseStatus string
	Name            string
	PhoneNumber     string
	Email           string
	MoveIn          string
	MoveOut         string
	LeaseStart      string
	LeaseEnd        string
	MarketAddl      string
	Rent            string
	// Tax             string
}

// CSVHeader contains index of headers and tells if that header is optional or not
type CSVHeader struct {
	Index      int
	IsOptional bool
}

// getCSVHeadersMap returns the map of fields with
// undetermined indexes
func getCSVHeadersMap() map[string]CSVHeader {

	// csvHeadersIndex holds the map of headers with its index
	csvHeadersIndex := map[string]CSVHeader{
		"Unit":            CSVHeader{-1, false},
		"FloorPlan":       CSVHeader{-1, false},
		"UnitDesignation": CSVHeader{-1, false},
		"SQFT":            CSVHeader{-1, false},
		"UnitLeaseStatus": CSVHeader{-1, false},
		"Name":            CSVHeader{-1, false},
		"PhoneNumber":     CSVHeader{-1, true},
		"Email":           CSVHeader{-1, true},
		"MoveIn":          CSVHeader{-1, false},
		"MoveOut":         CSVHeader{-1, false},
		"LeaseStart":      CSVHeader{-1, false},
		"LeaseEnd":        CSVHeader{-1, false},
		"MarketAddl":      CSVHeader{-1, false},
		"Rent":            CSVHeader{-1, false},
		// "Tax":             -1,
	}

	return csvHeadersIndex
}

// loadOneSiteCSVRow used to load data from slice
// into CSVRow struct and return that struct
func loadOneSiteCSVRow(csvHeadersIndex map[string]CSVHeader, data []string) (bool, CSVRow) {
	csvRow := reflect.New(reflect.TypeOf(CSVRow{}))
	rowLoaded := false

	for header, index := range csvHeadersIndex {
		if !index.IsOptional {
			value := strings.TrimSpace(data[index.Index])
			csvRow.Elem().FieldByName(header).Set(reflect.ValueOf(value))
		} else {
			value := ""
			csvRow.Elem().FieldByName(header).Set(reflect.ValueOf(value))
		}
	}

	// if blank data has not been passed then only need to return true
	if (CSVRow{}) != csvRow.Elem().Interface().(CSVRow) {
		rowLoaded = true
	}

	return rowLoaded, csvRow.Elem().Interface().(CSVRow)
}
