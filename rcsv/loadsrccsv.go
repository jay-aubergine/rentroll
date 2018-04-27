package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strings"
)

// CSV FIELDS FOR THIS MODULE
//    0    1     2
//    BUD, Name, Industry
//    REX, FAA,  Aviation
//    REX, IRS,  Excessive Rules

// CreateSourceCSV reads an assessment type string array and creates a database record for the assessment type
func CreateSourceCSV(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateSourceCSV"

	var (
		err error
		a   rlib.DemandSource
	)

	const (
		BUD     = 0
		Name    = iota
		Industy = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"Name", Name},
		{"Industry", Industy},
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
	// Business
	//-------------------------------------------------------------------
	var b rlib.Business
	if len(des) > 0 {
		b, err = rlib.GetBusinessByDesignation(ctx, des)
		if err != nil {
			errMsg := fmt.Sprintf("error while getting business by designation(%s): %s", sa[BUD], err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		if b.BID < 1 {
			errMsg := fmt.Sprintf("CreateRentalSpecialtyType: rlib.Business named %s not found", sa[BUD])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
	}
	a.BID = b.BID

	//-------------------------------------------------------------------
	// Name
	//-------------------------------------------------------------------
	s := strings.TrimSpace(sa[Name])
	if len(s) > 0 {
		var src rlib.DemandSource
		// TODO(Steve): ignore error?
		_ = rlib.GetDemandSourceByName(ctx, b.BID, s, &src)
		if len(src.Name) > 0 {
			errMsg := fmt.Sprintf("DemandSource named %s already exists", sa[Name])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Name, -1, errMsg)
		}
	}
	a.Name = s

	//-------------------------------------------------------------------
	// Industry
	//-------------------------------------------------------------------
	a.Industry = strings.TrimSpace(sa[Industy])

	_, err = rlib.InsertDemandSource(ctx, &a)
	if err != nil {
		errMsg := fmt.Sprintf("error inserting DemandSource: %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	return 0, nil
}

// LoadSourcesCSV loads a csv file with a chart of accounts and creates rlib.GLAccount markers for each
func LoadSourcesCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateSourceCSV)
}
