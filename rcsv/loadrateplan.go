package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strings"
)

//      RatePlan
// |<------|--------------->|
// 0    1        2
// BUD, Name,  Exports
// REX, X1,    GDS,Sabre
// REX, X2,

// CreateRatePlans reads a RatePlan string array and creates a database record
func CreateRatePlans(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateRatePlans"

	var (
		err   error
		rp    rlib.RatePlan
		FLAGS uint64
	)

	const (
		BUD     = 0
		Name    = iota
		Exports = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"Name", Name},
		{"Exports", Exports},
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
		rp.BID = b1.BID
	}
	rp.Name = strings.TrimSpace(sa[1])
	if len(rp.Name) == 0 {
		errMsg := fmt.Sprintf("No Name found for the RatePlan")
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Name, -1, errMsg)
	}
	// need to check for another RatePlan of the same name

	//-------------------------------------------------------------------
	// Exports
	//-------------------------------------------------------------------
	ss := strings.TrimSpace(sa[2])
	if len(ss) > 0 {
		ssa := strings.Split(ss, ",")
		for i := 0; i < len(ssa); i++ {
			switch ssa[i] {
			case "GDS":
				FLAGS |= rlib.FlRatePlanGDS
			case "Sabre":
				FLAGS |= rlib.FlRatePlanSabre
			default:
				errMsg := fmt.Sprintf("Unrecognized export flag: %s", ssa[i])
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Exports, i, errMsg)
			}
		}
	}

	//return CsvErrorSensitivity, fmt.Errorf("FLAGS = 0x%x", FLAGS)

	rpid, err := rlib.InsertRatePlan(ctx, &rp)
	if err != nil {
		errMsg := fmt.Sprintf("Error inserting RatePlan.  err = %s", err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	// Now add the FLAGS as a custom attribute to the RatePlan
	var c rlib.CustomAttribute     // This is the custom attribute
	var cr rlib.CustomAttributeRef // This is the reference that binds it to an object
	c.Name = "FLAGS"
	c.BID = rp.BID
	c.Type = rlib.CUSTUINT
	c.Value = fmt.Sprintf("%d", FLAGS)
	cid, err := rlib.InsertCustomAttribute(ctx, &c)
	if err != nil {
		errMsg := fmt.Sprintf("Could not insert CustomAttribute. err = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	cr.ElementType = rlib.ELEMRATEPLAN
	cr.ID = rpid
	cr.BID = rp.BID
	cr.CID = cid
	_, err = rlib.InsertCustomAttributeRef(ctx, &cr)
	if err != nil {
		errMsg := fmt.Sprintf("Could not insert CustomAttributeRef. err = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	return 0, nil
}

// LoadRatePlansCSV loads a csv file with note types
func LoadRatePlansCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateRatePlans)
}
