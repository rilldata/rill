/**
 * The goal of this test suite is to have coverage for the 
 * state transformations that result from dataset API calls.
 * We mock the API calls here & check that the correct part of the data modeler
 * state is updated.
 */
import { getByID, createDatasetActions } from "../dataset"
import { mockState, createAPI, createDispatcher, topK, numericHistogram } from './mocks'

import type { DataModelerState, Source } from "src/types";

describe("dataset actions", () => {

    let state:DataModelerState;
    let api:any;
    let actions;
    let dispatch:Function;


    beforeEach(() => {
        state = mockState();
        api = createAPI();
        dispatch = createDispatcher(state);
        actions = createDatasetActions(api);
    })

    it("getTopKAndCardinality: runs the getTopKAndCardinality API call and updatse the topK and cardinality summary fields", async () => {
    
        await 
            actions.summarizeCategoricalField('12345', 'test', 'test-field-01')
            (dispatch, () => state);
        
        
        expect(api.getTopKAndCardinality).toHaveBeenCalledTimes(1);
        
        const src = getByID(state.sources, '12345') as Source;
        const profile = src.profile[0];

        expect(profile.summary.cardinality).toBe(5);
        expect(profile.summary.topK).toEqual(topK);
    })

    it("getTopKAndCardinality: does not re-run the getTopKAndCardinality API call more than once for a given id, table, and field", async () => {
        await 
            actions.summarizeCategoricalField('12345', 'test', 'test-field-01')
                (dispatch, () => state);
        await 
            actions.summarizeCategoricalField('12345', 'test', 'test-field-01')
                (dispatch, () => state);
    
        expect(api.getTopKAndCardinality).toHaveBeenCalledTimes(1);
    })

    it("summarizeNumericField: runs the numericHistogram API function for the chosen id, table, and field and updates the state", async () => {
        await 
            actions.summarizeNumericField('12345', 'test', 'test-field-02', 'DOUBLE')
            (dispatch, () => state);
        expect(api.numericHistogram).toHaveBeenCalledTimes(1);
        const src = getByID(state.sources, '12345') as Source;
        const profile = src.profile[1];
        expect(profile.summary.histogram).toEqual(numericHistogram);
    })

    it("summarizeNumericField: does not re-run the numericHistogram API on multiple calls", async () => {
        await 
            actions.summarizeNumericField('12345', 'test', 'test-field-02', 'DOUBLE')
            (dispatch, () => state);
        await 
            actions.summarizeNumericField('12345', 'test', 'test-field-02', 'DOUBLE')
            (dispatch, () => state);

        expect(api.numericHistogram).toHaveBeenCalledTimes(1);
        const src = getByID(state.sources, '12345') as Source;
        const profile = src.profile[1];
        expect(profile.summary.histogram).toEqual(numericHistogram);
    });

    it("summarizeNullCount: runs the getNullCount API call and updates the nullCount field", async() => {
        await 
            actions.summarizeNullCount('12345', 'test', 'test-field-02')
            (dispatch, () => state);
        expect(api.getNullCount).toHaveBeenCalledTimes(1);
        const src = getByID(state.sources, '12345') as Source;
        const profile = src.profile[1];
        expect(profile.nullCount).toEqual(10);
    })

    it("summarizeNullCount: does not re-run the getNullCount API on multiple calls", async () => {
        await 
            actions.summarizeNullCount('12345', 'test', 'test-field-02')
            (dispatch, () => state);
        await 
            actions.summarizeNullCount('12345', 'test', 'test-field-02')
            (dispatch, () => state);

        expect(api.getNullCount).toHaveBeenCalledTimes(1);
        const src = getByID(state.sources, '12345') as Source;
        const profile = src.profile[1];
        expect(profile.nullCount).toEqual(10);
    });
})
