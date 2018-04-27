package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strconv"
	"strings"
)

//  CSV file format:
//  0      1               2             3             4                                           5               6        7                  8                                                              9
//  BUD,   RATemplateName, AgreementStart,  AgreementStop,   Payor,dtStart,dtStop;...              UserSpec,       Renewal, SpecialProvisions, "RentableName1,ContractRent2;RentableName2,ContractName2;...", Notes
//  BUD,   RATemplateName, AgreementStart,  AgreementStop,   PayorSpec,                            Usr1,d1,d2;..., Renewal, SpecialProvisions, "RentableName1,ContractRent2;RentableName2,ContractName2;...", Notes
// 	REH,   "RAT001",       "2004-01-01", "2015-11-08", "866-123-4567,dtStart,dtStop;bill@x.com...",UserSpec,       1,       "",                “U101,2500.00;U102,2350.00”,
// 	REH,   "RAT001",       "2004-01-01", "2017-07-04", "866-123-4567,dtStart,dtStop;bill@x.com",   UserSpec,       1,       "",                “U101,2500.00;U102,2350.00”,
// 	REH,   "RAT001",       "2015-11-21", "2016-11-21", "866-123-4567,,;bill@x.com,,",              UserSpec,       1,       "",                “U101,2500.00;U102,2350.00”,

// CreateRentalAgreement creates database records for the rental agreement defined in sa[]
func CreateRentalAgreement(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateRentalAgreement"
	var (
		err error
		ra  rlib.RentalAgreement
		m   []rlib.RentalAgreementRentable
	)

	const (
		BUD                 = 0
		RATemplateName      = iota
		AgreementStart      = iota
		AgreementStop       = iota
		PossessionStart     = iota
		PossessionStop      = iota
		RentStart           = iota
		RentStop            = iota
		RentCycleEpoch      = iota
		PayorSpec           = iota
		UserSpec            = iota
		UnspecifiedAdults   = iota
		UnspecifiedChildren = iota
		Renewal             = iota
		SpecialProvisions   = iota
		RentableSpec        = iota
		Notes               = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"RATemplateName", RATemplateName},
		{"AgreementStart", AgreementStart},
		{"AgreementStop", AgreementStop},
		{"PossessionStart", PossessionStart},
		{"PossessionStop", PossessionStop},
		{"RentStart", RentStart},
		{"RentStop", RentStop},
		{"RentCycleEpoch", RentCycleEpoch},
		{"PayorSpec", PayorSpec},
		{"UserSpec", UserSpec},
		{"UnspecifiedAdults", UnspecifiedAdults},
		{"UnspecifiedChildren", UnspecifiedChildren},
		{"Renewal", Renewal},
		{"SpecialProvisions", SpecialProvisions},
		{"RentableSpec", RentableSpec},
		{"Notes", Notes},
	}

	y, err := ValidateCSVColumnsErr(csvCols, sa, funcname, lineno)
	if y {
		return 1, err
	}
	if lineno == 1 {
		return 0, nil // we've validated the col headings, all is good, send the next line
	}

	//-------------------------------------------------------------------
	// BUD
	//-------------------------------------------------------------------
	cmpdes := strings.TrimSpace(sa[BUD])
	if len(cmpdes) > 0 {
		b2, err := rlib.GetBusinessByDesignation(ctx, cmpdes)
		if err != nil {
			errMsg := fmt.Sprintf("error while getting business by designation(%s), error: %s", sa[BUD], err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		if b2.BID == 0 {
			errMsg := fmt.Sprintf("could not find rlib.Business named %s", sa[BUD])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
		}
		ra.BID = b2.BID
	}

	//-------------------------------------------------------------------
	// Make sure the RentalTemplate exists
	//-------------------------------------------------------------------
	des := strings.ToLower(strings.TrimSpace(sa[RATemplateName]))
	if len(des) > 0 {
		b1, err := rlib.GetRentalAgreementByRATemplateName(ctx, des)
		if err != nil {
			errMsg := fmt.Sprintf("error while getting ra template %s: %s", sa[RATemplateName], err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RATemplateName, -1, errMsg)
		}
		if len(b1.RATemplateName) == 0 {
			errMsg := fmt.Sprintf("ra template %s does not exist", sa[RATemplateName])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RATemplateName, -1, errMsg)
		}
		ra.RATID = b1.RATID
	}

	//-------------------------------------------------------------------
	// AgreementStartDate
	//-------------------------------------------------------------------
	dfltStart := sa[AgreementStart]
	DtStart, err := rlib.StringToDate(dfltStart)
	if err != nil {
		errMsg := fmt.Sprintf("invalid agreement start date:  %s", sa[AgreementStart])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, AgreementStart, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// AgreementStopDate
	//-------------------------------------------------------------------
	dfltStop := sa[AgreementStop]
	DtStop, err := rlib.StringToDate(dfltStop)
	if err != nil {
		errMsg := fmt.Sprintf("invalid agreement stop date:  %s", sa[AgreementStop])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, AgreementStop, -1, errMsg)
	}
	ra.AgreementStop = DtStop

	// Initialize to default values
	ra.AgreementStart = DtStart
	ra.PossessionStart = DtStart
	ra.RentStart = DtStart
	ra.RentCycleEpoch = DtStart
	ra.PossessionStop = ra.AgreementStop
	ra.RentStop = ra.AgreementStop

	if len(sa[PossessionStart]) > 0 {
		ra.PossessionStart, err = rlib.StringToDate(sa[PossessionStart])
		if err != nil {
			errMsg := fmt.Sprintf("invalid possession start date:  %s", sa[PossessionStart])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, PossessionStart, -1, errMsg)
		}
	}
	if len(sa[PossessionStop]) > 0 {
		ra.PossessionStop, err = rlib.StringToDate(sa[PossessionStop])
		if err != nil {
			errMsg := fmt.Sprintf("invalid possession stop date:  %s", sa[PossessionStop])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, PossessionStop, -1, errMsg)
		}
	}
	if len(sa[RentStart]) > 0 {
		ra.RentStart, err = rlib.StringToDate(sa[RentStart])
		if err != nil {
			errMsg := fmt.Sprintf("invalid Rent start date:  %s", sa[RentStart])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentStart, -1, errMsg)
		}
	}
	if len(sa[RentStop]) > 0 {
		ra.RentStop, err = rlib.StringToDate(sa[RentStop])
		if err != nil {
			errMsg := fmt.Sprintf("invalid Rent stop date:  %s", sa[RentStop])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentStop, -1, errMsg)
		}
	}
	if len(sa[RentCycleEpoch]) > 0 {
		ra.RentCycleEpoch, err = rlib.StringToDate(sa[RentCycleEpoch])
		if err != nil {
			errMsg := fmt.Sprintf("invalid Rent cycle epoch date:  %s", sa[RentCycleEpoch])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentCycleEpoch, -1, errMsg)
		}
	}

	//-------------------------------------------------------------------
	//  The Payors
	//-------------------------------------------------------------------
	payors, err := BuildPayorList(ctx, ra.BID, sa[PayorSpec], dfltStart, dfltStop, funcname, lineno, PayorSpec)
	if err != nil { // save the full list
		return CsvErrorSensitivity, err
	}

	//-------------------------------------------------------------------
	//  The Users
	//-------------------------------------------------------------------
	users, err := BuildUserList(ctx, ra.BID, sa[UserSpec], dfltStart, dfltStop, funcname, lineno, UserSpec)
	if err != nil { // save the full list
		return CsvErrorSensitivity, err
	}
	//-------------------------------------------------------------------
	//  The Unspecified Adults and Children
	//-------------------------------------------------------------------
	s := strings.TrimSpace(sa[UnspecifiedAdults])
	if len(s) > 0 {
		i, err := strconv.Atoi(s)
		if err != nil {
			errMsg := fmt.Sprintf("UnspecifiedAdults value is invalid: %s", sa[UnspecifiedAdults])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, UnspecifiedAdults, -1, errMsg)
		}
		ra.UnspecifiedAdults = int64(i)
	}
	s = strings.TrimSpace(sa[UnspecifiedChildren])
	if len(s) > 0 {
		i, err := strconv.Atoi(s)
		if err != nil {
			errMsg := fmt.Sprintf("UnspecifiedChildren value is invalid: %s", sa[UnspecifiedChildren])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, UnspecifiedChildren, -1, errMsg)
		}
		ra.UnspecifiedChildren = int64(i)
	}

	//-------------------------------------------------------------------
	// Renewal
	//-------------------------------------------------------------------
	s = strings.TrimSpace(sa[Renewal])
	if len(s) > 0 {
		i, err := strconv.Atoi(s)
		if err != nil {
			errMsg := fmt.Sprintf("Renewal value is invalid: %s", sa[Renewal])
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Renewal, -1, errMsg)
		}
		ra.Renewal = int64(i)
	}

	//-------------------------------------------------------------------
	// Special Provisions
	//-------------------------------------------------------------------
	ra.SpecialProvisions = sa[SpecialProvisions]

	//-------------------------------------------------------------------
	// Rentables  -- all remaining columns are rentables
	//-------------------------------------------------------------------
	ss := strings.Split(strings.TrimSpace(sa[RentableSpec]), ";")
	if len(ss) > 0 && len(ss[0]) > 0 {
		for i := 0; i < len(ss); i++ {
			sss := strings.Split(ss[i], ",")
			if len(sss) != 2 {
				errMsg := fmt.Sprintf("Badly formatted string: %s . Format for each semicolon delimited part must be RentableName,ContractRent", ss)
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableSpec, i, errMsg)
			}
			var rar rlib.RentalAgreementRentable
			rnt, err := rlib.GetRentableByName(ctx, sss[0], ra.BID)
			if err != nil {
				errMsg := fmt.Sprintf("Could not load rentable named: %s  err = %s", sss[0], err.Error())
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableSpec, i, errMsg)
			}
			x, err := strconv.ParseFloat(strings.TrimSpace(sss[1]), 64)
			if err != nil {
				errMsg := fmt.Sprintf("Invalid amount: %s", sss[1])
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableSpec, i, errMsg)
			}
			rar.RID = rnt.RID
			rar.RARDtStart = DtStart
			rar.RARDtStop = DtStop
			rar.ContractRent = x
			m = append(m, rar)
		}
	}

	//-------------------------------------------------------------------
	// Notes
	//-------------------------------------------------------------------
	note := strings.TrimSpace(sa[Notes])
	if len(note) > 0 {
		var nl rlib.NoteList
		nl.BID = ra.BID
		nl.NLID, err = rlib.InsertNoteList(ctx, &nl)
		if err != nil {
			errMsg := fmt.Sprintf("error creating NoteList = %s", err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
		var n rlib.Note
		n.Comment = note
		n.NTID = 1 // first comment type
		n.BID = nl.BID
		n.NLID = nl.NLID
		_, err = rlib.InsertNote(ctx, &n)
		if err != nil {
			errMsg := fmt.Sprintf("error creating NoteList = %s", err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Notes, -1, errMsg)
		}
		ra.NLID = nl.NLID
	}

	//-------------------------------------------------------------------
	// look for any rental agreements already in existence that cover
	// the rentables referenced in this one...
	//-------------------------------------------------------------------
	for i := 0; i < len(m); i++ {
		// TODO(Steve): ignore error?
		rra, _ := rlib.GetAgreementsForRentable(ctx, m[i].RID, &ra.AgreementStart, &ra.AgreementStop)
		for j := 0; j < len(rra); j++ {
			errMsg := fmt.Sprintf("%s:: Rentable %s is already included in Rental Agreement %s from %s to %s",
				RentableAlreadyRented, rlib.IDtoString("R", rra[j].RID), rlib.IDtoString("RA", rra[j].RAID),
				rra[j].RARDtStart.Format(rlib.RRDATEFMT4), rra[j].RARDtStop.Format(rlib.RRDATEFMT4))
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
	}

	//-----------------------------------------------
	// Validate that we have at least one payor...
	//-----------------------------------------------
	if len(payors) == 0 {
		errMsg := fmt.Sprintf("No valid payors for this rental agreement")
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	//------------------------------------
	// Write the rental agreement record
	//-----------------------------------
	RAID, err := rlib.InsertRentalAgreement(ctx, &ra)
	if nil != err {
		errMsg := fmt.Sprintf("error inserting rlib.RentalAgreement = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	var lm rlib.LedgerMarker
	lm.Dt = ra.AgreementStart
	lm.RAID = ra.RAID
	lm.State = rlib.LMINITIAL
	_, err = rlib.InsertLedgerMarker(ctx, &lm)

	//------------------------------------------------------------
	// Add the rentables, and the users of those rentables...
	//------------------------------------------------------------
	for i := 0; i < len(m); i++ {
		m[i].RAID = RAID
		m[i].BID = ra.BID
		rlib.InsertRentalAgreementRentable(ctx, &m[i])
		//-----------------------------------------------------
		// Create a Rentable Ledger marker
		//-----------------------------------------------------
		var rlm = rlib.LedgerMarker{
			BID:     ra.BID,
			RAID:    RAID,
			RID:     m[i].RID,
			Dt:      m[i].RARDtStart,
			Balance: float64(0),
			State:   rlib.LMINITIAL,
		}
		_, err = rlib.InsertLedgerMarker(ctx, &rlm)
		if nil != err {
			errMsg := fmt.Sprintf("error inserting Rentable LedgerMarker = %v", err)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}

		//------------------------------
		// Add the users
		//------------------------------
		for j := 0; j < len(users); j++ {
			users[j].RID = m[i].RID
			users[j].BID = ra.BID
			_, err := rlib.InsertRentableUser(ctx, &users[j])
			if err != nil {
				errMsg := fmt.Sprintf("error inserting RentableUser = %v", err)
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
			}
		}
	}

	//------------------------------
	// Add the payors
	//------------------------------
	for i := 0; i < len(payors); i++ {
		payors[i].RAID = RAID
		payors[i].BID = ra.BID
		_, err := rlib.InsertRentalAgreementPayor(ctx, &payors[i])
		if err != nil {
			errMsg := fmt.Sprintf("error inserting RentablePayor = %v", err)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
	}
	return 0, nil
}

// LoadRentalAgreementCSV loads a csv file with rental specialty types and processes each one
func LoadRentalAgreementCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateRentalAgreement)
}
