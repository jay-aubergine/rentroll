package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strings"
)

//  CSV file format:
//                        RT name or style  string with or without %
// 0    1             2   3                 4
// BUD, RPName, RPRID     RentableType,     Amount
// REX, FAA-P,  RPR0001,  GM,               85%
// REX, FAA-P,  1,        Flat Studio,      1400
// REX, FAA-P,  1,        SBL,    			1500
// REX, FAA-P,  1,        KDS,    			75%
// REX, FAA-T,  1,        GM,               90%
// REX, FAA-T,  1,        Flat Studio,      90%
// REX, FAA-T,  1,        SBL,    			1500
// REX, FAA-T,  1,        KDS,    			80%

// CreateRatePlanRefRTRate reads a rental specialty type string array and creates a database record for the rental specialty type.
func CreateRatePlanRefRTRate(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateRatePlanRefRTRate"
	var (
		b   rlib.Business
		err error
	)

	const (
		BUD          = 0
		RPName       = iota
		RPRID        = iota
		RentableType = iota
		Amount       = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"RPName", RPName},
		{"RPRID", RPRID},
		{"RentableType", RentableType},
		{"Amount", Amount},
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
	// BUD
	//-------------------------------------------------------------------
	if len(des) > 0 {
		b, err = rlib.GetBusinessByDesignation(ctx, des)
		if err != nil {
			errMsg := fmt.Sprintf("error while getting business by designation(%s): %s", sa[BUD], err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		if len(b.Designation) == 0 {
			errMsg := fmt.Sprintf("rlib.Business with designation %s does not exist", sa[BUD])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
	}

	//-------------------------------------------------------------------
	// RatePlan Name
	//-------------------------------------------------------------------
	var rp rlib.RatePlan
	rpname := strings.ToLower(strings.TrimSpace(sa[RPName]))
	if len(rpname) > 0 {
		err = rlib.GetRatePlanByName(ctx, b.BID, rpname, &rp)
		if err != nil {
			errMsg := fmt.Sprintf("error getting RatePlan named %s not found: %s", sa[RPName], err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RPName, -1, errMsg)
		}
		if rp.RPID < 1 {
			errMsg := fmt.Sprintf("RatePlan named %s not found", sa[RPName])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RPName, -1, errMsg)
		}
	}

	var a rlib.RatePlanRefRTRate
	var errmsg string

	a.BID = b.BID

	//-------------------------------------------------------------------
	// RPRef
	//-------------------------------------------------------------------
	a.RPRID = CSVLoaderGetRPRID(strings.TrimSpace(sa[RPRID]))
	if 0 == a.RPRID {
		errMsg := fmt.Sprintf("Bad value for RatePlanRef ID: %s", sa[RPRID])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RPRID, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// RT Style
	// identifies the RentableType
	//-------------------------------------------------------------------
	name := strings.TrimSpace(sa[RentableType])
	rt, err := rlib.GetRentableTypeByStyle(ctx, name, b.BID)
	if err != nil {
		errMsg := fmt.Sprintf("could not load RentableType with Style = %s,  err:  %s", sa[RentableType], err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableType, -1, errMsg)
	}
	a.RTID = rt.RTID

	//-------------------------------------------------------------------
	// Amount
	// Entered as a string. If the string contains a % then the amount
	// is a percentage and we set the % flag. Otherwise, it is an absolute amount
	//-------------------------------------------------------------------
	amt := strings.TrimSpace(sa[Amount])
	a.Val, errmsg = rlib.FloatFromString(amt, "bad amount")
	if len(errmsg) > 0 {
		errMsg := errmsg
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Amount, -1, errMsg)
	}
	if strings.Contains(amt, "%") {
		a.FLAGS |= rlib.FlRTRpct
	}

	//-------------------------------------------------------------------
	// Insert the record
	//-------------------------------------------------------------------
	_, err = rlib.InsertRatePlanRefRTRate(ctx, &a)
	if nil != err {
		errMsg := fmt.Sprintf("error inserting RatePlanRefRTRate = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	return 0, nil
}

// LoadRatePlanRefRTRatesCSV loads a csv file with RatePlan rates for specific rentable types
func LoadRatePlanRefRTRatesCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateRatePlanRefRTRate)

}
