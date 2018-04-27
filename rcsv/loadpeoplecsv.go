package rcsv

import (
	"context"
	"fmt"
	"rentroll/bizlogic"
	"rentroll/rlib"
	"strconv"
	"strings"
	"time"
)

// BUD et all are constants used by multiple programs
// for the column headings in csv files.
const (
	BUD                       = 0
	FirstName                 = iota
	MiddleName                = iota
	LastName                  = iota
	CompanyName               = iota
	IsCompany                 = iota
	PrimaryEmail              = iota
	SecondaryEmail            = iota
	WorkPhone                 = iota
	CellPhone                 = iota
	Address                   = iota
	Address2                  = iota
	City                      = iota
	State                     = iota
	PostalCode                = iota
	Country                   = iota
	Points                    = iota
	AccountRep                = iota
	DateofBirth               = iota
	EmergencyContactName      = iota
	EmergencyContactAddress   = iota
	EmergencyContactTelephone = iota
	EmergencyEmail            = iota
	AlternateAddress          = iota
	EligibleFutureUser        = iota
	Industry                  = iota
	SourceSLSID               = iota
	CreditLimit               = iota
	TaxpayorID                = iota
	EmployerName              = iota
	EmployerStreetAddress     = iota
	EmployerCity              = iota
	EmployerState             = iota
	EmployerPostalCode        = iota
	EmployerEmail             = iota
	EmployerPhone             = iota
	Occupation                = iota
	ApplicationFee            = iota
	Notes                     = iota
	DesiredUsageStartDate     = iota
	RentableTypePreference    = iota
	Approver                  = iota
	DeclineReasonSLSID        = iota
	OtherPreferences          = iota
	FollowUpDate              = iota
	CSAgent                   = iota
	OutcomeSLSID              = iota
	FloatingDeposit           = iota
	RAID                      = iota
)

// csvCols is an array that defines all the columns that should be in this csv file
var csvCols = []CSVColumn{
	{"BUD", BUD},
	{"FirstName", FirstName},
	{"MiddleName", MiddleName},
	{"LastName", LastName},
	{"CompanyName", CompanyName},
	{"IsCompany", IsCompany},
	{"PrimaryEmail", PrimaryEmail},
	{"SecondaryEmail", SecondaryEmail},
	{"WorkPhone", WorkPhone},
	{"CellPhone", CellPhone},
	{"Address", Address},
	{"Address2", Address2},
	{"City", City},
	{"State", State},
	{"PostalCode", PostalCode},
	{"Country", Country},
	{"Points", Points},
	{"AccountRep", AccountRep},
	{"DateofBirth", DateofBirth},
	{"EmergencyContactName", EmergencyContactName},
	{"EmergencyContactAddress", EmergencyContactAddress},
	{"EmergencyContactTelephone", EmergencyContactTelephone},
	{"EmergencyEmail", EmergencyEmail},
	{"AlternateAddress", AlternateAddress},
	{"EligibleFutureUser", EligibleFutureUser},
	{"Industry", Industry},
	{"SourceSLSID", SourceSLSID},
	{"CreditLimit", CreditLimit},
	{"TaxpayorID", TaxpayorID},
	{"EmployerName", EmployerName},
	{"EmployerStreetAddress", EmployerStreetAddress},
	{"EmployerCity", EmployerCity},
	{"EmployerState", EmployerState},
	{"EmployerPostalCode", EmployerPostalCode},
	{"EmployerEmail", EmployerEmail},
	{"EmployerPhone", EmployerPhone},
	{"Occupation", Occupation},
	{"ApplicationFee", ApplicationFee},
	{"Notes", Notes},
	{"DesiredUsageStartDate", DesiredUsageStartDate},
	{"RentableTypePreference", RentableTypePreference},
	{"Approver", Approver},
	{"DeclineReasonSLSID", DeclineReasonSLSID},
	{"OtherPreferences", OtherPreferences},
	{"FollowUpDate", FollowUpDate},
	{"CSAgent", CSAgent},
	{"OutcomeSLSID", OutcomeSLSID},
	{"FloatingDeposit", FloatingDeposit},
	{"RAID", RAID},
}

func rcsvCopyString(p *string, s string) error {
	*p = s
	return nil
}

// CSV file format:
//  |<------------------------------------------------------------------  TRANSACTANT ----------------------------------------------------------------------------->|  |<-------------------------------------------------------------------------------------------------------------  rlib.User  ------------------------------------------------------------------------------------------------------------------------------------------------------------------------>|<----------------------------------------------------------------------------- rlib.Payor ------------------------------------------------->|
//   0   1          2           3         4            5          6             7               8          9          10       11        12    13     14          15       16      17       18        19        20       21                 22                  23                   24          25           26                    27                       28                         29              30                31                          32        33            34           35         36            37                     38            39             40                  41             42             43          44             45    46                     47                      48        49                  50                51            52       53            54               55
// 	BUD, FirstName, MiddleName, LastName, CompanyName, IsCompany, PrimaryEmail, SecondaryEmail, WorkPhone, CellPhone, Address, Address2, City, State, PostalCode, Country, Points, VehicleMake, VehicleModel, VehicleColor, VehicleYear, LicensePlateState, LicensePlateNumber, ParkingPermitNumber, AccountRep, DateofBirth, EmergencyContactName, EmergencyContactAddress, EmergencyContactTelephone, EmergencyEmail, AlternateAddress, EligibleFutureUser, Industry, SourceSLSID, CreditLimit, TaxpayorID, EmployerName, EmployerStreetAddress, EmployerCity, EmployerState, EmployerPostalCode, EmployerEmail, EmployerPhone, Occupation, ApplicationFee,Notes,DesiredUsageStartDate, RentableTypePreference, Approver, DeclineReasonSLSID, OtherPreferences, FollowUpDate, CSAgent, OutcomeSLSID, FloatingDeposit, RAID
// 	Edna,,Krabappel,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
// 	Ned,,Flanders,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
// 	Moe,,Szyslak,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
// 	Montgomery,,Burns,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
// 	Nelson,,Muntz,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
// 	Milhouse,,Van Houten,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
// 	Clancey,,Wiggum,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
// 	Homer,J,Simpson,homerj@springfield.com,,408-654-8732,,744 Evergreen Terrace,,Springfield,MO,64001,USA,5987,,Canyonero,red,,MO,BR549,,,,Marge Simpson,744 Evergreen Terrace,654=183-7946,,,,,,,,,,,,,,,,,"note: Homer is an idiot"

// CreatePeopleFromCSV reads a rental specialty type string array and creates a database record for the rental specialty type.
//
// Return Values
// int   -->  0 = everything is fine, process the next line;  1 abort the csv load
// error -->  nil if no problems
func CreatePeopleFromCSV(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreatePeopleFromCSV"

	var (
		err      error
		tr       rlib.Transactant
		t        rlib.User
		p        rlib.Payor
		pr       rlib.Prospect
		x        float64
		userNote string
	)

	var rcsvPeopleHandlers = []struct {
		ID      int
		Handler func(*string, string) error
		p       *string
	}{
		{BUD, nil, nil},
		{FirstName, rcsvCopyString, &tr.FirstName},
		{MiddleName, rcsvCopyString, &tr.MiddleName},
		{LastName, rcsvCopyString, &tr.LastName},
		{CompanyName, rcsvCopyString, &tr.CompanyName},
		{IsCompany, nil, nil},
		{PrimaryEmail, rcsvCopyString, &tr.PrimaryEmail},
		{SecondaryEmail, rcsvCopyString, &tr.SecondaryEmail},
		{WorkPhone, rcsvCopyString, &tr.WorkPhone},
		{CellPhone, nil, nil},
		{Address, rcsvCopyString, &tr.Address},
		{Address2, rcsvCopyString, &tr.Address2},
		{City, rcsvCopyString, &tr.City},
		{State, rcsvCopyString, &tr.State},
		{PostalCode, rcsvCopyString, &tr.PostalCode},
		{Country, rcsvCopyString, &tr.Country},
		{Points, nil, nil},
		{AccountRep, nil, nil},
		{DateofBirth, nil, nil},
		{EmergencyContactName, rcsvCopyString, &t.EmergencyContactName},
		{EmergencyContactAddress, rcsvCopyString, &t.EmergencyContactAddress},
		{EmergencyContactTelephone, rcsvCopyString, &t.EmergencyContactTelephone},
		{EmergencyEmail, rcsvCopyString, &t.EmergencyEmail},
		{AlternateAddress, rcsvCopyString, &t.AlternateAddress},
		{EligibleFutureUser, nil, nil},
		{Industry, rcsvCopyString, &t.Industry},
		{SourceSLSID, nil, nil},
		{CreditLimit, nil, nil},
		{TaxpayorID, rcsvCopyString, &p.TaxpayorID},
		{EmployerName, rcsvCopyString, &pr.EmployerName},
		{EmployerStreetAddress, rcsvCopyString, &pr.EmployerStreetAddress},
		{EmployerCity, rcsvCopyString, &pr.EmployerCity},
		{EmployerState, rcsvCopyString, &pr.EmployerState},
		{EmployerPostalCode, rcsvCopyString, &pr.EmployerPostalCode},
		{EmployerEmail, rcsvCopyString, &pr.EmployerEmail},
		{EmployerPhone, rcsvCopyString, &pr.EmployerPhone},
		{Occupation, rcsvCopyString, &pr.Occupation},
		{ApplicationFee, nil, nil},
		{Notes, nil, nil},
		{DesiredUsageStartDate, nil, nil},
		{RentableTypePreference, nil, nil},
		{Approver, nil, nil},
		{DeclineReasonSLSID, nil, nil},
		{OtherPreferences, nil, nil},
		{FollowUpDate, nil, nil},
		{CSAgent, nil, nil},
		{OutcomeSLSID, nil, nil},
		{FloatingDeposit, nil, nil},
		{RAID, nil, nil},
	}

	ignoreDupPhone := false

	y, err := ValidateCSVColumnsErr(csvCols, sa, funcname, lineno)
	if y {
		return 1, err
	}
	if lineno == 1 {
		return 0, nil // we've validated the col headings, all is good, send the next line
	}

	dateform := "2006-01-02"
	pr.OtherPreferences = ""

	for i := 0; i < len(sa); i++ {
		s := strings.TrimSpace(sa[i])
		if rcsvPeopleHandlers[i].p != nil {
			rcsvPeopleHandlers[i].Handler(rcsvPeopleHandlers[i].p, s)
			continue
		}
		switch i {
		case BUD: // business
			des := strings.ToLower(strings.TrimSpace(sa[0])) // this should be BUD

			//-------------------------------------------------------------------
			// Make sure the rlib.Business is in the database
			//-------------------------------------------------------------------
			if len(des) > 0 { // make sure it's not empty
				b1, err := rlib.GetBusinessByDesignation(ctx, des) // see if we can find the biz
				if err != nil {
					errMsg := fmt.Sprintf("error while getting business by designation(%s), error: %s", sa[BUD], err.Error())
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
				}
				if len(b1.Designation) == 0 {
					errMsg := fmt.Sprintf("Business with designation %s does not exist", sa[BUD])
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
				}
				tr.BID = b1.BID
			}
		case IsCompany:
			if len(s) > 0 {
				ic, err := rlib.YesNoToInt(s)
				if err != nil {
					errMsg := fmt.Sprintf("IsCompany value is invalid: %s", s)
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, IsCompany, -1, errMsg)
				}
				tr.IsCompany = int64(ic)
			}
		case CellPhone:
			if len(s) > 0 && s[0] == '*' {
				s = s[1:]
				ignoreDupPhone = true
			}
			tr.CellPhone = s
		case Points:
			if len(s) > 0 {
				i, err := strconv.Atoi(strings.TrimSpace(s))
				if err != nil {
					errMsg := fmt.Sprintf("Points value is invalid: %s", s)
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Points, -1, errMsg)
				}
				t.Points = int64(i)
			}
		case AccountRep:
			if len(s) > 0 {
				i, err := strconv.Atoi(strings.TrimSpace(s))
				if err != nil {
					errMsg := fmt.Sprintf("AccountRep value is invalid: %s", s)
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, AccountRep, -1, errMsg)
				}
				p.AccountRep = int64(i)
			}
		case DateofBirth:
			if len(s) > 0 {
				t.DateofBirth, err = rlib.StringToDate(s)
				if err != nil {
					errMsg := fmt.Sprintf("Bad date of birth: %s, error = %s", s, err.Error())
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, DateofBirth, -1, errMsg)
				}
			}
		case EligibleFutureUser:
			if len(s) > 0 {
				var err error
				t.EligibleFutureUser, err = rlib.YesNoToInt(s)
				if err != nil {
					errMsg := fmt.Sprintf(err.Error())
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, EligibleFutureUser, -1, errMsg)
				}
			}
		case SourceSLSID:
			if len(s) > 0 {
				var y int64
				if y, err = strconv.ParseInt(strings.TrimSpace(s), 10, 64); err != nil {
					errMsg := fmt.Sprintf("Invalid SourceSLSID value: %s", s)
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, SourceSLSID, -1, errMsg)
				}
				t.SourceSLSID = y
			}
		case CreditLimit:
			if len(s) > 0 {
				if x, err = strconv.ParseFloat(strings.TrimSpace(s), 64); err != nil {
					errMsg := fmt.Sprintf("Invalid Credit Limit value: %s", s)
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, CreditLimit, -1, errMsg)
				}
				p.CreditLimit = x
			}
		case ApplicationFee:
			if len(s) > 0 {
				if x, err = strconv.ParseFloat(strings.TrimSpace(s), 64); err != nil {
					errMsg := fmt.Sprintf("Invalid ApplicationFee value: %s", s)
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, ApplicationFee, -1, errMsg)
				}
				pr.ApplicationFee = x
			}
		case Notes:
			if len(s) > 0 {
				userNote = s
			}
		case DesiredUsageStartDate:
			if len(s) > 0 {
				pr.DesiredUsageStartDate, err = rlib.StringToDate(s)
				if err != nil {
					errMsg := fmt.Sprintf("Invalid DesiredUsageStartDate value: %s", s)
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, DesiredUsageStartDate, -1, errMsg)
				}
			}
		case RentableTypePreference:
			if len(s) > 0 {
				rt, err := rlib.GetRentableTypeByStyle(ctx, s, tr.BID)
				if err != nil || rt.RTID == 0 {
					errMsg := fmt.Sprintf("Invalid DesiredUsageStartDate value: %s", s)
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RentableTypePreference, -1, errMsg)
				}
				pr.RentableTypePreference = rt.RTID
			}
		case Approver: // Approver ID
			if len(s) > 0 {
				var y int64
				if y, err = strconv.ParseInt(strings.TrimSpace(s), 10, 64); err != nil {
					errMsg := fmt.Sprintf("Invalid Approver UID value: %s", s)
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Approver, -1, errMsg)
				}
				pr.Approver = y
			}
		case DeclineReasonSLSID:
			if len(s) > 0 {
				var y int64
				if y, err = strconv.ParseInt(strings.TrimSpace(s), 10, 64); err != nil {
					errMsg := fmt.Sprintf("Invalid DeclineReasonSLSID value: %s", s)
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, DeclineReasonSLSID, -1, errMsg)
				}
				pr.DeclineReasonSLSID = y
			}
		case OtherPreferences:
			if len(s) > 0 {
				pr.OtherPreferences = s
			}
		case FollowUpDate:
			if len(s) > 0 {
				pr.FollowUpDate, _ = time.Parse(dateform, s)
			}
		case CSAgent:
			if len(s) > 0 {
				var y int64
				if y, err = strconv.ParseInt(strings.TrimSpace(s), 10, 64); err != nil {
					errMsg := fmt.Sprintf("Invalid CSAgent ID value: %s", s)
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, CSAgent, -1, errMsg)
				}
				pr.CSAgent = y
			}
		case OutcomeSLSID:
			if len(s) > 0 {
				var y int64
				if y, err = strconv.ParseInt(strings.TrimSpace(s), 10, 64); err != nil {
					errMsg := fmt.Sprintf("Invalid OutcomeSLSID value: %s", s)
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, OutcomeSLSID, -1, errMsg)
				}
				pr.OutcomeSLSID = y
			}

		case FloatingDeposit:
			if len(s) > 0 {
				if x, err = strconv.ParseFloat(strings.TrimSpace(s), 64); err != nil {
					errMsg := fmt.Sprintf("Invalid FloatingDeposit value: %s", s)
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, FloatingDeposit, -1, errMsg)
				}
				pr.FloatingDeposit = x
			}
		case RAID:
			if len(s) > 0 {
				var y int64
				if y, err = strconv.ParseInt(strings.TrimSpace(s), 10, 64); err != nil {
					errMsg := fmt.Sprintf("Invalid RAID value: %s", s)
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RAID, -1, errMsg)
				}
				pr.RAID = y
			}
		default:
			errMsg := fmt.Sprintf("Unknown field, column %s", s)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, i, -1, errMsg)
		}
	}

	//-------------------------------------------------------------------
	// Make sure BID is not 0
	//-------------------------------------------------------------------
	if tr.BID == 0 {
		errMsg := fmt.Sprintf("No Business found for BUD = %s", sa[BUD])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// Make sure this person doesn't already exist...
	//-------------------------------------------------------------------
	if len(tr.PrimaryEmail) > 0 {
		t1, err := rlib.GetTransactantByPhoneOrEmail(ctx, tr.BID, tr.PrimaryEmail)
		if err != nil { // if not "no rows error" then MUST return
			errMsg := fmt.Sprintf("Error while verifying Transactant with PrimaryEmail address = %s: %s", tr.PrimaryEmail, err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
		if t1.TCID > 0 {
			errMsg := fmt.Sprintf("%s:: Transactant with PrimaryEmail address = %s", DupTransactant, tr.PrimaryEmail)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
	}
	if len(tr.CellPhone) > 0 && !ignoreDupPhone {
		t1, err := rlib.GetTransactantByPhoneOrEmail(ctx, tr.BID, tr.CellPhone)
		if err != nil { // if not "no rows error" then MUST return
			errMsg := fmt.Sprintf("Error while verifying Transactant with CellPhone number = %s: %s", tr.CellPhone, err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
		if t1.TCID > 0 {
			errMsg := fmt.Sprintf("%s:: Transactant with CellPhone number = %s already exists", DupTransactant, tr.CellPhone)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
	}

	//-------------------------------------------------------------------
	// Make sure there's a name... if it's not a Company, then it needs
	// a first & last name.  If it is a company, then it needs a Company
	// name.
	//-------------------------------------------------------------------
	if tr.IsCompany == 0 && len(tr.FirstName) == 0 && len(tr.LastName) == 0 {
		errMsg := fmt.Sprintf("FirstName and LastName are required for a person")
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	if tr.IsCompany == 1 && len(tr.CompanyName) == 0 {
		errMsg := fmt.Sprintf("CompanyName is required for a company")
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// If there's a notelist, create it now...
	//-------------------------------------------------------------------
	if len(userNote) > 0 {
		var nl rlib.NoteList
		nl.BID = tr.BID
		nl.NLID, err = rlib.InsertNoteList(ctx, &nl)
		if err != nil {
			errMsg := fmt.Sprintf("error creating NoteList = %s", err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
		var n rlib.Note
		n.Comment = userNote
		n.NTID = 1 // first comment type
		n.NLID = nl.NLID
		n.BID = nl.BID
		_, err = rlib.InsertNote(ctx, &n)
		if err != nil {
			errMsg := fmt.Sprintf("error creating NoteList = %s", err.Error())
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
		tr.NLID = nl.NLID // start a notelist for this transactant
	}

	//-------------------------------------------------------------------
	// OK, just insert the records and we're done
	//-------------------------------------------------------------------
	tcid, err := rlib.InsertTransactant(ctx, &tr)
	if nil != err {
		errMsg := fmt.Sprintf("error inserting Transactant = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	tr.TCID = tcid
	t.TCID = tcid
	t.BID = tr.BID
	p.TCID = tcid
	p.BID = tr.BID
	pr.TCID = tcid
	pr.BID = tr.BID

	if tcid == 0 {
		errMsg := fmt.Sprintf("after InsertTransactant tcid = %d", tcid)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	// fmt.Printf("tcid = %d\n", tcid)
	// fmt.Printf("inserting user = %#v\n", t)
	_, err = rlib.InsertUser(ctx, &t)
	if nil != err {
		errMsg := fmt.Sprintf("error inserting rlib.User = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	_, err = rlib.InsertPayor(ctx, &p)
	if nil != err {
		errMsg := fmt.Sprintf("error inserting rlib.Payor = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	_, err = rlib.InsertProspect(ctx, &pr)
	if nil != err {
		errMsg := fmt.Sprintf("error inserting rlib.Prospect = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	errlist := bizlogic.FinalizeTransactant(ctx, &tr)
	if len(errlist) > 0 {
		errMsg := fmt.Sprintf("error inserting Transactant LedgerMarker = %s", errlist[0].Message)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	return 0, nil
}

// LoadPeopleCSV loads a csv file with people information
func LoadPeopleCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreatePeopleFromCSV)
}
