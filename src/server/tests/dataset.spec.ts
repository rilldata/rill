import { createServerActions } from "../actions";
import { getByID, summarizeCategoricalField, summarizeNumericField } from "../dataset"

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

const createAPI = () => ({
    getTopKAndCardinality: jest.fn(async (table, field) : Promise<CategoricalSummary> => ({
        topK,
        cardinality: 5
    })),
    numericHistogram: jest.fn(async (table, field, fieldType) : Promise<NumericHistogramBin[]> => (numericHistogram))
})

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


describe("summarizeCategorical", () => {

    let state:DataModelerState;
    let api:any;
    let getState = () => state;

    let notifier = jest.fn();
    let actions; //= createServerActions(createAPI(), notifier)(undefined, undefined);

    function dispatch(fcn:Function) {
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


    beforeEach(() => {
        // just mutate the state. Don't sweat it!
        state = mockState();
        api = createAPI();
        actions = createServerActions(api, notifier)(undefined, undefined);
    })

    it("produces a top-k table and cardinality in summary field of source profile", async () => {
    
        await 
            actions.summarizeCategoricalField('12345', 'test', 'test-field-01')
            (dispatch, () => state);
        
        
        expect(api.getTopKAndCardinality).toHaveBeenCalledTimes(1);
        
        const src = getByID(state.sources, '12345') as Source;
        const profile = src.profile[0];

        expect(profile.summary.cardinality).toBe(5);
        expect(profile.summary.topK).toEqual(topK);
    })

    it("only computes the top-k table and cardinality once", async () => {
        await 
            actions.summarizeCategoricalField('12345', 'test', 'test-field-01')
                (dispatch, () => state);
        await 
            actions.summarizeCategoricalField('12345', 'test', 'test-field-01')
                (dispatch, () => state);
    
        expect(api.getTopKAndCardinality).toHaveBeenCalledTimes(1);
    })
})

describe("summarizeNumericField", () => {

    let dispatch:Function;
    let state:DataModelerState;
    let api:any;
    let getState = () => state;

    beforeEach(() => {
        // just mutate the state. Don't sweat it!
        state = mockState();
        api = createAPI();
        dispatch = (fcn:Function) => {
            fcn(state);
        }
    })

    it("produces a histogram in summary field of source profile", async () => {
        //                         tbl?,  field?
        await summarizeNumericField('12345', 'test', 'test-field-02', 'DOUBLE', api, getState, dispatch);
        expect(api.numericHistogram).toHaveBeenCalledTimes(1);
        const src = getByID(state.sources, '12345') as Source;
        const profile = src.profile[1];
        expect(profile.summary.histogram).toEqual(numericHistogram);
    })

    it("only computes the top-k table and cardinality once", async () => {
        await summarizeNumericField('12345', 'test', 'test-field-02', 'INTEGER', api, getState, dispatch);
        await summarizeNumericField('12345', 'test', 'test-field-02', 'INTEGER', api, getState, dispatch);
        expect(api.numericHistogram).toHaveBeenCalledTimes(1);
    })
})