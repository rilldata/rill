import { format } from "d3-format";
// stuff for figuring out what to do with which view?

export const config = {
  // width of summary value
  // width of null %
  nullPercentageWidth: 74,
  mediumCutoff: 300,
  compactBreakpoint: 350,
  hideRight: 325,
  hideNullPercentage: 400,
  summaryVizWidth: { medium: 98, small: 60 },
  exampleWidth: { medium: 204, small: 132 },
};

export const percentage = format(".1%");
