import type { 
    DataModelerState, 
    Source, 
    CategoricalSummary, TopKEntry,
    NumericHistogramBin,  
    
} from "src/types";
import { getByID, summarizeCategoricalField, summarizeNumericField } from "./"

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

    it("produces a top-k table and cardinality in summary field of source profile", async () => {
        //                         tbl?,  field?
        await summarizeCategoricalField('12345', 'test', 'test-field-01', api, getState, dispatch);
        expect(api.getTopKAndCardinality).toHaveBeenCalledTimes(1);
        const src = getByID(state.sources, '12345') as Source;
        const profile = src.profile[0];
        expect(profile.summary.cardinality).toBe(5);
        expect(profile.summary.topK).toEqual(topK);
    })

    it("only computes the top-k table and cardinality once", async () => {
        await summarizeCategoricalField('12345', 'test', 'test-field-01', api, getState, dispatch);
        await summarizeCategoricalField('12345', 'test', 'test-field-01', api, getState, dispatch);
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