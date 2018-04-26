package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strconv"
	"strings"
	"time"
)

// CSV file format:
//   0  1     2               3                       4                                 5
//                            "usr1;usr2;..usrN"      "S1,Strt1,Stp1;S2,Strt2,Stp2...", “A2,1/10/16,6/1/16;B2,6/1/16,”
// BUD, Name, AssignmentTime, RentableUsers,          RentableStatus,                   RentableTypeRef
// REX, 101,  1,              "bill@x.com;sue@x.com"  "1,1/1/14,6/15/16;2,6/15/16,",    "A2,1/1/14,6/1/16;B2,6/1/16,"
// REX, 102,  1,                                      "1,1/1/14,6/15/16;2,6/15/16,",    "A2,1/1/14,6/1/16;B2,6/1/16,"
// REX, 103,  1,                                      "1,1/1/14,6/15/16;2,6/15/16,",    "A2,1/1/14,6/1/16;B2,6/1/16,"
// REX, 104,  1,                                      "1,1/1/14,6/15/16;2,6/15/16,",    "A2,1/1/14,6/1/16;B2,6/1/16,"
// REX, 105,  1,                                      "1,1/1/14,6/15/16;2,6/15/16,",    "A2,1/1/14,6/1/16;B2,6/1/16,"
// REX, 106,  1,                                      "1,1/1/14,6/15/16;2,6/15/16,",    "A2,1/1/14,6/1/16;B2,6/1/16,"

// readTwoDates assumes that a date string is in ss[1] and ss[2].  It will parse and return the dates
// along with any error it finds.
func readTwoDates(s1, s2 string) (time.Time, time.Time, error) {
	var DtStart, DtStop time.Time
	var err error
	DtStart, err = rlib.StringToDate(s1) // required field
	if err != nil {
		err = fmt.Errorf("invalid start date:  %s", s1)
		return DtStart, DtStop, err
	}

	end := "1/1/9999"
	if len(s2) > 0 { //optional field -- MAYBE, if not present assume year 9999
		if len(strings.TrimSpace(s2)) > 0 {
			end = s2
		}
	}
	DtStop, err = rlib.StringToDate(end)
	if err != nil {
		err = fmt.Errorf("invalid stop date:  %s", s2)
	}
	return DtStart, DtStop, err
}

// CreateRentables reads a rental specialty type string array and creates a database record for the rental specialty type.
func CreateRentables(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateRentables"
	var (
		err error
		r   rlib.Rentable
	)

	const (
		BUD             = 0
		Name            = iota
		AssignmentTime  = iota
		RUserSpec       = iota
		RentableStatus  = iota
		RentableTypeRef = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"Name", Name},
		{"AssignmentTime", AssignmentTime},
		{"RUserSpec", RUserSpec},
		{"RentableStatus", RentableStatus},
		{"RentableTypeRef", RentableTypeRef},
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
			errMsg := fmt.Sprintf("error while getting business by designation(%s): %s", des, err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		if len(b1.Designation) == 0 {
			errMsg := fmt.Sprintf("Business with bud %s does not exist", des)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		r.BID = b1.BID
	}

	//-------------------------------------------------------------------
	// The name must be unique. Make sure we don't have any other rlib.Rentable
	// with this name...
	//-------------------------------------------------------------------
	r.RentableName = strings.TrimSpace(sa[Name])
	r1, err := rlib.GetRentableByName(ctx, r.RentableName, r.BID)
	if err != nil {
		s := err.Error()
		if !strings.Contains(s, "no rows") {
			errMsg := fmt.Sprintf("error with rlib.GetRentableByName: %s", err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Name, -1, errMsg)
		}
	}
	if r1.RID > 0 {
		errMsg := fmt.Sprintf("%s:: Rentable with name \"%s\" already exists. Skipping.", DupRentable, r.RentableName)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Name, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// parse out the AssignmentTime value
	// Unknown = 0, Pre-assign = 1, assign at occupy commencement = 2
	//-------------------------------------------------------------------
	if len(sa[AssignmentTime]) > 0 {
		i, err := strconv.Atoi(sa[AssignmentTime])
		if err != nil || i < 0 || i > 2 {
			errMsg := fmt.Sprintf("invalid AssignmentTime number: %s", sa[AssignmentTime])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, AssignmentTime, -1, errMsg)
		}
		r.AssignmentTime = int64(i)
	}

	//-----------------------------------------------------------------------------------
	// USER 3-TUPLEs
	// "user1,dtstart1,dtstop1;user2,dtstart2,dtstop2;..."
	// example:
	// "ednak@springfield.com,1/1/2013,11/9/2015;homerj@springfield.com,11/20/2015,;marge@springfield.com,11/20/2015,"
	//-----------------------------------------------------------------------------------
	var rul []rlib.RentableUser // keep every rlib.RentableUser we find in an array
	if 0 < len(strings.TrimSpace(sa[RUserSpec])) {
		st := strings.Split(sa[RUserSpec], ";") // split it on Status 3-tuple separator (;)
		for i := 0; i < len(st); i++ {          //spin through the 3-tuples
			ss := strings.Split(st[i], ",")
			if len(ss) != 3 {
				errMsg := fmt.Sprintf("invalid User Specification. Each semi-colon separated field must have 3 values. Found %d in \"%s\"", len(ss), ss)
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RUserSpec, i, errMsg)
			}

			var ru rlib.RentableUser // struct for the data in this 3-tuple
			name := strings.TrimSpace(ss[0])
			n, err := CSVLoaderTransactantList(ctx, r.BID, name)
			if err != nil {
				errMsg := fmt.Sprintf("Error Loading transactant list: %s", err.Error())
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RUserSpec, i, errMsg)
			}
			if len(n) == 0 || n[0].TCID == 0 {
				errMsg := fmt.Sprintf("could not find Transactant with contact information %s", name)
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RUserSpec, i, errMsg)
			}
			ru.TCID = n[0].TCID

			ru.DtStart, ru.DtStop, err = readTwoDates(ss[1], ss[2])
			if err != nil {
				errMsg := err.Error()
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RUserSpec, i, errMsg)
			}
			rul = append(rul, ru) // add this struct to the list
		}
	}

	//-----------------------------------------------------------------------------------
	// STATUS 3-TUPLEs
	// "S1,Strt1,Stp1;S2,Strt2,Stp2 ..."
	//-----------------------------------------------------------------------------------
	if 0 == len(strings.TrimSpace(sa[RentableStatus])) {
		errMsg := fmt.Sprintf("rlib.RentableStatus value is required")
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableStatus, -1, errMsg)

	}
	var m []rlib.RentableStatus                  // keep every rlib.RentableStatus we find in an array
	st := strings.Split(sa[RentableStatus], ";") // split it on Status 3-tuple separator (;)
	for i := 0; i < len(st); i++ {               //spin through the 3-tuples
		ss := strings.Split(st[i], ",")
		if len(ss) != 3 {
			errMsg := fmt.Sprintf("invalid Rentable Status. Each semi-colon separated field must have 3 values. Found %d in \"%s\"", len(ss), ss)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableStatus, i, errMsg)
		}

		var rst rlib.RentableStatus // struct for the data in this 3-tuple
		ix, err := strconv.Atoi(ss[0])
		if err != nil || ix < rlib.USESTATUSunknown || ix > rlib.USESTATUSLAST {
			errMsg := fmt.Sprintf("invalid Status value: %s.  Must be in the range %d to %d", ss[0], rlib.USESTATUSunknown, rlib.USESTATUSLAST)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableStatus, i, errMsg)
		}
		rst.UseStatus = int64(ix)

		rst.DtStart, rst.DtStop, err = readTwoDates(ss[1], ss[2])
		if err != nil {
			errMsg := err.Error()
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableStatus, i, errMsg)
		}
		m = append(m, rst) // add this struct to the list
	}
	if len(m) == 0 {
		errMsg := fmt.Sprintf("RentableStatus value is required")
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableStatus, -1, errMsg)
	}

	//-----------------------------------------------------------------------------------
	// RTID 3-TUPLEs
	// "RTname1,Amount,startDate1,stopDate1;RTname2,startDate2,stopDate2;..."
	//-----------------------------------------------------------------------------------
	if 0 == len(strings.TrimSpace(sa[RentableTypeRef])) {
		errMsg := fmt.Sprintf("Rentable RTID Ref value is required")
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableTypeRef, -1, errMsg)
	}
	var n []rlib.RentableTypeRef
	st = strings.Split(sa[RentableTypeRef], ";") // split on RTID 3-tuple separator (;)
	for i := 0; i < len(st); i++ {               // spin through the 3-tuples
		ss := strings.Split(st[i], ",") // separate the 3 parts
		if len(ss) != 3 {
			errMsg := fmt.Sprintf("invalid RTID syntax. Each semi-colon separated field must have 3 values. Found %d in \"%s\"", len(ss), ss)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableTypeRef, i, errMsg)
		}

		var rt rlib.RentableTypeRef                                                       // struct for the data in this 3-tuple
		rstruct, err := rlib.GetRentableTypeByStyle(ctx, strings.TrimSpace(ss[0]), r.BID) // find the rlib.RentableType being referenced
		if err != nil {
			errMsg := fmt.Sprintf("Could not load rentable type with style name: %s  -- error = %s", ss[0], err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableTypeRef, i, errMsg)
		}
		if rstruct.RTID == 0 {
			errMsg := fmt.Sprintf("Could not load rentable type with style name: %s", ss[0])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableTypeRef, i, errMsg)
		}
		rt.RTID = rstruct.RTID

		rt.DtStart, rt.DtStop, err = readTwoDates(ss[1], ss[2])
		if err != nil {
			errMsg := err.Error()
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableTypeRef, i, errMsg)
		}
		n = append(n, rt) // add this struct to the list
	}

	//-------------------------------------------------------------------
	// OK, just insert the record and its sub-records and we're done
	//-------------------------------------------------------------------
	rid, err := rlib.InsertRentable(ctx, &r)
	if nil != err {
		errMsg := fmt.Sprintf("error inserting rlib.Rentable = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	if rid > 0 {
		for i := 0; i < len(rul); i++ {
			rul[i].RID = rid
			rul[i].BID = r.BID
			_, err := rlib.InsertRentableUser(ctx, &rul[i])
			if err != nil {
				errMsg := fmt.Sprintf("error saving rlib.RentableUser: %s", err.Error())
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
			}
		}
		for i := 0; i < len(m); i++ {
			m[i].RID = rid
			m[i].BID = r.BID
			_, err := rlib.InsertRentableStatus(ctx, &m[i])
			if err != nil {
				errMsg := fmt.Sprintf("error saving rlib.RentableStatus: %s", err.Error())
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
			}
		}
		for i := 0; i < len(n); i++ {
			n[i].RID = rid
			n[i].BID = r.BID
			_, err := rlib.InsertRentableTypeRef(ctx, &n[i])
			if err != nil {
				errMsg := fmt.Sprintf("error saving rlib.RentableTypeRef: %s", err.Error())
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
			}
		}
	}
	return 0, nil

}

// LoadRentablesCSV loads a csv file with rental specialty types and processes each one
func LoadRentablesCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateRentables)
}
