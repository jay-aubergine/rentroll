"use strict";

import * as constants from '../support/utils/constants';
import * as selectors from '../support/utils/get_selectors';
import * as common from '../support/utils/common';

// --- Collections ---
const section = require('../support/components/rentalAgreements'); // Rental Agreements

// this contain app variable of the application
let appSettings;

// holds the test configuration for the modules
let testConfig;

// -- Start Cypress UI tests for AIR Roller Application --
describe('AIR Roller UI Tests - Rental Agreements', function () {

    // // records list of module from the API response
    let recordsAPIResponse;

    let noRecordsInAPIResponse;

    // -- Perform operation before all tests starts. It runs once before all tests in the block --
    before(function () {

        testConfig = section.conf;

        // --- Login into Application before starting any tests ---
        // Check custom login command for more detail. File path: ./../support/commands.js
        cy.login();

        cy.visit(constants.ROLLER_APPLICATION_PATH).wait(constants.PAGE_LOAD_TIME);

        // Starting a server to begin routing responses to cy.route()
        cy.server();

        // To manage the behavior of network requests. Routing the response for the requests.
        cy.route(testConfig.methodType, common.getAPIEndPoint("flow")).as('getRecords');

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

        // waiting for response of second call on api at date change
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

    // it('Existing Rental Agreement', function (){
    //     cy.server();
    //
    //     // Click on the first record
    //     cy.route(testConfig.methodType, common.getDetailRecordAPIEndPoint("flow", 0)).as('raRecord');
    //
    //     cy.get(selectors.getFirstRecordInGridSelector(testConfig.grid)).click();
    //
    //     // Check http status
    //     cy.wait('@raRecord').its('status').should('eq', constants.HTTP_OK_STATUS);
    //
    //     cy.get('@raRecord').then(function (xhr){
    //         // Check key `status` in responseBody
    //         expect(xhr.responseBody).to.have.property('status', constants.API_RESPONSE_SUCCESS_FLAG);
    //
    //         cy.log(xhr);
    //     });
    //
    //     cy.wait(5000);
    //
    //     // Edit RAFlow
    //     cy.route(testConfig.methodType, common.getDetailRecordAPIEndPoint("flow", 0)).as('editRARecord');
    //
    //     cy.get(selectors.getEditRAFlowButtonSelector()).click();
    //
    //     // Check http status
    //     cy.wait('@editRARecord').its('status').should('eq', constants.HTTP_OK_STATUS);
    //
    //     cy.get('@editRARecord').then(function (xhr){
    //         // Check key `status` in responseBody
    //         expect(xhr.responseBody).to.have.property('status', constants.API_RESPONSE_SUCCESS_FLAG);
    //
    //         cy.log(xhr);
    //
    //         // TODO: [WIP]Write test for verifying grids/forms for each section
    //         // let flowData = xhr.response.body.record.Flow.Data;
    //         //
    //         // cy.log("people response");
    //         //
    //         // cy.log(flowData.people);
    //         //
    //         // cy.wait(5000);
    //         //
    //         // // people section
    //         // cy.get('#people').click();
    //         //
    //         // cy.wait(10000);
    //         //
    //         // testConfig.grid = "RAPeopleGrid";
    //         // testConfig.excludeGridColumns = ["haveError"];
    //         // common.testGridRecords(flowData.people, flowData.people.length, testConfig);
    //
    //
    //     });
    //
    // });

    // -- Perform operation after all tests finish. It runs once after all tests in the block --
    after(function () {

        // --- Logout from the Application after finishing all tests ---
        // Check custom login command for more detail. File path: ./../support/commands.js
        cy.logout();
    });
});
