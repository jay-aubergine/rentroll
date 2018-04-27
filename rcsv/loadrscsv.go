package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strconv"
	"strings"
)

//  CSV file format:
// 0            1                2      3
// Designation, Name,            Fee,   Description
// REH,         "Lake View",     50.0,  "Overlooks the lake"
// REH,         "Courtyard View",50.0,  "Rear windows view the courtyard"
// REH,         "Top Floor",     100.0, "Penthouse"
// REH,         "Fireplace",     20.0,  "Wood burning, gas fireplace"

// CreateRentalSpecialty reads a rental specialty type string array and creates a database record for the rental specialty type.
func CreateRentalSpecialty(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateRentalSpecialty"
	var (
		err error
	)

	const (
		BUD         = 0
		Name        = iota
		Fee         = iota
		Description = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"Name", Name},
		{"Fee", Fee},
		{"Description", Description},
	}

	y, err := ValidateCSVColumnsErr(csvCols, sa, funcname, lineno)
	if y {
		return 1, err
	}
	if lineno == 1 {
		return 0, nil // we've validated the col headings, all is good, send the next line
	}

	des := strings.ToLower(strings.TrimSpace(sa[BUD]))

	//-------------------------------------------------------------------
	// Check to see if this rental specialty type is already in the database
	//-------------------------------------------------------------------
	var b rlib.Business

	if len(des) > 0 {
		b, err = rlib.GetBusinessByDesignation(ctx, des)
		if err != nil {
			errMsg := fmt.Sprintf("error while getting business by designation(%s), error: %s", sa[BUD], err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		if b.BID < 1 {
			errMsg := fmt.Sprintf("rlib.Business named %s not found", sa[BUD])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
	}

	var a rlib.RentableSpecialty
	var x float64

	a.Name = strings.TrimSpace(sa[Name])
	if x, err = strconv.ParseFloat(strings.TrimSpace(sa[Fee]), 64); err != nil {
		errMsg := fmt.Sprintf("Invalid floating point number: %s", sa[Fee])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Fee, -1, errMsg)
	}
	a.Fee = x
	a.Description = strings.TrimSpace(sa[Description])
	a.BID = b.BID

	//-------------------------------------------------------------------
	// Make sure we don't already have an exact rlib.Business,name match
	//-------------------------------------------------------------------
	// TODO(Steve): ignore error?
	rsp, _ := rlib.GetRentableSpecialtyTypeByName(ctx, a.BID, a.Name)
	if rsp.RSPID > 0 {
		errMsg := fmt.Sprintf("rlib.Business %s already has a rlib.RentableSpecialty named %s", sa[BUD], a.Name)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// OK, just insert the record and we're done
	//-------------------------------------------------------------------
	_, err = rlib.InsertRentableSpecialty(ctx, &a)
	if nil != err {
		errMsg := fmt.Sprintf("error inserting RentalSpecialty = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	return 0, nil
}

// LoadRentalSpecialtiesCSV loads a csv file with rental specialty types and processes each one
func LoadRentalSpecialtiesCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateRentalSpecialty)
}
