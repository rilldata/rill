const markColor = "royalblue";
const axisColor = "#E5E7EB";
const axisLabelColor = "#727883";

export const getRillTheme = () => ({
  mark: {
    color: markColor,
  },
  arc: { fill: markColor },
  area: { fill: markColor },
  line: { stroke: markColor },
  path: { stroke: markColor },
  rect: { fill: markColor },
  shape: { stroke: markColor },
  symbol: { fill: markColor },

  axisY: {
    gridColor: axisColor,
    tickColor: axisColor,
    domain: false,
    labelFont: "Inter, sans-serif",
    labelFontSize: 10,
    labelFontWeight: "normal",
    labelColor: axisLabelColor,
    labelPadding: 5,
    titleColor: axisLabelColor,
    titleFont: "Inter, sans-serif",
    titleFontSize: 12,
    titleFontWeight: "bold",
    titlePadding: 10,
    labelOverlap: false,
  },
  axisX: {
    gridColor: axisColor,
    tickColor: axisColor,
    tickSize: 0,
    domain: false,
    labelFont: "Inter, sans-serif",
    labelFontSize: 10,
    labelFontWeight: "normal",
    labelPadding: 5,
    labelColor: axisLabelColor,
    titleColor: axisLabelColor,
    titleFont: "Inter, sans-serif",
    titleFontSize: 12,
    titleFontWeight: "bold",
    titlePadding: 10,
  },
  view: {
    strokeWidth: 0,
  },
  range: {
    category: [
      "#3e5c69",
      "#6793a6",
      "#182429",
      "#0570b0",
      "#3690c0",
      "#74a9cf",
      "#a6bddb",
      "#e2ddf2",
    ],
  },
});
