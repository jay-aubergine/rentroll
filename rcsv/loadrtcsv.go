package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strconv"
	"strings"
)

// GetBusinessBID returns the BID for the rlib.Business with the supplied designation
func GetBusinessBID(ctx context.Context, des string) (int64, error) {
	//-------------------------------------------------------------------
	// Make sure the rlib.Business exists...
	//-------------------------------------------------------------------
	b, err := rlib.GetBusinessByDesignation(ctx, des)
	return b.BID, err
}

// CreateRentableType reads an rlib.Rentable type string array and creates a database record for the rlib.Rentable type
//
//                                                                               Repeat as many 3-tuples as needed
//                                                                                   /----------^-------------\
//  [0]        [1]      [2]   			[3]            [4]     5      6              7          8       9
// Designation,Style,	Name, 			RentCycle,  Proration, GSRPC, ManageToBudget,MarketRate,DtStart,DtStop
// REH,        "GM",	"Geezer Miser", 6,		       4,      4,     1,             1100.00,   1/1/2015, 1/1/2017
// REH,        "FS",	"Flat Studio",  6,		       4,      4,     1,             1500.00,   1/1/2015, 1/1/2017
// REH,        "SBL",	"SB Loft",     	6,		       4,      4,     1,             1750.00,   1/1/2015, 1/1/2017
// REH,        "KDS",	"KD Suite",    	6,		       4,      4,     1,             2000.00,   1/1/2015, 1/1/2017
// REH,        "VEH",	Vehicle,       	3,		       0,      4,     1,             10.0,   1/1/2015, 1/1/2017
// REH,        "CPT",	Carport,       	6,		       4,      4,     1,             35.0,   1/1/2015, 1/1/2017
func CreateRentableType(ctx context.Context, sa []string, lineno int) (int, error) {
	funcname := "CreateRentableType"
	const (
		BUD            = 0
		Style          = iota
		Name           = iota
		RentCycle      = iota
		Proration      = iota
		GSRPC          = iota
		ManageToBudget = iota
		MarketRate     = iota
		DtStart        = iota
		DtStop         = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"Style", Style},
		{"Name", Name},
		{"RentCycle", RentCycle},
		{"Proration", Proration},
		{"GSRPC", GSRPC},
		{"ManageToBudget", ManageToBudget},
		{"MarketRate", MarketRate},
		{"DtStart", DtStart},
		{"DtStop", DtStop},
	}

	y, err := ValidateCSVColumnsErr(csvCols, sa, funcname, lineno)
	if y {
		return 1, err
	}
	if lineno == 1 {
		return 0, nil // we've validated the col headings, all is good, send the next line
	}

	//-------------------------------------------------------------------
	// Check to see if this rlib.Rentable type is already in the database
	//-------------------------------------------------------------------
	des := strings.TrimSpace(sa[0])
	var a rlib.RentableType
	bid, err := GetBusinessBID(ctx, des)
	if err != nil {
		errMsg := fmt.Sprintf("error while getting business by designation(%s): %s", sa[BUD], err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
	}
	if bid == 0 {
		errMsg := fmt.Sprintf("rlib.Business named %s not found", sa[BUD])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
	}

	a.BID = bid
	a.Style = strings.TrimSpace(sa[1])
	if len(a.Style) > 0 {
		rt, err := rlib.GetRentableTypeByStyle(ctx, a.Style, bid)
		if err != nil {
			errMsg := fmt.Sprintf("err = %v", err)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Style, -1, errMsg)
		}
		if rt.RTID > 0 {
			errMsg := fmt.Sprintf("%s:: rlib.RentableType named %s already exists", DupRentableType, a.Style)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Style, -1, errMsg)
		}
	}

	a.Name = strings.TrimSpace(sa[2])

	//-------------------------------------------------------------------
	// Load the values based on csv input
	//-------------------------------------------------------------------
	n, err := strconv.Atoi(strings.TrimSpace(sa[3])) // frequency
	if err != nil || !rlib.IsValidAccrual(int64(n)) {
		errMsg := fmt.Sprintf("Invalid rental frequency: %s", sa[RentCycle])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentCycle, -1, errMsg)
	}
	a.RentCycle = int64(n)

	n, err = strconv.Atoi(strings.TrimSpace(sa[4])) // Proration
	if err != nil || !rlib.IsValidAccrual(int64(n)) {
		errMsg := fmt.Sprintf("Invalid rental proration frequency: %s", sa[Proration])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Proration, -1, errMsg)
	}
	a.Proration = int64(n)
	if a.Proration > a.RentCycle {
		errMsg := fmt.Sprintf("Proration frequency (%d) must be greater than rental frequency (%d)", a.Proration, a.RentCycle)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	n, err = strconv.Atoi(strings.TrimSpace(sa[5])) // GSRPC
	if err != nil || !rlib.IsValidAccrual(int64(n)) {
		errMsg := fmt.Sprintf("Invalid rental GSRPC: %s", sa[GSRPC])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, GSRPC, -1, errMsg)
	}
	a.GSRPC = int64(n)

	n64, err := rlib.YesNoToInt(strings.TrimSpace(sa[6])) // manage to budget
	if err != nil {
		errMsg := fmt.Sprintf("Invalid manage to budget flag: %s", sa[ManageToBudget])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, ManageToBudget, -1, errMsg)
	}
	a.ManageToBudget = int64(n64)

	rtid, err := rlib.InsertRentableType(ctx, &a)
	if err != nil {
		errMsg := fmt.Sprintf("Error inserting Rentable Type: %s", err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	// rlib.Rentable Market Rates are provided in 3-tuples starting at index 7 - Amount,startdata,enddate
	// rlib.Console("LoadRTcsv: rtid = %d\n", rtid)
	if rtid > 0 {
		// rlib.Console("LoadRTcsv: preparing to parse market rate.  len(sa) = %d\n", len(sa))
		for i := MarketRate; i < len(sa); i += 3 {
			if len(sa[i]) == 0 { // this will happen when programs like excel save the csv file
				continue
			}
			var x float64
			var err error
			var m rlib.RentableMarketRate
			m.RTID = rtid
			if x, err = strconv.ParseFloat(strings.TrimSpace(sa[i]), 64); err != nil {
				errMsg := fmt.Sprintf("Invalid floating point number: %s   err = %s", sa[i], err.Error())
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, i, -1, errMsg)
			}
			m.MarketRate = x
			DtStart, err := rlib.StringToDate(sa[i+1])
			if err != nil {
				errMsg := fmt.Sprintf("invalid start date:  %s", sa[i+1])
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, i+1, -1, errMsg)
			}
			m.DtStart = DtStart
			DtStop, err := rlib.StringToDate(sa[i+2])
			if err != nil {
				errMsg := fmt.Sprintf("invalid stop date:  %s", sa[i+2])
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, i+2, -1, errMsg)
			}
			m.DtStop = DtStop
			if m.DtStart.After(m.DtStop) {
				errMsg := fmt.Sprintf("Stop date (%s) must be after Start date (%s)", m.DtStop, m.DtStart)
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
			}
			m.BID = a.BID
			_, err = rlib.InsertRentableMarketRates(ctx, &m)
			if err != nil {
				errMsg := fmt.Sprintf("error saving RentableMarketRate:  %s", err.Error())
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
			}
		}
	}
	return 0, nil
}

// LoadRentableTypesCSV loads a csv file with rlib.Rentable types and processes each one
func LoadRentableTypesCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateRentableType)
}
