package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strings"
)

//  0     1            2        3
//  BUD,  ElementType, ID,      CID
//  REX,  5 ,          123,     456

// CreateCustomAttributeRefs reads a rlib.CustomAttributeRefs string array and creates a database record
func CreateCustomAttributeRefs(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "Createrlib.CustomAttributeRefs"
	var (
		err error
		c   rlib.CustomAttributeRef
	)

	const (
		BUD         = 0
		ElementType = iota
		ID          = iota
		CID         = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"ElementType", ElementType},
		{"ID", ID},
		{"CID", CID},
	}

	y, err := ValidateCSVColumnsErr(csvCols, sa, funcname, lineno)
	if y {
		return 1, err
	}
	if lineno == 1 {
		return 0, nil // we've validated the col headings, all is good, send the next line
	}

	//-------------------------------------------------------------------
	// BUD
	//-------------------------------------------------------------------
	cmpdes := strings.TrimSpace(sa[BUD])
	if len(cmpdes) > 0 {
		b2, err := rlib.GetBusinessByDesignation(ctx, cmpdes)
		if err != nil {
			errMsg := fmt.Sprintf("could not find Business named %s", cmpdes)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		if b2.BID == 0 {
			errMsg := fmt.Sprintf("could not find Business named %s", cmpdes)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		c.BID = b2.BID
	}

	c.ElementType, err = rlib.IntFromString(sa[ElementType], "ElementType is invalid")
	if err != nil {
		errMsg := fmt.Sprintf(err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, ElementType, -1, errMsg)
	}
	if c.ElementType < rlib.ELEMRENTABLETYPE || c.ElementType > rlib.ELEMLAST {
		errMsg := fmt.Sprintf("ElementType value must be a number from %d to %d", rlib.ELEMRENTABLETYPE, rlib.ELEMLAST)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, ElementType, -1, errMsg)
	}

	c.ID, err = rlib.IntFromString(sa[ID], "ID value cannot be converted to an integer")
	if err != nil {
		errMsg := fmt.Sprintf(err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, ID, -1, errMsg)
	}
	c.CID, err = rlib.IntFromString(sa[CID], "CID value cannot be converted to an integer")
	if err != nil {
		errMsg := fmt.Sprintf(err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, CID, -1, errMsg)
	}

	switch c.ElementType {
	case rlib.ELEMRENTABLETYPE:
		var rt rlib.RentableType
		err := rlib.GetRentableType(ctx, c.ID, &rt)
		if err != nil {
			errMsg := fmt.Sprintf("Could not load rlib.RentableType with id %d:  error = %v", c.ID, err)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, ID, -1, errMsg)
		}
	}

	// TODO(Steve): ignore error?
	ref, _ := rlib.GetCustomAttributeRef(ctx, c.ElementType, c.ID, c.CID)
	if ref.ElementType == c.ElementType && ref.CID == c.CID && ref.ID == c.ID {
		errMsg := fmt.Sprintf("This reference already exists, no changes made")
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	_, err = rlib.InsertCustomAttributeRef(ctx, &c)
	if err != nil {
		errMsg := fmt.Sprintf("Could not insert CustomAttributeRef. err = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	return 0, nil
}

// LoadCustomAttributeRefsCSV loads a csv file with a chart of accounts and creates rlib.GLAccount markers for each
func LoadCustomAttributeRefsCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateCustomAttributeRefs)
}
