package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strings"
	"time"
)

// // AcctRule is a structure of the 3-tuple that makes up a whole part of an AcctRule
// type CSVAcctRule struct {
// 	Action  string // "d" = debit, "c" = credit
// 	Account string // GL No for the account
// 	Amount  string // use the entire amount of the assessment or deposit, otherwise the amount to use
// 	ASMID   string // Used only for rlib.ReceiptAllocation; the assessment that caused this payment
// }

// CVS record format:
// 0    1,    2           3      4      5             6             7        8                     9                                           10
// BID, TCID, RAID,       PMTID, DEPID, Dt,           DocNo,        Amount,  AR,                   AcctRule,                                   Comment
// REH, TCID, RA00000001, 2,     1,     "2004-01-01", 1254,         1000.00, "Rent Payment Check", "ASM(7) d ${rlib.DFLT} _, ASM(7) c 11002 _",
// REH, TCID, RA00000001, 1,     1,     "2015-11-21", 883789238746, 294.66,  "Rent Payment Check", "ASM(1) c ${GLGENRCV} 266.67, ASM(1) d ${rlib.DFLT} 266.67, ASM(3) c ${GLGENRCV} 13.33, ASM(3) d ${rlib.DFLT} 13.33, ASM(4) c ${GLGENRCV} 5.33, ASM(4) d ${rlib.DFLT} 5.33, ASM(9) c ${GLGENRCV} 9.33,ASM(9) d ${rlib.DFLT} 9.33", "I am a comment"

// GenerateReceiptAllocations processes the AcctRule for the supplied rlib.Receipt and generates rlib.ReceiptAllocation records
func GenerateReceiptAllocations(ctx context.Context, rcpt *rlib.Receipt, raid int64, xbiz *rlib.XBusiness) error {
	var d1 = time.Date(rcpt.Dt.Year(), rcpt.Dt.Month(), 1, 0, 0, 0, 0, time.UTC)
	var d2 = d1.AddDate(0, 0, 31)

	t, err := rlib.ParseAcctRule(ctx, xbiz, 0, &d1, &d2, rcpt.AcctRuleApply, rcpt.Amount, 1.0)
	if err != nil {
		return err
	}
	u := make(map[int64][]int64)

	// First, group together all entries that refer to a single ASMID into a list of lists
	for i := int64(0); i < int64(len(t)); i++ {
		u[t[i].ASMID] = append(u[t[i].ASMID], i)
	}
	// Process each list in the list of lists.
	for k, v := range u {
		var a rlib.ReceiptAllocation
		a.AcctRule = ""
		a.ASMID = k
		a.Amount = t[int(v[0])].Amount
		a.RCPTID = rcpt.RCPTID
		a.RAID = raid
		a.Dt = rcpt.Dt

		// make sure the referenced assessment actually exists
		a1, err := rlib.GetAssessment(ctx, a.ASMID)
		if err != nil {
			return fmt.Errorf("GenerateReceiptAllocations: GetAssessment with ID(%d) error: %s", a.ASMID, err.Error())
		}
		if a1.ASMID == 0 {
			return fmt.Errorf("GenerateReceiptAllocations: Referenced assessment ID %d does not exist", a.ASMID)
		}

		// for each index in the list, build its part of the AcctRule
		lim := int64(len(v))
		for i := int64(0); i < lim; i++ {
			j := int(v[i])
			a.AcctRule += fmt.Sprintf("ASM(%d) %s %s %s", t[j].ASMID, t[j].Action, t[j].AcctExpr, t[j].Expr)
			if i+1 < lim {
				a.AcctRule += ","
			}
		}
		a.BID = rcpt.BID
		_, err = rlib.InsertReceiptAllocation(ctx, &a)
		if err != nil {
			return fmt.Errorf("GenerateReceiptAllocations: Error inserting ReceiptAllocation: %s", err.Error())
		}
		rcpt.RA = append(rcpt.RA, a)
	}

	return nil
}

//var pmtTypes = map[int64]rlib.PaymentType{}

// CreateReceiptsFromCSV reads an assessment type string array and creates a database record for the assessment type
func CreateReceiptsFromCSV(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateReceiptsFromCSV"
	var (
		err  error
		xbiz rlib.XBusiness
		r    rlib.Receipt
		bud  = strings.ToLower(strings.TrimSpace(sa[0]))
	)

	const (
		BUD      = 0
		TCID     = iota
		RAID     = iota
		PMTID    = iota
		DEPID    = iota
		Dt       = iota
		DocNo    = iota
		Amount   = iota
		AR       = iota
		AcctRule = iota
		Comment  = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"TCID", TCID},
		{"RAID", RAID},
		{"PMTID", PMTID},
		{"DEPID", DEPID},
		{"Dt", Dt},
		{"DocNo", DocNo},
		{"Amount", Amount},
		{"AR", AR},
		{"AcctRule", AcctRule},
		{"Comment", Comment},
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
	if len(bud) > 0 {
		b1, err := rlib.GetBusinessByDesignation(ctx, bud)
		if err != nil {
			errMsg := fmt.Sprintf("error while getting business by designation(%s), error: %s", sa[BUD], err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		if len(b1.Designation) == 0 {
			errMsg := fmt.Sprintf("rlib.Business with designation %s does not exist", sa[BUD])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		r.BID = b1.BID
		err = rlib.GetXBusiness(ctx, r.BID, &xbiz)
		if err != nil {
			errMsg := fmt.Sprintf("error while getting business BID(%d): %s", r.BID, err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
	}

	//-------------------------------------------------------------------
	// Who is the payor?
	//-------------------------------------------------------------------
	payors, err := CSVLoaderTransactantList(ctx, r.BID, sa[TCID])
	if err != nil {
		errMsg := fmt.Sprintf("error: %s", err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, TCID, -1, errMsg)
	}
	// here we are expecting only one payor
	if len(payors) != 1 {
		errMsg := fmt.Sprintf("only one payor can be assigned: %s", sa[TCID])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, TCID, -1, errMsg)
	}
	if payors[0].TCID == 0 {
		errMsg := fmt.Sprintf("payor cannot be found: %s", sa[TCID])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, TCID, -1, errMsg)
	}
	r.TCID = payors[0].TCID

	pmtTypes, err := rlib.GetPaymentTypesByBusiness(ctx, r.BID)
	if err != nil {
		errMsg := fmt.Sprintf("error while getting payment types for BID(%d): %s", r.BID, err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// Find Rental Agreement
	//-------------------------------------------------------------------
	raid := CSVLoaderGetRAID(sa[RAID]) // this should probably go away, we should select it from an Assessment in the AcctRule

	ra, err := rlib.GetRentalAgreement(ctx, raid)
	if nil != err {
		errMsg := fmt.Sprintf("error loading Rental Agreement %s, err = %v", sa[RAID], err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RAID, -1, errMsg)
	}
	if ra.RAID == 0 {
		errMsg := fmt.Sprintf("error loading Rental Agreement %s", sa[RAID])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RAID, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// Get the rlib.PaymentType
	//-------------------------------------------------------------------
	r.PMTID, _ = rlib.IntFromString(sa[PMTID], "Payment type is invalid")
	_, ok := pmtTypes[r.PMTID]
	if !ok {
		errMsg := fmt.Sprintf("Payment type is invalid: %s", sa[PMTID])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, PMTID, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// Get the Depository
	//-------------------------------------------------------------------
	r.DEPID, err = rlib.IntFromString(sa[DEPID], "Depository ID is invalid")
	if err != nil {
		errMsg := fmt.Sprintf("Depository ID is invalid: %s", sa[DEPID])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, DEPID, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// Get the date
	//-------------------------------------------------------------------
	dt, err := rlib.StringToDate(sa[Dt])
	if err != nil {
		errMsg := fmt.Sprintf("invalid rlib.Receipt date:  %s", sa[Dt])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Dt, -1, errMsg)
	}
	r.Dt = dt

	//-------------------------------------------------------------------
	// Determine the DocNo
	//-------------------------------------------------------------------
	r.DocNo = strings.TrimSpace(sa[DocNo])

	//-------------------------------------------------------------------
	// Determine the amount
	//-------------------------------------------------------------------
	r.Amount, _ = rlib.FloatFromString(sa[Amount], "rlib.Receipt Amount is invalid")

	//-------------------------------------------------------------------
	// Set the ARID
	//-------------------------------------------------------------------
	s := strings.TrimSpace(sa[AR])
	if len(s) > 0 {
		rule, err := rlib.GetARByName(ctx, r.BID, s)
		if err != nil {
			errMsg := fmt.Sprintf("Could not load AR named %s: %s", sa[AR], err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, AR, -1, errMsg)
		}
		r.ARID = rule.ARID
	}

	//-------------------------------------------------------------------
	// Set the AcctRule.  No checking for now...
	//-------------------------------------------------------------------
	r.AcctRuleApply = strings.TrimSpace(sa[AcctRule])

	r.Comment = strings.TrimSpace(sa[Comment])

	//-------------------------------------------------------------------
	// Make sure everything that needs to be set actually got set...
	//-------------------------------------------------------------------
	if len(r.AcctRuleApply) == 0 || r.PMTID == 0 || r.Amount == 0 || r.BID == 0 {
		errMsg := fmt.Sprintf("Skipping this record")
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// Make sure there's no duplicate...
	//-------------------------------------------------------------------
	// TODO(Steve): ignore error?
	rdup, _ := rlib.GetReceiptDuplicate(ctx, &r.Dt, r.Amount, r.DocNo)
	if rdup.RCPTID != 0 {
		errMsg := fmt.Sprintf("this is a duplicate of an existing receipt: %s", rdup.IDtoString())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	rcptid, err := rlib.InsertReceipt(ctx, &r)
	if err != nil {
		errMsg := fmt.Sprintf("error inserting receipt: %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	r.RCPTID = rcptid

	//-------------------------------------------------------------------
	// Create the allocations...
	//-------------------------------------------------------------------
	err = GenerateReceiptAllocations(ctx, &r, raid, &xbiz)
	if err != nil {
		// TODO(Steve): ignore error?
		_ = rlib.DeleteReceipt(ctx, r.RCPTID)
		// TODO(Steve): ignore error?
		_ = rlib.DeleteReceiptAllocations(ctx, r.RCPTID)
		errMsg := fmt.Sprintf("error processing receipt: %s", err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	// fmt.Printf("Completed generating receipt %d.\n", r.RCPTID)
	//-------------------------------------------------------------------
	// first, make a complete pass through the Assessments to see if any
	// of them have already been marked as paid
	//-------------------------------------------------------------------
	for i := 0; i < len(r.RA); i++ {
		// fmt.Printf("Checking receipt allocation: %#v\n", r.RA[i])
		a, err := rlib.GetAssessment(ctx, r.RA[i].ASMID)
		if err != nil {
			errMsg := fmt.Sprintf("error marking assessments as paid: %s", err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
		// fmt.Printf("a.FLAGS = 0x%x\n", a.FLAGS)
		if a.FLAGS&1<<0 != 0 {
			errMsg := fmt.Sprintf("assessment %d is already marked as paid", a.ASMID)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
	}
	//-------------------------------------------------------------------
	// Now mark the allocated assessments as paid
	//-------------------------------------------------------------------
	for i := 0; i < len(r.RA); i++ {
		a, err := rlib.GetAssessment(ctx, r.RA[i].ASMID)
		if err != nil {
			errMsg := fmt.Sprintf("error marking assessments as paid: %s", err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
		a.FLAGS |= 1 << 0 // bit 0 is the "paid" flag
		err = rlib.UpdateAssessment(ctx, &a)
		if err != nil {
			errMsg := fmt.Sprintf("error marking assessments as paid: %s", err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
	}

	//-------------------------------------------------------------------
	// Process the receipt...
	//-------------------------------------------------------------------
	_, err = rlib.ProcessNewReceipt(ctx, Rcsv.Xbiz, &Rcsv.DtStart, &Rcsv.DtStop, &r)
	if err != nil {
		errMsg := fmt.Sprintf("error while processing new receipt: %s", err.Error())
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	return 0, nil
}

// LoadReceiptsCSV loads a csv file with a chart of accounts and creates rlib.GLAccount markers for each
func LoadReceiptsCSV(ctx context.Context, fname string) []error {
	var m []error

	t := rlib.LoadCSV(fname)
	if len(t) > 1 {
		//-------------------------------------------------------------------
		// Check to see if this rental specialty type is already in the database
		//-------------------------------------------------------------------
		des := strings.TrimSpace(t[1][0])
		if len(des) > 0 {
			b, err := rlib.GetBusinessByDesignation(ctx, des)
			if err != nil {
				m = append(m, err)
				return m
			}
			if b.BID < 1 {
				err := fmt.Errorf("LoadReceiptsCSV: rlib.Business named %s not found", des)
				m = append(m, err)
				return m
			}
			rlib.InitBusinessFields(b.BID)
			// _ = rlib.GetDefaultLedgers(ctx, b.BID) // the actually loads the RRdb.BizTypes array which is needed by rpn
		}
	}

	return LoadRentRollCSV(ctx, fname, CreateReceiptsFromCSV)
}
