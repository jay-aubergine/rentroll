package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strconv"
	"strings"
)

// CVS record format:
//	                GLAccount can be Account Name, GLNumber, or LID
// 0    1           2         3
// BUD, GLAccount,  Name,     AccountNo

// CreateDepositoriesFromCSV reads an assessment type string array and creates a database record for the assessment type
func CreateDepositoriesFromCSV(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateDepositoriesFromCSV"
	var (
		err error
		d   rlib.Depository
	)

	const (
		BUD       = 0
		LID       = iota
		Name      = iota
		AccountNo = iota
	)
	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"GLAccount", LID},
		{"Name", Name},
		{"AccountNo", AccountNo},
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
	if len(sa[BUD]) > 0 {
		b1, err := rlib.GetBusinessByDesignation(ctx, sa[BUD])
		if err != nil {
			errMsg := fmt.Sprintf("rlib.Business with designation %s does not exist", sa[BUD])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		if len(b1.Designation) == 0 {
			errMsg := fmt.Sprintf("rlib.Business with designation %s does not exist", sa[BUD])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		d.BID = b1.BID
	}

	if len(sa[LID]) > 0 {
		var acct rlib.GLAccount
		i, err := strconv.Atoi(sa[LID])

		// if LID is provided for the GLAccount then
		if err == nil {
			d.LID = int64(i)
		}

		// validate that this is a valid LID
		if d.LID > 0 {
			acct, _ = rlib.GetLedger(ctx, d.LID)
		}

		// If account was not found by LID, then try to find by GLNumber
		// maybe GLNumber is provided in sa[LID] column
		if acct.LID == 0 {
			gl, _ := rlib.GetLedgerByGLNo(ctx, d.BID, sa[LID])
			if gl.LID == 0 {
				// If again not found then try to find by AccountName
				// maybe AccountName is provided
				// strip the whilespaces in column value then try to find
				gl, _ = rlib.GetLedgerByName(ctx, d.BID, strings.TrimSpace(sa[LID])) // see if we can find it by name
			}
			d.LID = gl.LID
		}
	}
	if d.LID == 0 {
		errMsg := fmt.Sprintf("No GL Account with Name or AccountNumber = %s", sa[LID])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, LID, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// Name
	//-------------------------------------------------------------------
	d.Name = strings.TrimSpace(sa[Name])
	if len(d.Name) == 0 {
		errMsg := fmt.Sprintf("no name for Depository. Please supply a name")
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Name, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// AccountNo
	//-------------------------------------------------------------------
	d.AccountNo = strings.TrimSpace(sa[AccountNo])
	if len(d.AccountNo) == 0 {
		errMsg := fmt.Sprintf("no AccountNo for Depository. Please supply AccountNo")
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, AccountNo, -1, errMsg)
	}

	dup, _ := rlib.GetDepositoryByAccount(ctx, d.BID, d.AccountNo)
	if dup.DEPID != 0 {
		errMsg := fmt.Sprintf("depository with account number %s already exists", d.AccountNo)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, AccountNo, -1, errMsg)
	}

	_, err = rlib.InsertDepository(ctx, &d)
	if err != nil {
		errMsg := fmt.Sprintf("error inserting depository: %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	return 0, nil
}

// LoadDepositoryCSV loads a csv file with a chart of accounts and creates rlib.GLAccount markers for each
func LoadDepositoryCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateDepositoriesFromCSV)
}
