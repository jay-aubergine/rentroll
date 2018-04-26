package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strings"
)

//                            Ledger NAME or GLAccountID or LID works
// 0   1                      2             3
// Bud,Name,                  ARType,       Debit,              Credit,                    Allocated,RAIDRequired,SubARSpec             ,Description
// REX,Rent,                  Assessment,   2,                  8,                         No,       No,         ,                      ,Rent assessment
// REX,AutoGenFloatSECDEPAsmt,SubAssessment,RentRollReceivables,BankAcct,                  No,       No,         ,                      ,auto create assessment
// REX,RCVFloatSecDep,        Receipt,      Undeposited Funds,  Floating Security Deposits,Yes,      Yes,        ,AutoGenFloatSECDEPAsmt,take payment and apply to auto gen'd asmt
// REX,FNB,                   Receipt,      Receivables,        7,                         Yes,      Yes,        ,                      ,payments that are deposited in First National Bank

// CreateAR creates AR database records from the supplied CSV file lines
func CreateAR(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateAR"
	var (
		b   rlib.AR
		err error
	)

	const (
		BUD          = 0
		Name         = iota
		ARType       = iota
		DebitLID     = iota
		CreditLID    = iota
		Allocated    = iota
		ShowOnRA     = iota
		RAIDRequired = iota
		SubARSpec    = iota
		Description  = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"Name", Name},
		{"ARType", ARType},
		{"Debit", DebitLID},
		{"Credit", CreditLID},
		{"Allocated", Allocated},
		{"ShowOnRA", ShowOnRA},
		{"RAIDRequired", RAIDRequired},
		{"SubARSpec", SubARSpec},
		{"Description", Description},
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
	des := strings.TrimSpace(sa[BUD])
	if len(des) > 0 { // make sure it's not empty
		b1, err := rlib.GetBusinessByDesignation(ctx, des) // see if we can find the biz
		if err != nil {
			errMsg := fmt.Sprintf("Business with designation %s does not exist", sa[0])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		if len(b1.Designation) == 0 {
			errMsg := fmt.Sprintf("Business with designation %s does not exist", sa[0])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		b.BID = b1.BID
	}

	//-----------------------------------------
	// Get the name
	//-----------------------------------------
	b.Name = sa[Name]
	b2, err := rlib.GetARByName(ctx, b.BID, b.Name)
	if err != nil {
		errMsg := fmt.Sprintf("Error attempting to read existing records with name = %s: %s", b.Name, err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Name, -1, errMsg)
	}
	if b2.Name == b.Name {
		errMsg := fmt.Sprintf("An AR rule with name = %s already exists. Ignoring this line", b.Name)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Name, -1, errMsg)
	}

	//-----------------------------------------
	// Get the type
	//-----------------------------------------
	s := strings.TrimSpace(strings.ToLower(sa[ARType]))
	switch s {
	case "assessment":
		b.ARType = rlib.ARASSESSMENT
	case "receipt":
		b.ARType = rlib.ARRECEIPT
	case "expense":
		b.ARType = rlib.AREXPENSE
	case "sub-assessment":
		b.ARType = rlib.ARSUBASSESSMENT
	default:
		errMsg := fmt.Sprintf("ARType must be either Assessment or Receipt.  Found: %s", s)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, ARType, -1, errMsg)
	}

	//----------------------------------------------------------------
	// Get the Debit account
	//----------------------------------------------------------------
	var gl rlib.GLAccount
	// rlib.Console("sa[DebitLID] = %s\n", sa[DebitLID])
	b.DebitLID, err = rlib.IntFromString(sa[DebitLID], "Invalid DebitLID") // first see if it is a LID
	if err == nil && b.DebitLID > 0 {                                      // try the LID first
		gl, err = rlib.GetLedger(ctx, b.DebitLID)
		if err != nil {
			errMsg := fmt.Sprintf("error while getting ledger. Error: %s", err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, DebitLID, -1, errMsg)
		}
	}
	if gl.LID == 0 && len(sa[DebitLID]) > 0 {
		// if it's not a number, try the Name
		gl, _ = rlib.GetLedgerByName(ctx, b.BID, sa[DebitLID])
		if gl.LID == 0 {
			// if not a name, then try GLNumber
			gl, _ = rlib.GetLedgerByGLNo(ctx, b.BID, sa[DebitLID])
		}
	}
	// rlib.Console("DEBIT LID = %s\n", gl.Name)
	b.DebitLID = gl.LID
	if b.DebitLID == 0 {
		errMsg := fmt.Sprintf("Could not find GLAccount for = %s", sa[DebitLID])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, DebitLID, -1, errMsg)
	}

	//----------------------------------------------------------------
	// Get the Credit account
	//----------------------------------------------------------------
	var glc rlib.GLAccount
	// rlib.Console("sa[CreditLID] = %s\n", sa[CreditLID])
	b.CreditLID, err = rlib.IntFromString(sa[CreditLID], "Invalid CreditLID")
	if err == nil || b.CreditLID > 0 {
		// first try to search with LID
		glc, _ = rlib.GetLedger(ctx, b.CreditLID)
	}
	if glc.LID == 0 && len(sa[CreditLID]) > 0 {
		// if not a number, then try Name
		glc, _ = rlib.GetLedgerByName(ctx, b.BID, sa[CreditLID])
		if glc.LID == 0 {
			// if not a name, then try GLNumber
			glc, _ = rlib.GetLedgerByGLNo(ctx, b.BID, sa[CreditLID])
		}
	}
	// rlib.Console("CREDIT LID = %s\n", glc.Name)
	b.CreditLID = glc.LID
	if b.CreditLID == 0 {
		errMsg := fmt.Sprintf("Invalid GLAccountID: %s", sa[CreditLID])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, CreditLID, -1, errMsg)
	}

	//----------------------------------------------------------------
	// Set flags
	//----------------------------------------------------------------
	var alloc, show, rarqd int64
	var yn string
	yn = strings.TrimSpace(sa[Allocated])
	if len(yn) == 0 {
		yn = "no"
	}
	alloc, err = rlib.YesNoToInt(yn)
	if err != nil {
		errMsg := fmt.Sprintf("invalid Allocated column value: %s", sa[Allocated])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Allocated, -1, errMsg)
	}
	if alloc > 0 {
		b.FLAGS |= 1 << 0
	}
	yn = strings.TrimSpace(sa[ShowOnRA])
	if len(yn) == 0 {
		yn = "no"
	}
	show, err = rlib.YesNoToInt(yn)
	if err != nil {
		errMsg := fmt.Sprintf("invalid ShowOnRA column value: %s", sa[ShowOnRA])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, ShowOnRA, -1, errMsg)
	}
	if show > 0 {
		b.FLAGS |= 1 << 1
	}
	yn = strings.TrimSpace(sa[RAIDRequired])
	if len(yn) == 0 {
		yn = "no"
	}
	rarqd, err = rlib.YesNoToInt(yn)
	if err != nil {
		errMsg := fmt.Sprintf("invalid RAIDRequired column value: %s", sa[RAIDRequired])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RAIDRequired, -1, errMsg)
	}
	if rarqd > 0 {
		b.FLAGS |= 1 << 2
	}

	b.DtStart = rlib.TIME0
	b.DtStop = rlib.ENDOFTIME

	//----------------------------------------------------------------
	// Get SubARs - add the array of subars to b, then update the
	// subars with this ARID after we've saved b below
	//----------------------------------------------------------------
	sars := strings.TrimSpace(sa[SubARSpec])
	if len(sars) > 0 {
		sarsa := strings.Split(sars, ",")
		for i := 0; i < len(sarsa); i++ {
			var x rlib.SubAR
			x.BID = b.BID
			subar, err := rlib.GetARByName(ctx, b.BID, sarsa[i])
			if err != nil {
				errMsg := fmt.Sprintf("could not get Account Rule named: %s, err: %s", sarsa[i], err.Error())
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, SubARSpec, i, errMsg)
			}
			x.SubARID = subar.ARID
			b.SubARs = append(b.SubARs, x) // we will need to update these structs with b.ARID after we save it below
		}
		if len(b.SubARs) > 0 {
			b.FLAGS |= (1 << 3) // bit 3 indicates that there are subars on this rule
		}
	}

	//----------------------------------------------------------------
	// Get the Description
	//----------------------------------------------------------------
	b.Description = sa[Description]

	_, err = rlib.InsertAR(ctx, &b)
	if err != nil {
		errMsg := fmt.Sprintf("%s: error inserting AR = %v", funcname, err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Description, -1, errMsg)
	}

	for i := 0; i < len(b.SubARs); i++ {
		b.SubARs[i].ARID = b.ARID
		_, err = rlib.InsertSubAR(ctx, &b.SubARs[i])
		if err != nil {
			errMsg := fmt.Sprintf("error saving SubAR[%d]: %s", i, err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, SubARSpec, i, errMsg)
		}
	}

	return 0, nil
}

// LoadARCSV loads the values from the supplied csv file and creates AR records.
func LoadARCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateAR)
}
