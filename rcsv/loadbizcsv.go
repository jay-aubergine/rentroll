package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strconv"
	"strings"
)

// 0           1    2                3,                    4,
// Bud,Name,DefaultRentCycle,DefaultProrationCycle,DefaultGSRPC
// REH,,4,0
// BBBB,Big Bob's Barrel Barn,4,0

// GetAccrual sets the DefaultRentCycle attribute of the rlib.Business structure based on the provided string s
func GetAccrual(s string) (int64, bool) {
	if len(s) > 0 {
		i, err := strconv.Atoi(s)
		if err == nil && rlib.IsValidAccrual(int64(i)) {
			return int64(i), true
		}
	}
	return int64(0), false
}

// CreatePhonebookLinkedBusiness creates a new rlib.Business in the
// RentRoll database from the company in Phonebook with the supplied designation
func CreatePhonebookLinkedBusiness(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreatePhonebookLinkedBusiness"
	var (
		err   error
		b     rlib.Business
		des   = strings.TrimSpace(sa[0])
		found = true
		ok    bool
	)

	const (
		BUD                   = 0
		Name                  = iota
		DefaultRentCycle      = iota
		DefaultProrationCycle = iota
		DefaultGSRPC          = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"Name", Name},
		{"DefaultRentCycle", DefaultRentCycle},
		{"DefaultProrationCycle", DefaultProrationCycle},
		{"DefaultGSRPC", DefaultGSRPC},
	}

	y, err := ValidateCSVColumnsErr(csvCols, sa, funcname, lineno)
	if y {
		return 1, err
	}
	if lineno == 1 {
		return 0, nil // we've validated the col headings, all is good, send the next line
	}

	//-------------------------------------------------------------------
	// Check to see if this rlib.Business is already in the database
	//-------------------------------------------------------------------
	if len(des) > 0 {
		// TODO(Steve): ignore error?
		b1, _ := rlib.GetBusinessByDesignation(ctx, des)
		if len(b1.Designation) > 0 {
			errMsg := fmt.Sprintf("rs, rlib.Business Unit with designation %s already exists", des)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		found = false
	}

	//-------------------------------------------------------------------
	// It does not exist, see if we can find it in Phonebook...
	//-------------------------------------------------------------------
	if !found && len(des) > 0 {
		// TODO(Steve): ignore error?
		bu, _ := rlib.GetBusinessUnitByDesignation(ctx, des)
		if len(bu.Description) > 0 {
			found = true
			b.Name = bu.Name    // Phonebook rlib.Business Unit name
			b.Designation = des // rlib.Business unit designator
		}
	}

	//-----------------------------------------
	// DefaultRentCycle
	//-----------------------------------------
	if b.DefaultRentCycle, ok = GetAccrual(strings.TrimSpace(sa[2])); !ok {
		errMsg := fmt.Sprintf("Invalid Rent Cycle: %s", sa[DefaultRentCycle])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, DefaultRentCycle, -1, errMsg)
	}

	//-----------------------------------------
	// DefaultProrationCycle
	//-----------------------------------------
	if b.DefaultProrationCycle, ok = GetAccrual(strings.TrimSpace(sa[3])); !ok {
		errMsg := fmt.Sprintf("Invalid Proration Cycle: %s", sa[DefaultProrationCycle])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, DefaultProrationCycle, -1, errMsg)
	}

	//-----------------------------------------
	// DefaultGSRPC
	//-----------------------------------------
	if b.DefaultGSRPC, ok = GetAccrual(strings.TrimSpace(sa[4])); !ok {
		errMsg := fmt.Sprintf("Invalid GSRPC: %s", sa[DefaultGSRPC])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, DefaultGSRPC, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// If we did not find it in Phonebook, we still need to create it,
	// so use the fields we have...
	//-------------------------------------------------------------------
	if !found {
		b.Name = strings.TrimSpace(sa[1])
		b.Designation = des
	}
	// fmt.Printf("Business to save:  %#v\n", b)
	_, err = rlib.InsertBusiness(ctx, &b)
	if err != nil {
		errMsg := fmt.Sprintf("error inserting rlib.Business = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	// err = rlib.NewBusinessInit(bid)
	// if err != nil {
	// 	return CsvErrorSensitivity, fmt.Errorf("%s: error from NewBusinessInit = %v", funcname, err)
	// }
	return 0, nil
}

// LoadBusinessCSV loads the values from the supplied csv file and creates rlib.Business records
// as needed.
func LoadBusinessCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreatePhonebookLinkedBusiness)
}
