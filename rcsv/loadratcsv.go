package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strings"
)

//  CSV file format:
//
//        0    1
//        BUD, RATemplateName
//        REX, RAT001.doc
//        REX, RAT002.doc

// CreateRentalAgreementTemplate creates a database record for the values supplied in sa[]
func CreateRentalAgreementTemplate(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateRentalAgreementTemplate"
	var (
		err error
	)

	const (
		BUD            = 0
		RATemplateName = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"RATemplateName", RATemplateName},
	}

	y, err := ValidateCSVColumnsErr(csvCols, sa, funcname, lineno)
	if y {
		return 1, err
	}
	if lineno == 1 {
		return 0, nil // we've validated the col headings, all is good, send the next line
	}

	des := strings.ToLower(strings.TrimSpace(sa[0])) // this should be BUD

	//-------------------------------------------------------------------
	// Make sure the rlib.Business is in the database
	//-------------------------------------------------------------------
	var a rlib.RentalAgreementTemplate // start the struct we'll be saving
	if len(des) > 0 {                  // make sure it's not empty
		b1, err := rlib.GetBusinessByDesignation(ctx, des) // see if we can find the biz
		if err != nil {
			errMsg := fmt.Sprintf("error while getting business by designation(%s), error: %s", sa[BUD], err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		if len(b1.Designation) == 0 {
			errMsg := fmt.Sprintf("rlib.Business with designation %s does not exist", sa[BUD])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		a.BID = b1.BID
	}

	//-------------------------------------------------------------------
	// Check to see if this assessment type is already in the database
	//-------------------------------------------------------------------
	des = strings.TrimSpace(sa[1]) // this should be the RATemplateName
	if len(des) > 0 {
		// TODO(Steve): ignore error?
		a1, _ := rlib.GetRentalAgreementByRATemplateName(ctx, des)
		if len(a1.RATemplateName) > 0 {
			errMsg := fmt.Sprintf("RentalAgreementTemplate with RATemplateName %s already exists", sa[RATemplateName])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RATemplateName, -1, errMsg)
		}
	}

	a.RATemplateName = des
	_, err = rlib.InsertRentalAgreementTemplate(ctx, &a)
	if err != nil {
		errMsg := fmt.Sprintf("error while inserting RentalAgreementTemplate with RATemplateName %s: %s", a.RATemplateName, err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RATemplateName, -1, errMsg)
	}
	return 0, nil
}

// LoadRentalAgreementTemplatesCSV loads a csv file with assessment types and processes each one
func LoadRentalAgreementTemplatesCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateRentalAgreementTemplate)
}
