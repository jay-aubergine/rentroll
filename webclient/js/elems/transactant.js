"use strict";
window.getTransactantInitRecord = function (BID, BUD) {
    var y = new Date();

    return {
        recid: 0,
        FirstName: "",
        LastName: "",
        MiddleName: "",
        PreferredName: "",
        PrimaryEmail: "",
        TCID: 0,
        BID: BID,
        BUD: BUD,
        NLID: 0,
        CompanyName: "",
        IsCompany: 0,
        SecondaryEmail: "",
        WorkPhone: "",
        CellPhone: "",
        Address: "",
        Address2: "",
        City: "",
        State: "",
        PostalCode: "",
        Country: "",
        Website: "",
        LastModTime: y.toISOString(),
        LastModBy: 0,
        Points: 0,
        DateofBirth: "1/1/1900",
        EmergencyContactName: "",
        EmergencyContactAddress: "",
        EmergencyContactTelephone: "",
        EmergencyEmail: "",
        AlternateAddress: "",
        EligibleFutureUser: "yes",
        Industry: "",
        SourceSLSID: 0,
        CreditLimit: 0.00,
        TaxpayorID: "",
        AccountRep: 0,
        EligibleFuturePayor: "yes",
        EmployerName: "",
        EmployerStreetAddress: "",
        EmployerCity: "",
        EmployerState: "",
        EmployerPostalCode: "",
        EmployerEmail: "",
        EmployerPhone: "",
        Occupation: "",
        ApplicationFee: 0.00,
        DesiredUsageStartDate: "1/1/1900",
        RentableTypePreference: 0,
        FLAGS: 0,
        Approver: 0,
        DeclineReasonSLSID: 0,
        OtherPreferences: "",
        FollowUpDate: "1/1/1900",
        CSAgent: 0,
        OutcomeSLSID: 0,
        FloatingDeposit: 0.00,
        RAID: 0,
    };
};


window.buildTransactElements = function() {

//------------------------------------------------------------------------
//          transactantsGrid
//------------------------------------------------------------------------
$().w2grid({
    name: 'transactantsGrid',
    url: '/v1/transactants',
    multiSelect: false,
    show: {
        header: false,
        toolbar: true,
        footer: true,
        toolbarAdd: true,
        lineNumbers: false,
        selectColumn: false,
        expandColumn: false
    },
    columns: [
        {field: 'TCID',         caption: "TCID",          size: '50px',  sortable: true, style: 'text-align: right', hidden: false},
        {field: 'FirstName',    caption: "First Name",    size: '125px', sortable: true, hidden: false},
        {field: 'MiddleName',   caption: "Middle Name",   size: '20px',  sortable: true, hidden: true},
        {field: 'LastName',     caption: "Last Name",     size: '125px', sortable: true, hidden: false,
            render: function (record) {
                var s = '';
                if (typeof record === "undefined") {
                    return;
                }
                if (record.IsCompany === 0) {
                    s += '<span style="color:#999;font-size:16px"><i class="far fa-handshake" aria-hidden="true"></i></span>';
                }
                return s + ' ' + record.LastName;
            }
        },
        {field: 'CompanyName',  caption: "Company Name",  size: '125px', sortable: true, hidden: false,
            render: function (record) {
                var s = '';
                if (typeof record === "undefined") {
                    return;
                }
                if (record.IsCompany > 0) {
                    s += '<span style="color:#999;font-size:16px"><i class="far fa-handshake" aria-hidden="true"></i></span>';
                }
                return s + ' ' + record.CompanyName;
            }
        },
        {field: 'PrimaryEmail', caption: "Primary Email", size: '175px', sortable: true, hidden: false},
        {field: 'CellPhone',    caption: "Cell Phone",    size: '100px', sortable: true, hidden: false},
        {field: 'WorkPhone',    caption: "Work Phone",    size: '100px', sortable: true, hidden: false},
    ],
    onRefresh: function(event) {
        event.onComplete = function() {
            if (app.active_grid == this.name) {
                if (app.new_form_rec) {
                    this.selectNone();
                }
                else{
                    this.select(app.last.grid_sel_recid);
                }
            }
        };
    },
    onClick: function(event) {
        event.onComplete = function () {
            var yes_args = [this, event.recid],
                no_args = [this],
                no_callBack = function(grid) {
                    grid.select(app.last.grid_sel_recid);
                    return false;
                },
                yes_callBack = function(grid, recid) {
                    app.last.grid_sel_recid = parseInt(recid);
                    // keep highlighting current row in any case
                    grid.select(app.last.grid_sel_recid);
                    var rec = grid.get(recid);
                    setToForm('transactantForm', '/v1/person/' + rec.BID + '/' + rec.TCID, 700, true);
                };

            // warn user if form content has been changed
            form_dirty_alert(yes_callBack, no_callBack, yes_args, no_args);
         };
    },
    onAdd: function(/*event*/) {
        var yes_args = [this],
            no_callBack = function() { return false; },
            yes_callBack = function(grid) {
                // reset it
                app.last.grid_sel_recid = -1;
                grid.selectNone();

                // insert an empty record....
                var x = getCurrentBusiness();
                var BID=parseInt(x.value);
                var BUD = getBUDfromBID(BID);

                var record = getTransactantInitRecord(BID, BUD);
                w2ui.transactantForm.record = record;
                w2ui.transactantForm.refresh();
                setToForm('transactantForm', '/v1/person/' + BID + '/0', 700);
            };

        // warn user if form content has been changed
        form_dirty_alert(yes_callBack, no_callBack, yes_args);
    },
});


    //------------------------------------------------------------------------
    //          transactantForm
    //------------------------------------------------------------------------
    $().w2form({
        name: 'transactantForm',
        style: 'border: 0px; background-color: transparent;',
        header: app.sTransactant + ' Detail',
        url: '/v1/person',
        formURL: '/webclient/html/formtc.html',
        fields: [
            { field: 'recid', type: 'int', required: false, html: { page: 0, column: 0 } },
            { field: 'FirstName', type: 'text', required: false, html: { page: 0, column: 0 } },
            { field: 'LastName', type: 'text', required: false, html: { page: 0, column: 0 } },
            { field: 'MiddleName', type: 'text', required: false, html: { page: 0, column: 0 } },
            { field: 'PreferredName', type: 'text', required: false, html: { page: 0, column: 0 } },
            { field: 'PrimaryEmail', type: 'email', required: false, html: { page: 0, column: 0 } },
            { field: 'TCID', type: 'int', required: false, html: { page: 0, column: 0 } },
            { field: 'BID', type: 'int', required: false, html: { page: 0, column: 0 } },
            { field: 'BUD', type: 'list', options: {items: app.businesses}, required: false, html: { page: 0, column: 0 } },
            { field: 'NLID', type: 'int', required: false, html: { page: 0, column: 0 } },
            { field: 'CompanyName', type: 'text', required: false, html: { page: 0, column: 0 } },
            { field: 'IsCompany', type: 'list', options: {items: app.companyOrPerson}, required: true, html: { page: 0, column: 0 } },
            { field: 'SecondaryEmail', type: 'email', required: false, html: { page: 0, column: 0 } },
            { field: 'WorkPhone', type: 'phone', required: false, html: { page: 0, column: 0 } },
            { field: 'CellPhone', type: 'phone', required: false, html: { page: 0, column: 0 } },
            { field: 'Address', type: 'text', required: false, html: { page: 0, column: 0 } },
            { field: 'Address2', type: 'text', required: false, html: { page: 0, column: 0 } },
            { field: 'City', type: 'text', required: false, html: { page: 0, column: 0 } },
            { field: 'State', type: 'list', options: {items: app.usStateAbbr}, required: false, html: { page: 0, column: 0 } },
            { field: 'PostalCode', type: 'text', required: false, html: { page: 0, column: 0 } },
            { field: 'Country', type: 'text', required: false, html: { page: 0, column: 0 } },
            { field: 'Website', type: 'text', required: false, html: { page: 0, column: 0 } },
            { field: 'LastModTime', type: 'time', required: false, html: { page: 0, column: 0 } },
            { field: 'LastModBy', type: 'int', required: false, html: { page: 0, column: 0 } },
            { field: 'CreateTS', type: 'time', required: false, html: { page: 0, column: 0 } },
            { field: 'CreateBy', type: 'int', required: false, html: { page: 0, column: 0 } },
            { field: 'Points', type: 'text', required: false, html: { page: 1, column: 0 } },
            { field: 'DateofBirth', type: 'date', required: false, html: { page: 1, column: 0 } },
            { field: 'EmergencyContactName', type: 'text', required: false, html: { page: 1, column: 0 } },
            { field: 'EmergencyContactAddress', type: 'text', required: false, html: { page: 1, column: 0 } },
            { field: 'EmergencyContactTelephone', type: 'text', required: false, html: { page: 1, column: 0 } },
            { field: 'EmergencyEmail', type: 'text', required: false, html: { page: 1, column: 0 } },
            { field: 'AlternateAddress', type: 'text', required: false, html: { page: 1, column: 0 } },
            { field: 'EligibleFutureUser', type: 'list', options: {items: app.yesNoList}, required: false, html: { page: 1, column: 0 } },
            { field: 'Industry', type: 'text', required: false, html: { page: 1, column: 0 } },
            { field: 'SourceSLSID', type: 'w2int', required: false, html: { page: 1, column: 0 } },
            { field: 'CreditLimit', type: 'money', required: false, html: {page: 2, column: 0 } },
            { field: 'TaxpayorID', type: 'text', required: false, html: {page: 2, column: 0 } },
            { field: 'AccountRep', type: 'text', required: false, html: {page: 2, column: 0 } },
            { field: 'EligibleFuturePayor', type: 'list', options: {items: app.yesNoList}, required: false, html: {page: 2, column: 0 } },
            { field: 'EmployerName', type: 'text', required: false, html: {page: 3, column: 0 } },
            { field: 'EmployerStreetAddress', type: 'text', required: false, html: {page: 3, column: 0 } },
            { field: 'EmployerCity', type: 'text', required: false, html: {page: 3, column: 0 } },
            { field: 'EmployerState', type: 'list', options: {items: app.usStateAbbr}, required: false, html: {page: 3, column: 0 } },
            { field: 'EmployerPostalCode', type: 'text', required: false, html: {page: 3, column: 0 } },
            { field: 'EmployerEmail', type: 'text', required: false, html: {page: 3, column: 0 } },
            { field: 'EmployerPhone', type: 'text', required: false, html: {page: 3, column: 0 } },
            { field: 'Occupation', type: 'text', required: false, html: {page: 3, column: 0 } },
            { field: 'ApplicationFee', type: 'text', required: false, html: {page: 3, column: 0 } },
            { field: 'DesiredUsageStartDate', type: 'date', required: false, html: {page: 3, column: 0 } },
            { field: 'RentableTypePreference', type: 'text', required: false, html: {page: 3, column: 0 } },
            { field: 'FLAGS', type: 'text', required: false, html: {page: 3, column: 0 } },
            { field: 'Approver', type: 'text', required: false, html: {page: 3, column: 0 } },
            { field: 'DeclineReasonSLSID', type: 'w2int', required: false, html: {page: 3, column: 0 } },
            { field: 'OtherPreferences', type: 'text', required: false, html: {page: 3, column: 0 } },
            { field: 'FollowUpDate', type: 'date', required: false, html: {page: 3, column: 0 } },
            { field: 'CSAgent', type: 'text', required: false, html: {page: 3, column: 0 } },
            { field: 'OutcomeSLSID', type: 'text', required: false, html: {page: 3, column: 0 } },
            { field: 'FloatingDeposit', type: 'w2float', required: false, html: {page: 3, column: 0 } },
            { field: 'RAID', type: 'w2int', required: false, html: {page: 3, column: 0 } },
        ],
        tabs: [
            { id: 'tab1', caption: app.sTransactant },
            { id: 'tab2', caption: app.sUser },
            { id: 'tab3', caption: app.sPayor },
            { id: 'tab4', caption: app.sProspect },
        ],
        toolbar: {
            items: [
                { id: 'btnNotes', type: 'button', icon: 'far fa-sticky-note' },
                { id: 'bt3', type: 'spacer' },
                { id: 'btnClose', type: 'button', icon: 'fas fa-times' },
            ],
            onClick: function (event) {
                if (event.target == 'btnClose') {
                    var no_callBack = function() { return false; },
                        yes_callBack = function() {
                            w2ui.toplayout.hide('right',true);
                            w2ui.transactantsGrid.render();
                        };
                    form_dirty_alert(yes_callBack, no_callBack);
                }
                if (event.target == 'btnNotes') {
                    notesPopUp();
                }
            },
        },
        onValidate: function (event) {
            if (this.record.IsCompany.text == 'Person' && this.record.FirstName === '') {
                event.errors.push({
                    field: this.get('FirstName'),
                    error: 'FirstName required when "Person or Company" field is set to Person'
                });
            }
            if (this.record.IsCompany.text == 'Person' && this.record.LastName === '') {
                event.errors.push({
                    field: this.get('LastName'),
                    error: 'LastName required when "Person or Company" field is set to Person'
                });
            }
            if (this.record.IsCompany.text == 'Company' && this.record.CompanyName === '') {
                event.errors.push({
                    field: this.get('CompanyName'),
                    error: 'Company Name required when "Person or Company" field is set to Company'
                });
            }
        },
        actions: {
            save: function () {
                var tgrid = w2ui.transactantsGrid;
                tgrid.selectNone();
                console.log('before: tgrid.getSelection() = ' + tgrid.getSelection() );
                this.save({}, function (data) {
                    if (data.status == 'error') {
                        console.log('ERROR: '+ data.message);
                        return;
                    }
                    w2ui.toplayout.hide('right',true);
                    tgrid.render();
                });
            },
            saveadd: function() {
                var f = this,
                    grid = w2ui.transactantsGrid,
                    x = getCurrentBusiness(),
                    r = f.record,
                    BID=parseInt(x.value),
                    BUD=getBUDfromBID(BID);

                // clean dirty flag of form
                app.form_is_dirty = false;
                // clear the grid select recid
                app.last.grid_sel_recid  =-1;

                // select none if you're going to add new record
                grid.selectNone();

                f.save({}, function (data) {
                    if (data.status == 'error') {
                        console.log('ERROR: '+ data.message);
                        return;
                    }

                    // JUST RENDER THE GRID ONLY
                    grid.render();

                    // add new empty record and just refresh the form, don't need to do CLEAR form
                    var record = getTransactantInitRecord(BID, BUD);

                    f.record = record;
                    f.header = "Edit Transactant (new)"; // have to provide header here, otherwise have to call refresh method twice to get this change in form
                    f.url = '/v1/person/' + BID+'/0';
                    f.refresh();
                });
            },
            delete: function(/*target, data*/) {
                var form = this;
                w2confirm(delete_confirm_options)
                .yes(function() {
                    var tgrid = w2ui.transactantsGrid;
                    var params = {cmd: 'delete', formname: form.name, TCID: form.record.TCID };
                    var dat = JSON.stringify(params);

                    // delete Transactant request
                    $.post(form.url, dat, null, "json")
                    .done(function(data) {
                        if (data.status === "error") {
                            form.error(w2utils.lang(data.message));
                            return;
                        }
                        w2ui.toplayout.hide('right',true);
                        tgrid.remove(app.last.grid_sel_recid);
                        tgrid.render();
                    })
                    .fail(function(/*data*/){
                        form.error("Delete Transactant failed.");
                        return;
                    });
                })
                .no(function() {
                    return;
                });
            },
        },
        onRefresh: function(event) {
            event.onComplete = function() {
                var f = this,
                    r = f.record,
                    header="";

                // custom header
                if (r.TCID) {
                    if (f.original.IsCompany > 0) {
                        header = "Edit Transactant - {0} ({1})".format(r.CompanyName, r.TCID);
                    } else {
                        header = "Edit Transactant - {0} {1} ({2})".format(r.FirstName, r.LastName, r.TCID);
                    }
                } else {
                    header = "Edit Transactant ({0})".format("new");
                }

                formRefreshCallBack(f, "TCID", header);
            };
        },
        onChange: function(event) {
            event.onComplete = function() {
                // formRecDiffer: 1=current record, 2=original record, 3=diff object
                var diff = formRecDiffer(this.record, app.active_form_original, {});
                // if diff == {} then make dirty flag as false, else true
                if ($.isPlainObject(diff) && $.isEmptyObject(diff)) {
                    app.form_is_dirty = false;
                } else {
                    app.form_is_dirty = true;
                }
            };
        },
        onSubmit: function(target, data){
            delete data.postData.record.LastModTime;
            delete data.postData.record.LastModBy;
            delete data.postData.record.CreateTS;
            delete data.postData.record.CreateBy;
            // server request form data
            getFormSubmitData(data.postData.record);
        },
    });

};
