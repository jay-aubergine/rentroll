"use strict";

import * as constants from '../support/utils/constants';
import * as selectors from '../support/utils/get_selectors';
import * as common from '../support/utils/common';

// --- Collections ---
const section = require('../support/components/raStatements'); // RA Statements
const subSection = require('../support/components/raStatementsDetails'); // RA Statements

// this contain app variable of the application
let appSettings;

// holds the test configuration for the modules
let testConfig;
let testConfig2;

// -- Start Cypress UI tests for AIR Roller Application --
describe('AIR Roller UI Tests - RA Statements', function () {

    // // records list of module from the API response
    let recordsAPIResponse;

    let noRecordsInAPIResponse;

    // -- Perform operation before all tests starts. It runs once before all tests in the block --
    /********************************
     * Login into application
     * Select node from left sidebar
     * Route the response for grid records
     *
     * Expect:
     * Grid records response must have status flag as success.
     ********************************/
    before(function () {

        testConfig = section.conf;
        testConfig2 = subSection.conf;

        // --- Login into Application before starting any tests ---
        // Check custom login command for more detail. File path: ./../support/commands.js
        cy.login();

        cy.visit(constants.ROLLER_APPLICATION_PATH).wait(constants.PAGE_LOAD_TIME);

        // Starting a server to begin routing responses to cy.route()
        cy.server();

        // To manage the behavior of network requests. Routing the response for the requests.
        cy.route(testConfig.methodType, common.getAPIEndPoint(testConfig.sidebarID)).as('getRecords');

        /************************
         * Select right side node
         *************************/
        // Node should be visible and selected
        cy.get(selectors.getNodeSelector(testConfig.sidebarID))
            .scrollIntoView()
            .should('be.visible')
            .click().wait(constants.WAIT_TIME)
            .should('have.class', 'w2ui-selected');

        // If have date navigation bar than change from and to Date to get in between data
        if (testConfig.haveDateValue) {
            common.changeDate(testConfig.sidebarID, testConfig.fromDate, testConfig.toDate);
        }

        cy.wait(constants.WAIT_TIME);

        // Check http status
        cy.wait('@getRecords').its('status').should('eq', constants.HTTP_OK_STATUS);

        // get API endpoint's responseBody
        cy.get('@getRecords').then(function (xhr) {

            // Check key `status` in responseBody
            expect(xhr.responseBody).to.have.property('status', constants.API_RESPONSE_SUCCESS_FLAG);

            // get records list from the API response
            recordsAPIResponse = xhr.response.body.records;

            // -- Assigning number of records to 0 if no records are available in response --
            if (recordsAPIResponse) {
                noRecordsInAPIResponse = xhr.response.body.records.length;
            } else {
                noRecordsInAPIResponse = 0;
            }
        });

    });

    // -- Perform operation before each test(it()) starts. It runs before each test in the block. --
    beforeEach(function () {

        // -- get app variable from the window --
        /*
        * After successfully login into application it will have fixed app variable.
        * Fetching it after successful login.
        * */
        cy.window().then((win) => {
            appSettings = win.app;
        });

        cy.log(appSettings);

    });

    // -- Change business to REX --
    it('Change business to REX', function () {
        // onSuccessful test set BID value. If above test get fail below code will not be executed.
        constants.BID = common.changeBU(appSettings);
    });

    /***********************
     * Iterate through each cell.
     *
     * Expect:
     * Cell value must be same as record's field value from API Response.
     ***********************/
    it('Grid Records', function () {
        common.testGridRecords(recordsAPIResponse, noRecordsInAPIResponse, testConfig);
    });

    /*******************************
     * Click on first record of grid
     *
     * Expect:
     * White section must have proper detail same as api response
     * Grid's value must be same as record's field value from API Response
     *
     * Check visibility of CSV/PDF Button
     *
     * Close the form
     ********************************/
    it('Record Detail Form', function () {
        // ----------------------------------
        // -- Tests for detail record form --
        // ----------------------------------
        // Params:
        // recordsAPIResponse: list of record from the api response,
        // testConfig: configuration for running tests
        common.testDetailFormWithGrid(recordsAPIResponse, testConfig, testConfig2);

        // Check visibility of export to CSV button
        cy.get(selectors.getExportCSVButtonSelector(testConfig.form)).should('be.visible');

        // Check visibility of export to PDF button
        cy.get('#tb_stmtDetailForm_toolbar_item_pdfexport').should('be.visible');

        // -- Close the form. And assert that form isn't visible. --
        common.closeFormTests(selectors.getFormSelector(testConfig.form));
    });

    // -- Perform operation after all tests finish. It runs once after all tests in the block --
    after(function () {

        // --- Logout from the Application after finishing all tests ---
        // Check custom login command for more detail. File path: ./../support/commands.js
        cy.logout();
    });
});