package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strings"
)

//  CSV file format:
// 0    1             2         3         4              5              6                   7                8          9
// BUD, RPName,       DtStart,  DtStop,   FeeAppliesAge, MaxNoFeeUsers, AdditionalUserFee,  CancellationFee, PromoCode, Flags
// REX, A1-Transient, 1/1/2016, 7/1/2016, 12,            2,             10.0,               25.00,                    , Hide
// REX, A1-LongTerm,  1/1/2016, 7/1/2016, 12,            2,             50.0,               25.00,                    , Hide
// REX, A2-Transient, 1/1/2016, 7/1/2016, 12,            2,             15.0,               25.00,                    ,
// REX, A2-LongTerm,  1/1/2016, 7/1/2016, 12,            2,             75.0,               25.00,                    ,

// CreateRatePlanRef reads a rental specialty type string array and creates a database record for the rental specialty type.
func CreateRatePlanRef(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateRatePlanRef"
	var (
		err error
		b   rlib.Business
	)

	const (
		BUD               = 0
		RPName            = iota
		DtStart           = iota
		DtStop            = iota
		FeeAppliesAge     = iota
		MaxNoFeeUsers     = iota
		AdditionalUserFee = iota
		CancellationFee   = iota
		PromoCode         = iota
		Flags             = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"RPName", RPName},
		{"DtStart", DtStart},
		{"DtStop", DtStop},
		{"FeeAppliesAge", FeeAppliesAge},
		{"MaxNoFeeUsers", MaxNoFeeUsers},
		{"AdditionalUserFee", AdditionalUserFee},
		{"CancellationFee", CancellationFee},
		{"PromoCode", PromoCode},
		{"Flags", Flags},
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
			errMsg := fmt.Sprintf("error while getting business by designation(%s), error: %s", sa[BUD], err.Error())
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
			errMsg := fmt.Sprintf("error while getting RatePlan name(%s): %s", sa[RPName], err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RPName, -1, errMsg)
		}
		if rp.RPID < 1 {
			errMsg := fmt.Sprintf("RatePlan named %s not found", sa[RPName])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RPName, -1, errMsg)
		}
	}

	var a rlib.RatePlanRef
	var errmsg string

	a.BID = b.BID

	//-------------------------------------------------------------------
	// DtStart
	//-------------------------------------------------------------------
	dt := sa[DtStart]
	a.DtStart, err = rlib.StringToDate(dt)
	if err != nil {
		errMsg := fmt.Sprintf("invalid start date:  %s", sa[DtStart])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, DtStart, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// DtStop
	//-------------------------------------------------------------------
	dt = sa[DtStop]
	a.DtStop, err = rlib.StringToDate(dt)
	if err != nil {
		errMsg := fmt.Sprintf("invalid stop date:  %s", sa[DtStop])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, DtStop, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// Fee Applies Age
	//-------------------------------------------------------------------
	a.FeeAppliesAge, err = rlib.IntFromString(sa[FeeAppliesAge], "Invalid FeeAppliesAge")
	if err != nil {
		errMsg := fmt.Sprintf("Invalid number: %s", err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, FeeAppliesAge, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// Max No Fee Users
	//-------------------------------------------------------------------
	a.MaxNoFeeUsers, err = rlib.IntFromString(sa[MaxNoFeeUsers], "Invalid MaxNoFeeUsers")
	if err != nil {
		errMsg := fmt.Sprintf("Invalid number: %s", err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, MaxNoFeeUsers, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// AdditionalUserFee
	//-------------------------------------------------------------------
	a.AdditionalUserFee, errmsg = rlib.FloatFromString(sa[AdditionalUserFee], "Invalid Additional User Fee")
	if len(errmsg) > 0 {
		errMsg := fmt.Sprintf("Invalid number: %s", err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, AdditionalUserFee, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// CancellationFee
	//-------------------------------------------------------------------
	a.CancellationFee, errmsg = rlib.FloatFromString(sa[CancellationFee], "Invalid Cancellation Fee")
	if len(errmsg) > 0 {
		errMsg := fmt.Sprintf("Invalid number: %s", err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, CancellationFee, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// PromoCode
	//-------------------------------------------------------------------
	a.PromoCode = strings.TrimSpace(sa[PromoCode])

	//-------------------------------------------------------------------
	// FLAGS
	//-------------------------------------------------------------------
	ss := strings.TrimSpace(sa[Flags])
	if len(ss) > 0 {
		ssa := strings.Split(ss, ",")
		for i := 0; i < len(ssa); i++ {
			switch strings.ToLower(ssa[i]) {
			case "hide":
				a.FLAGS |= rlib.FlRTRRefHide // do not show this rate plan to users
			default:
				errMsg := fmt.Sprintf("Unrecognized export flag: %s", ssa[i])
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Flags, i, errMsg)
			}
		}
	}

	//-------------------------------------------------------------------
	// Insert the record
	//-------------------------------------------------------------------
	a.RPID = rp.RPID
	_, err = rlib.InsertRatePlanRef(ctx, &a)
	if nil != err {
		errMsg := fmt.Sprintf("error inserting RatePlanRef = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	return 0, nil
}

// LoadRatePlanRefsCSV loads a csv file with rental specialty types and processes each one
func LoadRatePlanRefsCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateRatePlanRef)
}
