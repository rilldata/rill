import { format } from "d3-format";
// stuff for figuring out what to do with which view?

export const config = {
    // width of summary value
    // width of null %
    nullPercentageWidth: "68px",
    summaryVizWidth: "108px",
    timestamp: {
        histogramColor: 'black',
        nullPercentageColor: 'black',
    },
    numeric: {
        histogramColor: 'red',
        nullPercentageColor: 'red',
    },
    categorical: {
        histogramColor: 'blue',
        nullPercentageColor: 'blue',
        cardinalityColor: 'blue',
        topKColor: 'blue'
    },
}

export const percentage = format('.1%');