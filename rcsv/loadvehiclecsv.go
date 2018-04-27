package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strconv"
	"strings"
)

// CSV file format:
//   0   1     2        3         4         5        6                  7                   8
// 	BUD, TCID, VehicleMake, VehicleModel, VehicleColor, VehicleYear, LicensePlateState, LicensePlateNumber, ParkingPermitNumber
// 	REX, 1
// 	REX, 1
// 	REX, 1
// 	REX, 1
// 	REX, 1
// 	REX, 1
// 	REX, 1

// CreateVehicleFromCSV reads a rental specialty type string array and creates a database record for the rental specialty type.
// If the return value is not 0, abort the csv load
func CreateVehicleFromCSV(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateVehicleFromCSV"

	var (
		err error
		tr  rlib.Transactant
		t   rlib.Vehicle
	)

	const (
		BUD                 = 0
		TCID                = iota
		VehicleType         = iota
		VehicleMake         = iota
		VehicleModel        = iota
		VehicleColor        = iota
		VehicleYear         = iota
		LicensePlateState   = iota
		LicensePlateNumber  = iota
		ParkingPermitNumber = iota
		DtStart             = iota
		DtStop              = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"User", TCID},
		{"VehicleType", VehicleType},
		{"VehicleMake", VehicleMake},
		{"VehicleModel", VehicleModel},
		{"VehicleColor", VehicleColor},
		{"VehicleYear", VehicleYear},
		{"LicensePlateState", LicensePlateState},
		{"LicensePlateNumber", LicensePlateNumber},
		{"ParkingPermitNumber", ParkingPermitNumber},
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

	for i := 0; i < len(sa); i++ {
		s := strings.TrimSpace(sa[i])
		switch i {
		case BUD: // business
			des := strings.ToLower(strings.TrimSpace(sa[0])) // this should be BUD

			//-------------------------------------------------------------------
			// Make sure the rlib.Business is in the database
			//-------------------------------------------------------------------
			if len(des) > 0 { // make sure it's not empty
				b1, err := rlib.GetBusinessByDesignation(ctx, des) // see if we can find the biz
				if err != nil {
					errMsg := fmt.Sprintf("error while getting business by designation(%s): %s", sa[BUD], err.Error())
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
				}
				if len(b1.Designation) == 0 {
					errMsg := fmt.Sprintf("Business with designation %s does not exist", sa[BUD])
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, BUD, -1, errMsg)
				}
				tr.BID = b1.BID
			}
		case TCID:
			tr, err = rlib.GetTransactantByPhoneOrEmail(ctx, tr.BID, s)
			if err != nil {
				errMsg := fmt.Sprintf("error getting Transactant with %s listed as a phone or email: %s", sa[TCID], err.Error())
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, TCID, -1, errMsg)
			}
			if tr.TCID < 1 {
				errMsg := fmt.Sprintf("no Transactant found with %s listed as a phone or email", sa[TCID])
				return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, TCID, -1, errMsg)
			}
			t.TCID = tr.TCID
		case VehicleType:
			t.VehicleType = s
		case VehicleMake:
			t.VehicleMake = s
		case VehicleModel:
			t.VehicleModel = s
		case VehicleColor:
			t.VehicleColor = s
		case VehicleYear:
			if len(s) > 0 {
				i, err := strconv.Atoi(strings.TrimSpace(s))
				if err != nil {
					errMsg := fmt.Sprintf("VehicleYear value is invalid: %s", sa[VehicleYear])
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, VehicleYear, -1, errMsg)
				}
				t.VehicleYear = int64(i)
			}
		case LicensePlateState:
			t.LicensePlateState = s
		case LicensePlateNumber:
			t.LicensePlateNumber = s
		case ParkingPermitNumber:
			t.ParkingPermitNumber = s
		case DtStart:
			if len(s) > 0 {
				t.DtStart, err = rlib.StringToDate(s) // required field
				if err != nil {
					errMsg := fmt.Sprintf("invalid start date.  Error = %s", err.Error())
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, DtStart, -1, errMsg)
				}
			}
		case DtStop:
			if len(s) > 0 {
				t.DtStop, err = rlib.StringToDate(s) // required field
				if err != nil {
					errMsg := fmt.Sprintf("invalid stop date.  Error = %s", err.Error())
					return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, DtStop, -1, errMsg)
				}
			}
		default:
			errMsg := fmt.Sprintf("i = %d, unknown field", i)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, i, -1, errMsg)
		}
	}

	//-------------------------------------------------------------------
	// Check for duplicate...
	//-------------------------------------------------------------------
	// TODO(Steve): ignore error?
	tm, _ := rlib.GetVehiclesByLicensePlate(ctx, t.LicensePlateNumber)
	for i := 0; i < len(tm); i++ {
		if t.LicensePlateNumber == tm[i].LicensePlateNumber && t.LicensePlateState == tm[i].LicensePlateState {
			errMsg := fmt.Sprintf("vehicle with License Plate %s in State = %s already exists", t.LicensePlateNumber, t.LicensePlateState)
			return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
		}
	}

	//-------------------------------------------------------------------
	// OK, just insert the records and we're done
	//-------------------------------------------------------------------
	t.TCID = tr.TCID
	t.BID = tr.BID
	vid, err := rlib.InsertVehicle(ctx, &t)
	if nil != err {
		errMsg := fmt.Sprintf("error inserting Vehicle = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}

	if vid == 0 {
		errMsg := fmt.Sprintf("after InsertVehicle vid = %d", vid)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	return 0, nil
}

// LoadVehicleCSV loads a csv file with vehicles
func LoadVehicleCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateVehicleFromCSV)
}
