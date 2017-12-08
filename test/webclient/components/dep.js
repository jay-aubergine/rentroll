"use strict";

var GRID = "depGrid";
var FORM = "depForm";
var SIDEBAR_ID = "dep";

exports.gridConf = {
    grid: "depGrid",
    sidebarID: "dep",
    capture: "depGridRequest.png"
};

exports.formConf = {
    grid: "depGrid",
    form: "depForm",
    sidebarID: "dep",
    row: "0",
    capture: "depFormRequest.png",
    captureAfterClosingForm: "depFormRequestAfterClosingForm.png",
    buttonName: ["save", "saveadd", "delete"],
    testCount: 5
};

exports.addNewConf = {
    grid: GRID,
    form: FORM,
    sidebarID: SIDEBAR_ID,
    capture: "depAddNewButton.png",
    inputField: ["Name", "AccountNo"],
    buttonName: ["save", "saveadd"],
    inputSelectField: [{"fieldID": "LID", "value":" -- Select GL Account -- "}],
    testCount: 12
};

