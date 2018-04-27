package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strings"
)

//  CSV file format:
//
// 0    1       2        3            4
// BUD, RPName,          RPRID        Specialty,   Amount,  Specialty2, Amount2, ...
// REX, FAA-P,  RPR0001, GM,          Lake View,   85%,     Fireplace,      90%
// REX, FAA-P,  1,       Flat Studio, Lake View,   100%,    Fireplace,
// REX, FAA-P,  1,       SBL,    	  Lake View,   10.25,   Fireplace,
// REX, FAA-P,  1,       KDS,    	  Lake View,   75%,     Fireplace,
// REX, FAA-T,  2,       GM,          Lake View,   90%,     Fireplace,
// REX, FAA-T,  2,       Flat Studio, Lake View,   90%,     Fireplace,
// REX, FAA-T,  2,       SBL,    	  Lake View,   11.50,   Fireplace,
// REX, FAA-T,  2,       KDS,    	  Lake View,   87%,     Fireplace,

// CreateRatePlanRefSPRate reads a rental specialty type string array and creates a database record for the rental specialty type.
func CreateRatePlanRefSPRate(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateRatePlanRefSPRate"
	var (
		err error
		b   rlib.Business
	)

	const (
		BUD          = 0
		RPName       = iota
		RPRID        = iota
		RentableType = iota
		Amount       = iota
	)

	required := 5
	if len(sa) < required {
		errMsg := fmt.Sprintf("found %d values, there must be at least %d", len(sa), required)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// BUD
	//-------------------------------------------------------------------
	des := strings.ToLower(strings.TrimSpace(sa[0]))
	if des == "bud" {
		return 0, nil // this is just the column heading
	}
	if len(des) > 0 {
		b, err = rlib.GetBusinessByDesignation(ctx, des)
		if err != nil {
			errMsg := fmt.Sprintf("error while getting business by designation(%s): %s", sa[BUD], err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		if len(b.Designation) == 0 {
			errMsg := fmt.Sprintf("Business with designation %s does not exist", sa[BUD])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
	}

	// knowing the Business we can get all the specialties and rentable types. The easy way is just to load an XBiz
	var xbiz rlib.XBusiness
	err = rlib.GetXBusiness(ctx, b.BID, &xbiz)
	if err != nil {
		errMsg := fmt.Sprintf("error while getting business BID(%d): %s", b.BID, err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// RatePlan Name
	//-------------------------------------------------------------------
	var rp rlib.RatePlan
	rpname := strings.ToLower(strings.TrimSpace(sa[1]))
	if len(rpname) > 0 {
		err = rlib.GetRatePlanByName(ctx, b.BID, rpname, &rp)
		if err != nil {
			errMsg := fmt.Sprintf("error getting RatePlan name %s: %s", sa[RPName], err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RPName, -1, errMsg)
		}
		if rp.RPID < 1 {
			errMsg := fmt.Sprintf("RatePlan named %s not found", sa[RPName])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RPName, -1, errMsg)
		}
	}

	var (
		a      rlib.RatePlanRefSPRate
		errmsg string
	)

	a.BID = b.BID

	//-------------------------------------------------------------------
	// RPRef
	//-------------------------------------------------------------------
	a.RPRID = CSVLoaderGetRPRID(strings.TrimSpace(sa[2]))
	if 0 == a.RPRID {
		errMsg := fmt.Sprintf("Bad value for RatePlanRef ID: %s", sa[RPRID])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RPRID, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// Rentable Type
	//-------------------------------------------------------------------
	rtname := strings.TrimSpace(sa[3])
	found := false
	for k, v := range xbiz.RT { // Make sure it's something we recognize...
		if v.Name == rtname || v.Style == rtname {
			found = true
			a.RTID = k // mark the RTID
			break
		}
	}
	if !found {
		errMsg := fmt.Sprintf("could not find Specialty with name = %s", sa[RentableType])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableType, -1, errMsg)
	}

	for i := 4; i < len(sa); i += 2 {
		p := a // start a new structure.  We just need to fill out the RSPID, Amount, and FLAG

		//-------------------------------------------------------------------
		// Specialty
		//-------------------------------------------------------------------
		name := strings.TrimSpace(sa[i])
		if len(name) == 0 { // if the specialty name is blank...
			continue // ... then ignore
		}
		// Make sure it's something we recognize...
		found = false
		for k, v := range xbiz.US {
			if v.Name == name {
				found = true
				p.RSPID = k
				break
			}
		}
		if !found {
			errMsg := fmt.Sprintf("could not find Specialty with name = %s", name)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, i, -1, errMsg)
		}

		//-------------------------------------------------------------------
		// Amount
		// Entered as a string. If the string contains a % then the amount
		// is a percentage and we set the % flag. Otherwise, it is an
		// absolute amount
		//-------------------------------------------------------------------
		amt := strings.TrimSpace(sa[i+1])
		p.Val, errmsg = rlib.FloatFromString(amt, "bad amount")
		if len(errmsg) > 0 {
			errMsg := errmsg
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, i+1, -1, errMsg)
		}
		if strings.Contains(amt, "%") {
			p.FLAGS |= rlib.FlSPRpct // mark it as a percentage
		}

		//-------------------------------------------------------------------
		// Insert the record
		//-------------------------------------------------------------------
		_, err = rlib.InsertRatePlanRefSPRate(ctx, &p)
		if nil != err {
			errMsg := fmt.Sprintf("error inserting RatePlanRefSPRate = %v", err)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
	}
	return 0, nil
}

// LoadRatePlanRefSPRatesCSV loads a csv file with RatePlan rates for specific rentable types
func LoadRatePlanRefSPRatesCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateRatePlanRefSPRate)
}
