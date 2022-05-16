import { swimLanePlacement } from './swim-lane-placement';

const testCases = [
    {
        shouldDoThis: "2 numbers map to 2 lanes",
        input: [1, 1],
        columns: 2,
        callback: v => v,
        output: [
            [1],
            [1],
        ]
    },
    {
        shouldDoThis: "2 numbers map to 2 lanes, with one empty",
        input: [1, 1],
        columns: 3,
        callback: v => v,
        output: [
            [1],
            [1],
            []
        ]
    },
    {
        shouldDoThis: "2 numbers map to 1 lane",
        input: [1, 1],
        columns: 1,
        callback: v => v,
        output: [
            [1, 1],
        ]
    },
    {
        shouldDoThis: "8 numbers map to 3 even lanes",
        input: [1, 1, 1, 1, 1, 1, 1, 1],
        columns: 3,
        callback: v => v,
        output: [
            [1, 1, 1],
            [1, 1, 1],
            [1, 1]
        ]
    },
    {
        shouldDoThis: "13 numbers map to 3 even lanes",
        input: Array.from({length: 13}).fill(1),
        columns: 3,
        callback: v => v,
        output: [
            Array.from({length: 5}).fill(1),
            Array.from({length: 4}).fill(1),
            Array.from({length: 4}).fill(1),
        ]
    },
    {
        shouldDoThis: "14 numbers map to 3 even lanes",
        input: Array.from({length: 14}).fill(1),
        columns: 3,
        callback: v => v,
        output: [
            Array.from({length: 5}).fill(1),
            Array.from({length: 5}).fill(1),
            Array.from({length: 4}).fill(1),
        ]
    },
    {
        shouldDoThis: "14 objects map to 3 even lanes using a callback",
        input: Array.from({length: 14}).fill({height: 1}),
        columns: 3,
        callback: (v:({height: number})) => v.height,
        output: [
            Array.from({length: 5}).fill({height: 1}),
            Array.from({length: 5}).fill({height: 1}),
            Array.from({length: 4}).fill({height: 1}),
        ]
    },
]

describe('swimLanePlacement', () => {
    /** iterate through the testCases */
    for (const testCase of testCases) {
        it(testCase.shouldDoThis, () => {

            expect(
                swimLanePlacement(
                    testCase.input,
                    testCase.callback,
                    testCase.columns
                )
            ).toEqual(
                testCase.output
            )

        })
    }
})