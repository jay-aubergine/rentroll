"use strict";

var GRID = "arsGrid";
var SIDEBAR_ID = "ars";
var FORM = "arsForm";

exports.gridConf = {
    grid: "arsGrid",
    sidebarID: "ars",
    capture: "arsGridRequest.png"
};

exports.formConf = {
    grid: "arsGrid",
    form: "arsForm",
    sidebarID: "ars",
    row: "0",
    capture: "arsFormRequest.png",
    captureAfterClosingForm: "arsFormRequestAfterClosingForm.png",
    buttonName: ["save", "saveadd", "delete"],
    testCount: 5
};

exports.addNewConf = {
    grid: GRID,
    form: FORM,
    sidebarID: SIDEBAR_ID,
    capture: "arsAddNewButton.png",
    inputField: ["Name", "Description"],
    buttonName: ["save", "saveadd"],
    inputSelectField: [{"fieldID": "ARType", "value":"Assessment"}, {"fieldID": "DebitLID", "value":" -- Select GL Account -- "}, {"fieldID": "CreditLID", "value":" -- Select GL Account -- "}],
    testCount: 16
};

