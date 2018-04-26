package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strconv"
	"strings"
	"time"
)

// TODO:
// Remove Associated
// Remove Collective
// Allow Posting
// RA Required

//   0   1                             2          3,               4                    5          6                 7           8
// BUD,  Name,                         GLNumber,  Parent GLNumber, Account Type,        Balance,   Account Status,   Date,       Description
// REH,  Bank Account FRB 2332352,     10001,     10000,           bank,                0,         active,         "2016-03-01", Bla bla bla
// REH,  General Accounts Receivable,  11001,     11000,           Accounts Receivable, 0,         active,         "2016-03-01", Bla bla bla
// REH,  Friday Lunch Fund,            11099,     11000,           Accounts Receivable, 0.00,      active,

// CreateLedgerMarkers reads an assessment type string array and creates a database record for the assessment type
func CreateLedgerMarkers(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateLedgerMarkers"
	var (
		err       error
		inserting = true // this may be changed, depends on the value for sa[7]
		lm        rlib.LedgerMarker
		l         rlib.GLAccount
		parent    rlib.GLAccount
	)

	const (
		BUD            = 0
		Name           = iota
		GLNumber       = iota
		ParentGLNumber = iota
		AccountType    = iota
		Balance        = iota
		AccountStatus  = iota
		Date           = iota
		Description    = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"Name", Name},
		{"GLNumber", GLNumber},
		{"ParentGLNumber", ParentGLNumber},
		{"AccountType", AccountType},
		{"Balance", Balance},
		{"AccountStatus", AccountStatus},
		{"Date", Date},
		{"Description", Description},
	}

	l.AllowPost = 1 // default is to allow, server will modify as needed

	y, err := ValidateCSVColumnsErr(csvCols, sa, funcname, lineno)
	if y {
		return 1, err
	}
	if lineno == 1 {
		return 0, nil // we've validated the col headings, all is good, send the next line
	}
	des := strings.ToLower(strings.TrimSpace(sa[0]))
	//-------------------------------------------------------------------
	// Make sure the rlib.Business is in the database
	//-------------------------------------------------------------------
	if len(des) > 0 {
		// rlib.Console("Looking for BUD:  %s\n", des)
		// TODO(Steve): ignore error?
		b1, _ := rlib.GetBusinessByDesignation(ctx, des)
		if len(b1.Designation) == 0 {
			errMsg := fmt.Sprintf("rlib.Business with designation %s does not exist", sa[BUD])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		lm.BID = b1.BID
		l.BID = b1.BID
		// rlib.Console("BUD %s  --->  BID = %d\n", des, l.BID)
	}

	lm.State = rlib.LMINITIAL // Initial marker, no prior records

	//----------------------------------------------------------------------
	// NAME
	//----------------------------------------------------------------------
	l.Name = strings.TrimSpace(sa[Name])

	// rlib.Console("B\n")
	//----------------------------------------------------------------------
	// GLNUMBER
	// Make sure the account number is unique
	//----------------------------------------------------------------------
	// rlib.Console("sa[GLNumber] = %q\n", sa[GLNumber])
	g := strings.TrimSpace(sa[GLNumber])
	// rlib.Console("len(g) = %d\n", len(sa[GLNumber]))
	if len(g) == 0 {
		errMsg := fmt.Sprintf("You must supply a GL Number for this entry")
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, GLNumber, -1, errMsg)
	}
	if len(g) > 0 {
		// if we're inserting a record then it must not already exist
		// rlib.Console("inserting = %t\n", inserting)
		if inserting {
			// rlib.Console("lm.BID = %d, getting ledger by GLNo:  %s\n", lm.BID, g)
			// TODO(Steve): ignore error?
			ldg, _ := rlib.GetLedgerByGLNo(ctx, lm.BID, g)
			// rlib.Console("ldg.LID = %d, name = %s\n", ldg.LID, ldg.Name)
			if ldg.LID > 0 {
				errMsg := fmt.Sprintf("Account already exists: %s", g)
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, GLNumber, -1, errMsg)
			}
			// // was there an error in finding an account with this GLNo?
			// if err != nil {
			// 	return CsvErrorSensitivity, fmt.Errorf("%s: line %d, GL Account %s already exists", funcname, lineno, g)
			// 	return rs,CsvErrorSensitivity
			// }
		}

		l.GLNumber = g
	}

	// rlib.Console("C\n")
	//----------------------------------------------------------------------
	// PARENT GLNUMBER
	//----------------------------------------------------------------------
	l.PLID = int64(0) // assume no parent
	g = strings.TrimSpace(sa[ParentGLNumber])
	if len(g) > 0 {
		parent, err = rlib.GetLedgerByGLNo(ctx, l.BID, g)
		if err != nil {
			errMsg := fmt.Sprintf("Error getting GLAccount: %s", g)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, ParentGLNumber, -1, errMsg)
		}
		if parent.LID == 0 {
			errMsg := fmt.Sprintf("Error getting GLAccount: %s", g)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, ParentGLNumber, -1, errMsg)
		}
		l.PLID = parent.LID
	}
	// rlib.Console("D\n")

	//----------------------------------------------------------------------
	// ACCOUNT TYPE
	//----------------------------------------------------------------------
	l.AcctType = strings.TrimSpace(sa[AccountType])

	//----------------------------------------------------------------------
	// OPENING BALANCE
	//----------------------------------------------------------------------
	lm.Balance = float64(0) // assume a 0 starting balance
	g = strings.TrimSpace(sa[Balance])
	if len(g) > 0 {
		x, err := strconv.ParseFloat(g, 64)
		if err != nil {
			errMsg := fmt.Sprintf("Invalid balance: %s", sa[Balance])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Balance, -1, errMsg)
		}
		lm.Balance = x
	}
	// rlib.Console("E\n")

	//----------------------------------------------------------------------
	// GLACCOUNT STATUS
	//----------------------------------------------------------------------
	s := strings.ToLower(strings.TrimSpace(sa[AccountStatus]))
	if "active" == s {
		l.Status = rlib.ACCTSTATUSACTIVE
	} else if "inactive" == s {
		l.Status = rlib.ACCTSTATUSINACTIVE
	} else {
		errMsg := fmt.Sprintf("Invalid account status: %s", sa[AccountStatus])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, AccountStatus, -1, errMsg)
	}

	// rlib.Console("F\n")

	//----------------------------------------------------------------------
	// DATE for opening balance
	//----------------------------------------------------------------------
	_, err = rlib.StringToDate(sa[Date])
	if err != nil {
		errMsg := fmt.Sprintf("Invalid stop date: %s", sa[Date])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Date, -1, errMsg)
	}
	lm.Dt = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC) // always force the initial ledger marker to "the beginning of time"

	// //----------------------------------------------------------------------
	// // ALLOW POST
	// //----------------------------------------------------------------------
	// l.AllowPost, err = rlib.YesNoToInt(sa[AllowPosting])
	// if err != nil {
	// 	return CsvErrorSensitivity, fmt.Errorf("%s: line %d - invalid value for AllowPost:  %s", funcname, lineno, sa[AllowPosting])
	// }

	//----------------------------------------------------------------------
	// DESCRIPTION
	//----------------------------------------------------------------------
	if len(sa[Description]) > 1024 {
		b := []byte(sa[Description])
		l.Description = string(b[:1024])
	} else {
		l.Description = sa[Description]
	}

	//=======================================================================================

	// rlib.Console("LOADCSV - SAVE:  Inserting = %v\n", inserting)
	// rlib.Console("                 l = %#v\n", l)

	// Insert / Update the rlib.GLAccount first, we may need the LID
	if inserting {
		var lid int64
		lid, err = rlib.InsertLedger(ctx, &l)
		// rlib.Console("Inserted new account:  BID = %d, LID = %d, Name = %s\n", l.BID, lid, l.Name)
		lm.LID = lid
	} else {
		err = rlib.UpdateLedger(ctx, &l)
		lm.LID = l.LID
	}
	if nil != err {
		errMsg := fmt.Sprintf("Could not save rlib.GLAccount marker, err = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	// Now update the markers
	if inserting {
		_, err = rlib.InsertLedgerMarker(ctx, &lm)
	} else {
		err = rlib.UpdateLedgerMarker(ctx, &lm)
	}
	if nil != err {
		errMsg := fmt.Sprintf("Could not save rlib.GLAccount marker, err = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	//--------------------------------------------------------------
	// If this entry called out a Parent, make sure the Parent's
	// AllowPost attribute is set properly
	//--------------------------------------------------------------
	if l.PLID > 0 && l.PLID == parent.LID {
		if parent.AllowPost == 1 {
			parent.AllowPost = 0
			err = rlib.UpdateLedger(ctx, &parent)
			if err != nil {
				errMsg := fmt.Sprintf("Could not save rlib.GLAccount marker, err = %v", err.Error())
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
			}
		}
	}

	return 0, nil
}

// LoadChartOfAccountsCSV loads a csv file with a chart of accounts and creates rlib.GLAccount markers for each
func LoadChartOfAccountsCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateLedgerMarkers)
}
