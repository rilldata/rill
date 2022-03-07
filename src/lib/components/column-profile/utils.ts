import { format } from "d3-format";
// stuff for figuring out what to do with which view?

export const config = {
    // width of summary value
    // width of null %
    nullPercentageWidth: 74,
    summaryVizWidth: {medium: 98, small:  60},
    exampleWidth: { medium: 204, small: 132}
}

export const percentage = format('.1%');