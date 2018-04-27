package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strings"
)

// 0    1
// BUD, Name
// REX,Payment
// REX,Deposit

// CreateNoteTypes reads a CustomAttributes string array and creates a database record
func CreateNoteTypes(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateNoteTypes"
	var (
		err error
		nt  rlib.NoteType
	)

	const (
		BUD  = 0
		Name = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"Name", Name},
	}

	y, err := ValidateCSVColumnsErr(csvCols, sa, funcname, lineno)
	if y {
		return 1, err
	}
	if lineno == 1 {
		return 0, nil // we've validated the col headings, all is good, send the next line
	}

	//-------------------------------------------------------------------
	// Make sure the rlib.Business is in the database
	//-------------------------------------------------------------------
	des := strings.ToLower(strings.TrimSpace(sa[BUD]))
	if len(des) > 0 {
		b1, err := rlib.GetBusinessByDesignation(ctx, des)
		if err != nil {
			errMsg := fmt.Sprintf("error while getting business by designation(%s), error: %s", sa[BUD], err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		if len(b1.Designation) == 0 {
			errMsg := fmt.Sprintf("rlib.Business with designation %s does not exist", sa[BUD])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		nt.BID = b1.BID
	}
	nt.Name = strings.TrimSpace(sa[1])
	if len(nt.Name) == 0 {
		errMsg := fmt.Sprintf("No Name found for the NoteType")
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Name, -1, errMsg)
	}
	_, err = rlib.InsertNoteType(ctx, &nt)
	if err != nil {
		errMsg := fmt.Sprintf("Error inserting NoteType.  err = %s", err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	return 0, nil
}

// LoadNoteTypesCSV loads a csv file with note types
func LoadNoteTypesCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateNoteTypes)
}
