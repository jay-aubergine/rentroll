package rcsv

import (
	"context"
	"fmt"
	"rentroll/rlib"
	"strings"
)

// 0     1     2                     3         4      5       6,       7
// BUD,  RAID, Name,                 Type, Breed,    Color, Weight, DtStart, DtStop
// REX,  8,    Santa's Little Helper,Dog,  Greyhound,gray,  34.5,  2014-01-01,

// CreateRentalAgreementPetsFromCSV reads an assessment type string array and creates a database record for a pet
func CreateRentalAgreementPetsFromCSV(ctx context.Context, sa []string, lineno int) (int, error) {
	const funcname = "CreateRentalAgreementPetsFromCSV"
	var (
		err    error
		pet    rlib.RentalAgreementPet
		errmsg string
	)

	const (
		BUD    = 0
		RAID   = iota
		Name   = iota
		Type   = iota
		Breed  = iota
		Color  = iota
		Weight = iota
		Dt1    = iota
		Dt2    = iota
	)

	// csvCols is an array that defines all the columns that should be in this csv file
	var csvCols = []CSVColumn{
		{"BUD", BUD},
		{"RAID", RAID},
		{"Name", Name},
		{"Type", Type},
		{"Breed", Breed},
		{"Color", Color},
		{"Weight", Weight},
		{"DtStart", Dt1},
		{"DtStop", Dt2},
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
		pet.BID = b2.BID
	}

	//-------------------------------------------------------------------
	// Find Rental Agreement
	//-------------------------------------------------------------------
	pet.RAID = CSVLoaderGetRAID(sa[RAID])
	_, err = rlib.GetRentalAgreement(ctx, pet.RAID)
	if nil != err {
		errMsg := fmt.Sprintf("error loading Rental Agreement %s, err = %v", sa[RAID], err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, RAID, -1, errMsg)
	}

	pet.Name = strings.TrimSpace(sa[Name])
	pet.Type = strings.TrimSpace(sa[Type])
	pet.Breed = strings.TrimSpace(sa[Breed])
	pet.Color = strings.TrimSpace(sa[Color])

	//-------------------------------------------------------------------
	// Get the Weight
	//-------------------------------------------------------------------
	pet.Weight, errmsg = rlib.FloatFromString(sa[Weight], "Weight is invalid")
	if len(errmsg) > 0 {
		errMsg := fmt.Sprintf("Weight is invalid: %s  (%s)", sa[Weight], err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Weight, -1, errMsg)
	}

	//-------------------------------------------------------------------
	// Get the dates
	//-------------------------------------------------------------------
	DtStart, err := rlib.StringToDate(sa[Dt1])
	if err != nil {
		errMsg := fmt.Sprintf("invalid start date:  %s", sa[Dt1])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Dt1, -1, errMsg)
	}
	pet.DtStart = DtStart

	end := "9999-01-01"
	if len(sa) > Dt2 {
		if len(sa[Dt2]) > 0 {
			end = sa[Dt2]
		}
	}
	DtStop, err := rlib.StringToDate(end)
	if err != nil {
		errMsg := fmt.Sprintf("invalid stop date:  %s", sa[Dt2])
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, Dt2, -1, errMsg)
	}
	pet.DtStop = DtStop

	_, err = rlib.InsertRentalAgreementPet(ctx, &pet)
	if nil != err {
		errMsg := fmt.Sprintf("Could not save pet, err = %v", err)
		return CsvErrorSensitivity, formatCSVErrors(funcname, lineno, -1, -1, errMsg)
	}
	return 0, nil
}

// LoadPetsCSV loads a csv file with a chart of accounts and creates rlib.GLAccount markers for each
func LoadPetsCSV(ctx context.Context, fname string) []error {
	return LoadRentRollCSV(ctx, fname, CreateRentalAgreementPetsFromCSV)
}
