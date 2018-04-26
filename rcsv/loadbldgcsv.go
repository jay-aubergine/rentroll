package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strconv"
	"strings"
)

//  CSV file format:
// 0           1      2       3        4    5     6          7
// BUD,BldgNo,Address,Address2,City,State,PostalCode,Country
// REH,1,"2001 Creaking Oak Drive","","Springfield","MO","65803","USA"

// CreateBuilding reads a rental specialty type string array and creates a database record for the rental specialty type.
func CreateBuilding(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateBuilding"
	var (
		err error
		b   rlib.Building
	)

	const (
		BUD        = 0
		BldgNo     = iota
		Address    = iota
		Address2   = iota
		City       = iota
		State      = iota
		PostalCode = iota
		Country    = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"BldgNo", BldgNo},
		{"Address", Address},
		{"Address2", Address2},
		{"City", City},
		{"State", State},
		{"PostalCode", PostalCode},
		{"Country", Country},
	}

	y, err := ValidateCSVColumnsErr(csvCols, sa, funcname, lineno)
	if y {
		return 1, err
	}
	if lineno == 1 {
		return 0, nil // we've validated the col headings, all is good, send the next line
	}

	des := strings.ToLower(strings.TrimSpace(sa[0]))

	//-------------------------------------------------------------------
	// Make sure the rlib.Business is in the database
	//-------------------------------------------------------------------
	if len(des) > 0 {
		// TODO(Steve): ignore error?
		b1, _ := rlib.GetBusinessByDesignation(ctx, des)
		if len(b1.Designation) == 0 {
			errMsg := fmt.Sprintf("rlib.Business with designation %s does not exist", des)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		b.BID = b1.BID
	}

	//-------------------------------------------------------------------
	// parse out the rlib.Building number
	//-------------------------------------------------------------------
	if len(sa[1]) > 0 {
		i, err := strconv.Atoi(sa[1])
		if err != nil || i < 0 {
			errMsg := fmt.Sprintf("invalid rlib.Building number: %s", sa[BldgNo])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BldgNo, -1, errMsg)
		}
		b.BLDGID = int64(i)
	}

	b.Address = strings.TrimSpace(sa[2])
	b.Address2 = strings.TrimSpace(sa[3])
	b.City = strings.TrimSpace(sa[4])
	b.State = strings.TrimSpace(sa[5])
	b.PostalCode = strings.TrimSpace(sa[6])
	b.Country = strings.TrimSpace(sa[7])

	//-------------------------------------------------------------------
	// OK, just insert the record and we're done
	//-------------------------------------------------------------------
	_, err = rlib.InsertBuildingWithID(ctx, &b)
	if nil != err {
		errMsg := fmt.Sprintf("error inserting rlib.Building = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	return 0, nil
}

// LoadBuildingCSV loads a csv file with rental specialty types and processes each one
func LoadBuildingCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateBuilding)
}
