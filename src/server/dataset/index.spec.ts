/**
 * The goal of this test suite is to have coverage for the 
 * state transformations that result from dataset API calls.
 * We mock the API calls here & check that the correct part of the data modeler
 * state is updated.
 */
import { getByID, createDatasetActions } from "./"

import type { 
    DataModelerState, 
    Source, 
    CategoricalSummary, TopKEntry,
    NumericHistogramBin,  
    
} from "src/types";

const topK:TopKEntry[] = [
    {value: 'a', count: 100},
    {value: 'b', count: 50},
    {value: 'c', count: 25},
    {value: 'd', count: 10},
    {value: 'e', count: 5},
]

const numericHistogram:NumericHistogramBin[] = [
    { bucket: 0, low: 0,   high: 1.5, count: 15 },
    { bucket: 1, low: 1.5, high: 3,   count: 10 },
    { bucket: 2, low: 3.5, high: 4.5, count :5 }
]

/**
 * Mock for API calls. We will test these API calls elsewhere.
 */
const createAPI = () => ({
    getTopKAndCardinality: jest.fn(async (table, field) : Promise<CategoricalSummary> => ({
        topK,
        cardinality: 5
    })),
    getNullCount: jest.fn(async (table, field) => 10),
    numericHistogram: jest.fn(async (table, field, fieldType) : Promise<NumericHistogramBin[]> => (numericHistogram))
})

const createDispatcher = (state:DataModelerState) => {
    return function dispatch(fcn:Function) {
        // this works very similar to what you'd expect in a redux setting.
        // eg. dispatch(changeChannel('beta')) should take the changeChannel
        // action, which returns a draft-mutating function to be fed into
        // immer's produce function.
        if (fcn.constructor.name === 'AsyncFunction') {
            // I thought about using func.length (if it has two args, then we are go)
            // but you may only have one. For now, I think marking a function a async
            // works.
            fcn(dispatch, () => state);
        } else {
            // atomic update (singular state change).
            fcn(state);
            //store.update(draft => fcn(state));
        }
    }
}

const mockState = () : DataModelerState => ({
    sources: [
        {
            id: '12345',
            name: 'test',
            path: './scripts/test.parquet',
            head: [],
            profile: [ 
                { name: 'test-field-01', type: 'BYTE_ARRAY', conceptualType: 'string' },
                { name: 'test-field-02', type: 'DOUBLE', conceptualType: 'double' }
             ]
        }
    ],
    queries: [],
    status: 'figure out later.'
})


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
