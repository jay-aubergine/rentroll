package core

import (
	"context"
	"regexp"
	"rentroll/rlib"
	"strings"
)

// StringInSlice used to check whether string a
// is present or not in slice list
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// IntegerInSlice used to check whether int a
// is present or not in slice list
func IntegerInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// IsValidEmail used to check valid email or not
func IsValidEmail(email string) bool {
	// TODO: confirm which regex to use
	// Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	Re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+$`)
	return Re.MatchString(email)
}

// GetImportedCount get map of summaryCount as an argument
// then it hit db to get imported count for each type
func GetImportedCount(ctx context.Context, summaryCount map[int]map[string]int, BID int64) error {
	for dbType := range summaryCount {
		switch dbType {
		case DBCustomAttrRef:
			n, err := rlib.GetCountBusinessCustomAttrRefs(ctx, BID)
			if err != nil {
				return err
			}
			summaryCount[DBCustomAttrRef]["imported"] += n
			break
		case DBCustomAttr:
			n, err := rlib.GetCountBusinessCustomAttributes(ctx, BID)
			if err != nil {
				return err
			}
			summaryCount[DBCustomAttr]["imported"] += n
			break
		case DBRentableType:
			n, err := rlib.GetCountBusinessRentableTypes(ctx, BID)
			if err != nil {
				return err
			}
			summaryCount[DBRentableType]["imported"] += n
			break
		case DBPeople:
			n, err := rlib.GetCountBusinessTransactants(ctx, BID)
			if err != nil {
				return err
			}
			summaryCount[DBPeople]["imported"] += n
			break
		case DBRentable:
			n, err := rlib.GetCountBusinessRentables(ctx, BID)
			if err != nil {
				return err
			}
			summaryCount[DBRentable]["imported"] += n
			break
		case DBRentalAgreement:
			n, err := rlib.GetCountBusinessRentalAgreements(ctx, BID)
			if err != nil {
				return err
			}
			summaryCount[DBRentalAgreement]["imported"] += n
			break
		}
	}
	return nil
}

// DgtGrpSepToDgts converts separated group of digits string to
// plain digits string without any separator
// ex., 1,200,000 -> 1200000
func DgtGrpSepToDgts(dstr string) string {
	return strings.NewReplacer(",", "").Replace(dstr)
}
