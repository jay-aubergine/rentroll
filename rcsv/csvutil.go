package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strings"
)

// CSVColumn defines a column of the CSV file
type CSVColumn struct {
	Name  string
	Index int
}

// CsvErrLoose et al, are constants used to control whether an error on a single line causes
// the entire CSV process to terminate or continue.   If LOOSE, then it will skip the error line
// and continue to process the remaining lines.  If STRICT, then the entire CSV loading process
// will terminate if any error is encountered
const (
	CsvErrLoose  = 0
	CsvErrStrict = 1

	// DupTransactant et al., are error identfiers for the CSV Loader
	DupTransactant        = "DuplicateTransactant"
	DupRentableType       = "DuplicateRentableType"
	DupCustomAttribute    = "DuplicateCustomAttribute"
	DupRentable           = "DuplicateRentable"
	RentableAlreadyRented = "RentableAlreadyRented"
)

// CsvErrorSensitivity is the error return value used by all the loadXYZcsv.go routines. We
// initialize to LOOSE as it is best for testing and should be OK for normal use as well.
var CsvErrorSensitivity = int(CsvErrLoose)

// CSVLoadHandlerFunc type of load handler function
type CSVLoadHandlerFunc func(context.Context, string) []error

// CSVLoadHandler struct is for routines that want to table-ize their loading.
type CSVLoadHandler struct {
	Fname   string
	Handler CSVLoadHandlerFunc
}

type csvHandlerFunc func(context.Context, []string, int) (int, error)

// LoadRentRollCSV performs a general purpose load.  It opens the supplied file name, and processes
// it line-by-line by calling the supplied handler function.
// Return Values
//		[]error  -  an array of errors encountered by the handler function during the load
//--------------------------------------------------------------------------------------------------
func LoadRentRollCSV(ctx context.Context, fname string, handler csvHandlerFunc) []error {
	var m []error
	t := rlib.LoadCSV(fname)
	for i := 0; i < len(t); i++ {
		if len(t[i][0]) == 0 {
			continue
		}
		if t[i][0][0] == '#' { // if it's a comment line, don't process it, just move on
			continue
		}
		s, err := handler(ctx, t[i], i+1)
		if err != nil {
			m = append(m, err)
		}
		if s > 0 { // if handler indicates that we need to stop...
			break //... then exit out of the loop
		}
	}
	return m
}

// ValidateCSVColumnsErr verifies the column titles with the supplied, expected titles.
// Returns:
//   bool --> false = everything is OK,  true = at least 1 column is wrong, error message already printed
//   err  --> nil if no problems
func ValidateCSVColumnsErr(csvCols []CSVColumn, sa []string, funcname string, lineno int) (bool, error) {
	required := len(csvCols)
	if len(sa) < required {
		l := len(sa)
		for i := 0; i < len(csvCols); i++ {
			if i < l {
				s := rlib.Stripchars(strings.ToLower(strings.TrimSpace(sa[i])), " ")
				if s != strings.ToLower(csvCols[i].Name) {
					errMsg := fmt.Sprintf("Error at column heading %d, expected %s, found %s", i+1, csvCols[i].Name, sa[i])
					return true, formatCSVErrors(funcname, lineno, i, -1, errMsg)
				}
			}
		}
		errMsg := fmt.Sprintf("found %d values, there must be at least %d", len(sa), required)
		return true, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	if lineno == 1 {
		for i := 0; i < len(csvCols); i++ {
			s := rlib.Stripchars(strings.ToLower(strings.TrimSpace(sa[i])), " ")
			if s != strings.ToLower(csvCols[i].Name) {
				errMsg := fmt.Sprintf("Error at column heading %d, expected %s, found %s", i+1, csvCols[i].Name, sa[i])
				return true, formatCSVErrors(funcname, lineno, i, -1, errMsg)
			}
		}
	}
	return false, nil
}

// ValidateCSVColumns wrapper for ValidateCSVColumnsErr
func ValidateCSVColumns(csvCols []CSVColumn, sa []string, funcname string, lineno int) (int, error) {
	t := 0
	b, err := ValidateCSVColumnsErr(csvCols, sa, funcname, lineno)
	if b {
		t = 1
	}
	return t, err
}

// CSVLoaderTransactantList takes a comma separated list of email addresses and phone numbers
// and returns an array of transactants for each.  If any of the addresses in the list
// cannot be resolved to a rlib.Transactant, then processing stops immediately and an error is returned.
func CSVLoaderTransactantList(ctx context.Context, BID int64, s string) ([]rlib.Transactant, error) {
	const funcname = "CSVLoaderTransactantList"

	var (
		err error
		m   []rlib.Transactant
	)

	if "" == s {
		return m, nil
	}

	s2 := strings.TrimSpace(s) // either the email address or the phone number
	ss := strings.Split(s2, ",")
	for i := 0; i < len(ss); i++ {
		var a rlib.Transactant
		s = strings.TrimSpace(ss[i])                          // either the email address or the phone number
		n, ok := readNumAndStatusFromExpr(s, "^TC0*(.*)", "") // "" suppresses error messages
		if len(ok) == 0 {
			err = rlib.GetTransactant(ctx, n, &a)
			if err != nil {
				rerr := fmt.Errorf("%s:  error retrieving Transactant with TCID, phone, or email: %s", funcname, s)
				return m, rerr
			}
		} else {
			a, err = rlib.GetTransactantByPhoneOrEmail(ctx, BID, s)
			if err != nil {
				rerr := fmt.Errorf("%s:  error retrieving Transactant with TCID, phone, or email: %s", funcname, s)
				return m, rerr
			}
		}
		if 0 == a.TCID {
			rerr := fmt.Errorf("%s:  error retrieving Transactant with TCID, phone, or email: %s", funcname, s)
			//fmt.Printf("%s\n", rerr.Error())
			return m, rerr
		}
		m = append(m, a)
	}
	return m, err
}

// ErrlistToString converts an errorlist into a string suitable for printout
func ErrlistToString(m *[]error) string {
	rs := ""
	for i := 0; i < len(*m); i++ {
		s := (*m)[i].Error()
		if s[len(s)-1:] != "\n" {
			s += "\n"
		}
		rs += s
	}
	return rs
}

// BuildPayorList takes a semi-colon separated list of email addresses and phone numbers
// and returns an array of rlib.RentalAgreementPayor records for each.  If any of the addresses in the list
// cannot be resolved to a rlib.Transactant, then processing stops immediately and an error is returned.
// Each value is time sensitive (has an associated time range). If the dates are not specified, then the
// default values of dfltStart and dfltStop -- which are the start/stop time of the rental agreement --
// are used instead. This is common because the payors will usually be the same for the entire rental
// agreement lifetime.
func BuildPayorList(ctx context.Context, BID int64, s string, dfltStart, dfltStop string, funcname string, lineno int, columnNo int) ([]rlib.RentalAgreementPayor, error) {
	var (
		m []rlib.RentalAgreementPayor
		// err error
	)
	// var noerr error
	s2 := strings.TrimSpace(s) // either the email address or the phone number
	if len(s2) == 0 {
		errMsg := fmt.Sprintf("Required Payor field is blank")
		return m, formatCSVErrors(funcname, lineno, columnNo, -1, errMsg)
	}
	s1 := strings.Split(s2, ";")
	for i := 0; i < len(s1); i++ {
		ss := strings.Split(s1[i], ",")
		if len(ss) != 3 {
			errMsg := fmt.Sprintf("invalid Payor Status syntax. Each semi-colon separated field must have 3 values. Found %d in \"%s\"", len(ss), ss)
			return m, formatCSVErrors(funcname, lineno, columnNo, i, errMsg)
		}
		s = strings.TrimSpace(ss[0]) // either the email address or the phone number or TransactantID (TC0003234)
		if len(s) == 0 {
			errMsg := fmt.Sprintf("Required Payor field is blank")
			return m, formatCSVErrors(funcname, lineno, columnNo, i, errMsg)
		}
		n, err := CSVLoaderTransactantList(ctx, BID, s)
		if err != nil {
			errMsg := fmt.Sprintf("could not find rlib.Transactant with contact information %s", s)
			return m, formatCSVErrors(funcname, lineno, columnNo, i, errMsg)
		}
		if len(n) == 0 {
			errMsg := fmt.Sprintf("could not find rlib.Transactant with contact information %s", s)
			return m, formatCSVErrors(funcname, lineno, columnNo, i, errMsg)
		}

		var payor rlib.RentalAgreementPayor
		payor.TCID = n[0].TCID

		// Now grab the dates
		if len(strings.TrimSpace(ss[1])) == 0 {
			ss[1] = dfltStart
		}
		if len(strings.TrimSpace(ss[2])) == 0 {
			ss[2] = dfltStop
		}
		// TODO(Steve): should we ignore the error?
		payor.DtStart, payor.DtStop, _ = readTwoDates(ss[1], ss[2])

		m = append(m, payor)
	}
	return m, nil
}

// BuildUserList parses a UserSpec and returns an array of RentableUser structs
func BuildUserList(ctx context.Context, BID int64, sa, dfltStart, dfltStop string, funcname string, lineno int, columnNo int) ([]rlib.RentableUser, error) {
	var (
		m []rlib.RentableUser
		// err error
	)

	s2 := strings.TrimSpace(sa) // TCID, email address, or the phone number
	if len(s2) == 0 {
		errMsg := fmt.Sprintf("Required User field is blank")
		return m, formatCSVErrors(funcname, lineno, columnNo, -1, errMsg)
	}
	s1 := strings.Split(s2, ";")
	var noerr error
	for i := 0; i < len(s1); i++ {
		ss := strings.Split(s1[i], ",")
		if len(ss) != 3 {
			errMsg := fmt.Sprintf("invalid Status syntax. Each semi-colon separated field must have 3 values. Found %d in \"%s\"", len(ss), ss)
			return m, formatCSVErrors(funcname, lineno, columnNo, i, errMsg)
		}
		s := strings.TrimSpace(ss[0]) // TCID, email address, or the phone number
		if len(s) == 0 {
			errMsg := fmt.Sprintf("Required User field is blank")
			return m, formatCSVErrors(funcname, lineno, columnNo, i, errMsg)
		}
		n, err := CSVLoaderTransactantList(ctx, BID, s)
		if err != nil {
			errMsg := fmt.Sprintf("invalid person identifier: %s. Error = %s", s, err.Error())
			return m, formatCSVErrors(funcname, lineno, columnNo, i, errMsg)
		}
		var p rlib.RentableUser
		p.TCID = n[0].TCID

		if len(strings.TrimSpace(ss[1])) == 0 {
			ss[1] = dfltStart
		}
		if len(strings.TrimSpace(ss[2])) == 0 {
			ss[2] = dfltStop
		}
		// TODO(Steve): should we ignore the error?
		p.DtStart, p.DtStop, _ = readTwoDates(ss[1], ss[2])
		m = append(m, p)
	}
	return m, noerr
}

// formatCSVErrors function formats the error and sends it
// eg. FunctionName: line 24, column 4, item 2 > ThisIsTheErrorMessage
// eg. FunctionName: line 24, column 4 > ThisIsTheErrorMessage
func formatCSVErrors(functionName string, lineNo int, columnNo int, itemNo int, errorMsg string) error {
	if columnNo > -1 {
		columnNo = columnNo + 1
	}
	if itemNo > -1 {
		itemNo = itemNo + 1
		return fmt.Errorf("%s: line %d, column %d, item %d >>> %s", functionName, lineNo, columnNo, itemNo, errorMsg)
	}
	return fmt.Errorf("%s: line %d, column %d >>> %s", functionName, lineNo, columnNo, errorMsg)
}

// // BID is the business id of the business unit to which the people belong
// func x(BID int64) {
// 	rows, err := rlib.RRdb.Prepstmt.GetAllTransactantsForBID.Query(BID)
// 	rlib.Errcheck(err)
// 	defer rows.Close()
// 	for rows.Next() {
// 		var tr rlib.Transactant
// 		rlib.ReadTransactants(rows, &tr)
// 		// Now dow whatever you need to do with the information in the transactant tr
// 	}
// 	rlib.Errcheck(rows.Err())
// }
