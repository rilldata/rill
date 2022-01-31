import {jest} from '@jest/globals'
import type { DataModelerState, MetricsModel, CategoricalSummary, NumericHistogramBin, TopKEntry } from "../../lib/types";

export const createAPI = () => ({
    getTopKAndCardinality: jest.fn(async (table, field) : Promise<CategoricalSummary> => ({
        topK,
        cardinality: 5
    })),
    getNullCount: jest.fn(async (table, field) => 10),
    numericHistogram: jest.fn(async (table, field, fieldType) : Promise<NumericHistogramBin[]> => (numericHistogram)),
    descriptiveStatistics: jest.fn(async (table, field) => ({max:10, min:5, mean: 7.5, q25: 6, q75: 8, median: 7}))
})

export const createDispatcher = (state:DataModelerState) => {
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


export const topK:TopKEntry[] = [
    {value: 'a', count: 100},
    {value: 'b', count: 50},
    {value: 'c', count: 25},
    {value: 'd', count: 10},
    {value: 'e', count: 5},
]

export const numericHistogram:NumericHistogramBin[] = [
    { bucket: 0, low: 0,   high: 1.5, count: 15 },
    { bucket: 1, low: 1.5, high: 3,   count: 10 },
    { bucket: 2, low: 3.5, high: 4.5, count :5 }
]


export const mockState = () : DataModelerState => ({
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
    metricsModels: [],
    exploreConfigurations: [],
    status: 'figure out later.'
})