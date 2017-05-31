package ws

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"rentroll/rlib"
	"strconv"
	"strings"
	"time"
)

// DepositoryGrid contains the data from Depository that is targeted to the UI Grid that displays
// a list of Depository structs
type DepositoryGrid struct {
	Recid       int64 `json:"recid"`
	DEPID       int64
	BID         int64
	BUD         rlib.XJSONBud
	LID         int64
	Name        string
	AccountNo   string
	LdgrName    string
	GLNumber    string
	LastModTime time.Time
	LastModBy   int64
}

// DepositorySearchResponse is a response string to the search request for Depository records
type DepositorySearchResponse struct {
	Status  string           `json:"status"`
	Total   int64            `json:"total"`
	Records []DepositoryGrid `json:"records"`
}

// DepositorySaveForm contains the data from Depository FORM that is targeted to the UI Form that displays
// a list of Depository structs
type DepositorySaveForm struct {
	Recid       int64 `json:"recid"`
	DEPID       int64
	BID         int64
	Name        string
	AccountNo   string
	LdgrName    string
	GLNumber    string
	LastModTime time.Time
	LastModBy   int64
}

// DepositoryGridSave is the input data format for a Save command
type DepositoryGridSave struct {
	Status   string             `json:"status"`
	Recid    int64              `json:"recid"`
	FormName string             `json:"name"`
	Record   DepositorySaveForm `json:"record"`
}

// DepositorySaveOther is a struct to handle the UI list box selections
type DepositorySaveOther struct {
	BUD rlib.W2uiHTMLSelect
	LID rlib.W2uiHTMLSelect
}

// SaveDepositoryOther is the input data format for the "other" data on the Save command
type SaveDepositoryOther struct {
	Status string              `json:"status"`
	Recid  int64               `json:"recid"`
	Name   string              `json:"name"`
	Record DepositorySaveOther `json:"record"`
}

// DepSaveOther is a struct to handle the UI list box selections
type DepSaveOther struct {
	BUD rlib.W2uiHTMLSelect
	LID rlib.W2uiHTMLSelect
}

// DepositoryGetResponse is the response to a GetDepository request
type DepositoryGetResponse struct {
	Status string         `json:"status"`
	Record DepositoryGrid `json:"record"`
}

// DeleteDepForm used to delete form
type DeleteDepForm struct {
	ID int64
}

// SvcHandlerDepository formats a complete data record for an assessment for use with the w2ui Form
// For this call, we expect the URI to contain the BID and the DEPID as follows:
//
// The server command can be:
//      get
//      save
//      delete
//-----------------------------------------------------------------------------------
func SvcHandlerDepository(w http.ResponseWriter, r *http.Request, d *ServiceData) {

	var (
		funcname = "SvcHandlerDepository"
		err      error
	)

	fmt.Printf("Entered %s\n", funcname)
	fmt.Printf("Request: %s:  BID = %d,  DEPID = %d\n", d.wsSearchReq.Cmd, d.BID, d.ID)

	switch d.wsSearchReq.Cmd {
	case "get":
		if d.ID <= 0 && d.wsSearchReq.Limit > 0 {
			SvcSearchHandlerDepositories(w, r, d) // it is a query for the grid.
		} else {
			if d.ID < 0 {
				err = fmt.Errorf("DepositoryID is required but was not specified")
				SvcGridErrorReturn(w, err, funcname)
				return
			}
			getDepository(w, r, d)
		}
		break
	case "save":
		saveDepository(w, r, d)
		break
	case "delete":
		deleteDepository(w, r, d)
	default:
		err = fmt.Errorf("Unhandled command: %s", d.wsSearchReq.Cmd)
		SvcGridErrorReturn(w, err, funcname)
		return
	}
}

// depGridRowScan scans a result from sql row and dump it in a PrARGrid struct
func depGridRowScan(rows *sql.Rows, q DepositoryGrid) (DepositoryGrid, error) {
	err := rows.Scan(&q.DEPID, &q.LID, &q.Name, &q.AccountNo, &q.LdgrName, &q.GLNumber, &q.LastModTime, &q.LastModBy)
	return q, err
}

var depSearchFieldMap = selectQueryFieldMap{
	"DEPID":       {"Depository.DEPID"},
	"LID":         {"Depository.LID"},
	"Name":        {"Depository.Name"},
	"AccountNo":   {"Depository.AccountNo"},
	"LdgrName":    {"GLAccount.Name"},
	"GLNumber":    {"GLAccount.GLNumber"},
	"LastModTime": {"Depository.LastModTime"},
	"LastModBy":   {"Depository.LastModBy"},
}

// which fields needs to be fetch to satisfy the struct
var depSearchSelectQueryFields = selectQueryFields{
	"Depository.DEPID",
	"Depository.LID",
	"Depository.Name",
	"Depository.AccountNo",
	"GLAccount.Name as LdgrName",
	"GLAccount.GLNumber",
	"Depository.LastModTime",
	"Depository.LastModBy",
}

// SvcSearchHandlerDepositories generates a report of all Depositories defined business d.BID
// wsdoc {
//  @Title  Search Depositories
//	@URL /v1/dep/:BUI
//  @Method  POST
//	@Synopsis Search Depositories
//  @Descr  Search all Depository and return those that match the Search Logic.
//  @Descr  The search criteria includes start and stop dates of interest.
//	@Input WebGridSearchRequest
//  @Response DepositorySearchResponse
// wsdoc }
func SvcSearchHandlerDepositories(w http.ResponseWriter, r *http.Request, d *ServiceData) {

	var (
		funcname = "SvcSearchHandlerDepositories"
		g        DepositorySearchResponse
		err      error
		order    = `Depository.DEPID ASC` // default ORDER in sql result
		whr      = fmt.Sprintf("Depository.BID=%d", d.BID)
	)
	fmt.Printf("Entered %s\n", funcname)

	// get where clause and order clause for sql query
	_, orderClause := GetSearchAndSortSQL(d, depSearchFieldMap)
	if len(orderClause) > 0 {
		order = orderClause
	}

	depSearchQuery := `
	SELECT
		{{.SelectClause}}
	FROM Depository
	LEFT JOIN GLAccount on GLAccount.LID=Depository.LID
	WHERE {{.WhereClause}}
	ORDER BY {{.OrderClause}}`

	qc := queryClauses{
		"SelectClause": strings.Join(depSearchSelectQueryFields, ","),
		"WhereClause":  whr,
		"OrderClause":  order,
	}

	// get TOTAL COUNT First
	countQuery := renderSQLQuery(depSearchQuery, qc)
	g.Total, err = GetQueryCount(countQuery, qc)
	if err != nil {
		fmt.Printf("%s: Error from GetQueryCount: %s\n", funcname, err.Error())
		SvcGridErrorReturn(w, err, funcname)
		return
	}
	fmt.Printf("g.Total = %d\n", g.Total)

	// FETCH the records WITH LIMIT AND OFFSET
	// limit the records to fetch from server, page by page
	limitAndOffsetClause := `
	LIMIT {{.LimitClause}}
	OFFSET {{.OffsetClause}};`

	// build query with limit and offset clause
	// if query ends with ';' then remove it
	depQueryWithLimit := depSearchQuery + limitAndOffsetClause

	// Add limit and offset value
	qc["LimitClause"] = strconv.Itoa(d.wsSearchReq.Limit)
	qc["OffsetClause"] = strconv.Itoa(d.wsSearchReq.Offset)

	// get formatted query with substitution of select, where, order clause
	qry := renderSQLQuery(depQueryWithLimit, qc)
	fmt.Printf("db query = %s\n", qry)

	rows, err := rlib.RRdb.Dbrr.Query(qry)
	if err != nil {
		fmt.Printf("%s: Error from DB Query: %s\n", funcname, err.Error())
		SvcGridErrorReturn(w, err, funcname)
		return
	}
	defer rows.Close()

	i := int64(d.wsSearchReq.Offset)
	count := 0
	for rows.Next() {
		var q DepositoryGrid
		q.Recid = i
		q.BID = d.BID
		q.BUD = getBUDFromBIDList(q.BID)

		q, err = depGridRowScan(rows, q)
		if err != nil {
			SvcGridErrorReturn(w, err, funcname)
			return
		}

		g.Records = append(g.Records, q)
		count++ // update the count only after adding the record
		if count >= d.wsSearchReq.Limit {
			break // if we've added the max number requested, then exit
		}
		i++
	}

	err = rows.Err()
	if err != nil {
		SvcGridErrorReturn(w, err, funcname)
		return
	}

	g.Status = "success"
	w.Header().Set("Content-Type", "application/json")
	SvcWriteResponse(&g, w)
}

// deleteDepository deletes a payment type from the database
// wsdoc {
//  @Title  Delete Depository
//	@URL /v1/dep/:BUI/:RAID
//  @Method  POST
//	@Synopsis Delete a Payment Type
//  @Desc  This service deletes a Depository.
//	@Input WebGridDelete
//  @Response SvcStatusResponse
// wsdoc }
func deleteDepository(w http.ResponseWriter, r *http.Request, d *ServiceData) {
	var (
		funcname = "deleteDepository"
	)
	fmt.Printf("Entered %s\n", funcname)
	fmt.Printf("record data = %s\n", d.data)

	var del DeleteDepForm
	if err := json.Unmarshal([]byte(d.data), &del); err != nil {
		e := fmt.Errorf("Error with json.Unmarshal:  %s", err.Error())
		SvcGridErrorReturn(w, e, funcname)
		return
	}

	if err := rlib.DeleteDepository(del.ID); err != nil {
		SvcGridErrorReturn(w, err, funcname)
		return
	}

	SvcWriteSuccessResponse(w)
}

// GetDepository returns the requested assessment
// wsdoc {
//  @Title  Save Depository
//	@URL /v1/dep/:BUI/:DEPID
//  @Method  GET
//	@Synopsis Update the information on a Depository with the supplied data
//  @Description  This service updates Depository :DEPID with the information supplied. All fields must be supplied.
//	@Input DepositoryGridSave
//  @Response SvcStatusResponse
// wsdoc }
func saveDepository(w http.ResponseWriter, r *http.Request, d *ServiceData) {

	var (
		funcname = "saveDepository"
		foo      DepositoryGridSave
		bar      SaveDepositoryOther
		err      error
	)

	fmt.Printf("Entered %s\n", funcname)
	fmt.Printf("record data = %s\n", d.data)

	// get data
	data := []byte(d.data)

	if err := json.Unmarshal(data, &foo); err != nil {
		e := fmt.Errorf("Error with json.Unmarshal:  %s", err.Error())
		SvcGridErrorReturn(w, e, funcname)
		return
	}

	if err := json.Unmarshal(data, &bar); err != nil {
		e := fmt.Errorf("Error with json.Unmarshal:  %s", err.Error())
		SvcGridErrorReturn(w, e, funcname)
		return
	}

	var a rlib.Depository
	rlib.MigrateStructVals(&foo.Record, &a) // the variables that don't need special handling

	var ok bool
	a.BID, ok = rlib.RRdb.BUDlist[bar.Record.BUD.ID]
	if !ok {
		e := fmt.Errorf("%s: Could not map BID value: %s", funcname, bar.Record.BUD.ID)
		rlib.Ulog("%s", e.Error())
		SvcGridErrorReturn(w, e, funcname)
		return
	}
	a.LID, ok = rlib.StringToInt64(bar.Record.LID.ID) // CreditLID has drop list
	if !ok {
		e := fmt.Errorf("%s: invalid LID value: %s", funcname, bar.Record.LID.ID)
		SvcGridErrorReturn(w, e, funcname)
		return
	}

	if a.DEPID == 0 && d.ID == 0 {
		// This is a new AR
		fmt.Printf(">>>> NEW DEPOSITORY IS BEING ADDED\n")
		_, err = rlib.InsertDepository(&a)
	} else {
		// update existing record
		fmt.Printf("Updating existing Depository: %d\n", a.DEPID)
		err = rlib.UpdateDepository(&a)
	}

	if err != nil {
		e := fmt.Errorf("%s: Error saving depository (DEPID=%d\n: %s", funcname, a.DEPID, err.Error())
		SvcGridErrorReturn(w, e, funcname)
		return
	}

	SvcWriteSuccessResponse(w)
}

// // depositoryUpdate unmarshals the supplied string. If Recid > 0 it updates the
// // Depository record using Recid as the DEPID.  If Recid == 0, then it inserts a
// // new Depository record.
// func depositoryUpdate(s string, d *ServiceData) error {
// 	var err error
// 	b := []byte(s)
// 	var rec DepositoryGrid
// 	if err = json.Unmarshal(b, &rec); err != nil { // first parse to determine the record ID we need to load
// 		return err
// 	}
// 	if rec.Recid > 0 { // is this an update?
// 		pt, err := rlib.GetDepository(rec.Recid) // now load that record...
// 		if err != nil {
// 			return err
// 		}
// 		if err = json.Unmarshal(b, &pt); err != nil { // merge in the changes...
// 			return err
// 		}
// 		return rlib.UpdateDepository(&pt) // and save the result
// 	}
// 	// no, it is a new table entry that has not been saved...
// 	var a rlib.Depository
// 	if err := json.Unmarshal(b, &a); err != nil { // merge in the changes...
// 		return err
// 	}
// 	a.BID = d.BID
// 	fmt.Printf("a = %#v\n", a)
// 	fmt.Printf(">>>> NEW DEPOSITORY IS BEING ADDED\n")
// 	_, err = rlib.InsertDepository(&a)
// 	return err
// }

// GetDepository returns the requested assessment
// wsdoc {
//  @Title  Get Depository
//	@URL /v1/dep/:BUI/:DEPID
//  @Method  GET
//	@Synopsis Get information on a Depository
//  @Description  Return all fields for assessment :DEPID
//	@Input WebGridSearchRequest
//  @Response DepositoryGetResponse
// wsdoc }
func getDepository(w http.ResponseWriter, r *http.Request, d *ServiceData) {

	var (
		funcname = "getDepository"
		g        DepositoryGetResponse
		whr      = fmt.Sprintf("Depository.DEPID=%d", d.ID)
	)

	fmt.Printf("entered %s\n", funcname)

	depQuery := `
	SELECT
		{{.SelectClause}}
	FROM Depository
	LEFT JOIN GLAccount on GLAccount.LID=Depository.LID
	WHERE {{.WhereClause}};`

	qc := queryClauses{
		"SelectClause": strings.Join(depSearchSelectQueryFields, ","),
		"WhereClause":  whr,
	}

	qry := renderSQLQuery(depQuery, qc)

	rows, err := rlib.RRdb.Dbrr.Query(qry)
	if err != nil {
		fmt.Printf("%s: Error from DB Query: %s\n", funcname, err.Error())
		SvcGridErrorReturn(w, err, funcname)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var q DepositoryGrid
		q.BID = d.BID
		q.BUD = getBUDFromBIDList(q.BID)

		q, err = depGridRowScan(rows, q)
		if err != nil {
			SvcGridErrorReturn(w, err, funcname)
			return
		}

		q.Recid = q.DEPID
		g.Record = q
	}
	err = rows.Err()
	if err != nil {
		SvcGridErrorReturn(w, err, funcname)
		return
	}

	g.Status = "success"
	SvcWriteResponse(&g, w)
}
