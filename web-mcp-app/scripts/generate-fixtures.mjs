// Generates the fixture files under harness/fixtures/. Each fixture is one tool call:
// the `create_chart` arguments and the result the host forwards to the widget, shaped
// like `CreateChartResult` in runtime/ai/create_chart.go. `bar_chart.json` is the
// canonical example the backend output is checked against.
// Run with `node scripts/generate-fixtures.mjs && npx prettier --write harness/fixtures`;
// the output is committed.
import { mkdirSync, writeFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

const outDir = fileURLToPath(new URL("../harness/fixtures/", import.meta.url));
mkdirSync(outDir, { recursive: true });

const RANGE = { start: "2023-09-09T00:00:00Z", end: "2023-09-16T00:00:00Z" };
const COMPARISON_RANGE = {
  start: "2023-09-02T00:00:00Z",
  end: "2023-09-09T00:00:00Z",
};
const OPEN_URL =
  "https://ui.rilldata.com/demo/bids/-/ai/session-0b1c/message/call-4f2a/-/open";
const THEME = {
  primaryColorRaw: "#4f46e5",
  light: { primary: "#4f46e5", secondary: "#f97316" },
  dark: { primary: "#818cf8", secondary: "#fb923c" },
};
const METRICS_VIEW = {
  name: "bids_metrics",
  display_name: "Bids",
  time_dimension: "__time",
  smallest_time_grain: "TIME_GRAIN_HOUR",
};

const FIELDS = {
  __time: {
    name: "__time",
    kind: "time",
    display_name: "Time",
    type: "TIMESTAMP",
  },
  advertiser_name: {
    name: "advertiser_name",
    kind: "dimension",
    display_name: "Advertiser Name",
    description: "Name of the advertiser bidding on the impression.",
    type: "STRING",
  },
  publisher: {
    name: "publisher",
    kind: "dimension",
    display_name: "Publisher",
    type: "STRING",
  },
  device: {
    name: "device",
    kind: "dimension",
    display_name: "Device",
    type: "STRING",
  },
  campaign_stage: {
    name: "campaign_stage",
    kind: "dimension",
    display_name: "Campaign Stage",
    type: "STRING",
  },
  total_bids: {
    name: "total_bids",
    kind: "measure",
    display_name: "Total Bids",
    description: "Count of bid requests.",
    type: "INT64",
    format_preset: "humanize",
  },
  impressions: {
    name: "impressions",
    kind: "measure",
    display_name: "Impressions",
    type: "INT64",
    format_preset: "humanize",
  },
  total_spend: {
    name: "total_spend",
    kind: "measure",
    display_name: "Total Spend",
    type: "FLOAT64",
    format_preset: "currency_usd",
  },
  win_rate: {
    name: "win_rate",
    kind: "measure",
    display_name: "Win Rate",
    type: "FLOAT64",
    format_preset: "percentage",
    lower_is_better: false,
  },
  avg_bid_price: {
    name: "avg_bid_price",
    kind: "measure",
    display_name: "Avg Bid Price",
    type: "FLOAT64",
    format_d3: "$.2f",
  },
  spend_change: {
    name: "spend_change",
    kind: "measure",
    display_name: "Spend Change",
    description: "Week-over-week change in spend.",
    type: "FLOAT64",
    format_preset: "currency_usd",
  },
};

const DAYS = Array.from(
  { length: 7 },
  (_, i) => `2023-09-${String(9 + i).padStart(2, "0")}T00:00:00Z`,
);
const ADVERTISERS = [
  "Acme Media",
  "Beta Brands",
  "Cobalt Retail",
  "Delta Travel",
  "Echo Games",
  "Foxtrot Foods",
  "Gamma Motors",
  "Helix Health",
];
const PUBLISHERS = [
  "news.example",
  "sports.example",
  "video.example",
  "music.example",
  "shop.example",
];
const DEVICES = ["Mobile", "Desktop", "Tablet"];
const STAGES = ["Impression", "Click", "Add to cart", "Purchase"];

// Deterministic numbers so the fixtures are stable across regenerations.
let seed = 42;
function rand() {
  seed = (seed * 1664525 + 1013904223) % 4294967296;
  return seed / 4294967296;
}
const int = (min, max) => Math.floor(min + rand() * (max - min));
const round = (value, digits = 2) =>
  Math.round(value * 10 ** digits) / 10 ** digits;

const pickFields = (...names) =>
  Object.fromEntries(names.map((name) => [name, FIELDS[name]]));

const CHART_LABELS = {
  bar_chart: "Bar",
  line_chart: "Line",
  area_chart: "Stacked area",
  stacked_bar: "Stacked bar",
  stacked_bar_normalized: "Stacked bar (normalized)",
  donut_chart: "Donut",
  pie_chart: "Pie",
  funnel_chart: "Funnel",
  heatmap: "Heatmap",
  combo_chart: "Combo",
  scatter_plot: "Scatter",
};

function textFallback(chartType, structured) {
  const {
    title,
    data = [],
    fields = {},
    resolved_time_range: range,
  } = structured;
  const header = `**${CHART_LABELS[chartType] ?? chartType}: ${title}**${
    range ? ` (${range.start.slice(0, 10)} to ${range.end.slice(0, 10)})` : ""
  }`;
  if (!data.length) return `${header}\n\n_No rows._`;
  const columns = Object.keys(data[0]);
  const label = (column) =>
    column.endsWith("_prev")
      ? `${fields[column.slice(0, -5)]?.display_name ?? column.slice(0, -5)} (prev)`
      : (fields[column]?.display_name ?? column);
  const lines = [
    `| ${columns.map(label).join(" | ")} |`,
    `| ${columns.map(() => "-").join(" | ")} |`,
    ...data
      .slice(0, 20)
      .map(
        (row) =>
          `| ${columns.map((column) => String(row[column] ?? "")).join(" | ")} |`,
      ),
  ];
  const parts = [
    header,
    "",
    ...lines,
    "",
    `_${Math.min(20, data.length)} of ${data.length} rows shown_`,
  ];
  if (structured.open_url)
    parts.push(
      "",
      `Open the interactive chart in Rill: ${structured.open_url}`,
    );
  return parts.join("\n");
}

function fixture({ name, description, expect, chartType, spec, structured }) {
  const structuredContent = {
    widget: "chart",
    chart_type: chartType,
    spec,
    message: `Chart created successfully: ${chartType}`,
    ...structured,
  };
  return {
    name,
    description,
    expect,
    arguments: { chart_type: chartType, spec },
    result: {
      content: [
        { type: "text", text: textFallback(chartType, structuredContent) },
      ],
      structuredContent,
    },
  };
}

function common(extra = {}) {
  return {
    metrics_view: METRICS_VIEW,
    resolved_time_range: RANGE,
    theme: THEME,
    open_url: OPEN_URL,
    truncated: false,
    ...extra,
  };
}

function write(name, content) {
  writeFileSync(
    `${outDir}${name}.json`,
    JSON.stringify(content, null, 2) + "\n",
  );
  console.log(`wrote harness/fixtures/${name}.json`);
}

const timeRange = { start: RANGE.start, end: RANGE.end };

// Canonical bar chart: nominal x, one measure, static primary color.
const barBids = [128430, 99210, 87650, 64320, 51200, 40880, 28100, 19750];
const barSpec = {
  metrics_view: "bids_metrics",
  time_range: timeRange,
  x: { field: "advertiser_name", type: "nominal", limit: 10, sort: "-y" },
  y: { field: "total_bids", type: "quantitative" },
  color: "primary",
};
const barStructured = common({
  title: "Total Bids per Advertiser Name",
  data: ADVERTISERS.map((advertiser_name, i) => ({
    advertiser_name,
    total_bids: barBids[i],
  })),
  fields: pickFields("advertiser_name", "total_bids"),
  domain_values: { advertiser_name: ADVERTISERS },
});
write(
  "bar_chart",
  fixture({
    name: "Bar chart",
    description:
      "Canonical create_chart result: top advertisers by total bids.",
    expect: "canvas",
    chartType: "bar_chart",
    spec: barSpec,
    structured: barStructured,
  }),
);

write(
  "horizontal_bar",
  fixture({
    name: "Horizontal bar chart",
    description: "Measure on x and the dimension on y draws horizontal bars.",
    expect: "canvas",
    chartType: "bar_chart",
    spec: {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      x: { field: "total_bids", type: "quantitative" },
      y: { field: "advertiser_name", type: "nominal", limit: 10, sort: "-x" },
      color: "primary",
    },
    structured: barStructured,
  }),
);

write(
  "multi_measure_bar",
  fixture({
    name: "Multi-measure bar chart",
    description:
      "Two measures side by side; the color channel is the synthetic rill_measures field.",
    expect: "canvas",
    chartType: "bar_chart",
    spec: {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      x: { field: "advertiser_name", type: "nominal", limit: 5, sort: "-y" },
      y: {
        field: "total_bids",
        type: "quantitative",
        fields: ["total_bids", "impressions"],
      },
      color: { field: "rill_measures", type: "nominal" },
    },
    structured: common({
      title: "Total Bids, Impressions by Advertiser Name",
      data: ADVERTISERS.slice(0, 5).map((advertiser_name, i) => ({
        advertiser_name,
        total_bids: barBids[i],
        impressions: Math.round(barBids[i] * (0.55 + rand() * 0.2)),
      })),
      fields: pickFields("advertiser_name", "total_bids", "impressions"),
      domain_values: {
        advertiser_name: ADVERTISERS.slice(0, 5),
        rill_measures: ["total_bids", "impressions"],
      },
    }),
  }),
);

// Daily series with one missing bucket (Sep 12) that the widget zero-fills.
const dailyBids = [41200, 45800, 39900, 0, 47300, 52100, 36400];
const lineSpec = {
  metrics_view: "bids_metrics",
  time_range: timeRange,
  time_grain: "TIME_GRAIN_DAY",
  x: { field: "__time", type: "temporal" },
  y: { field: "total_bids", type: "quantitative", zeroBasedOrigin: true },
  color: "primary",
};
write(
  "line_chart",
  fixture({
    name: "Line chart",
    description:
      "Daily total bids; Sep 12 is missing from the rows and gets zero-filled.",
    expect: "canvas",
    chartType: "line_chart",
    spec: lineSpec,
    structured: common({
      title: "Total Bids over Time",
      time_grain: "TIME_GRAIN_DAY",
      data: DAYS.filter((_, i) => i !== 3).map((__time, i) => ({
        __time,
        total_bids: dailyBids[i < 3 ? i : i + 1],
      })),
      fields: pickFields("__time", "total_bids"),
    }),
  }),
);

write(
  "comparison",
  fixture({
    name: "Line chart with comparison",
    description:
      "Daily total bids compared with the previous week (`_prev` columns).",
    expect: "canvas",
    chartType: "line_chart",
    spec: {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      comparison_time_range: {
        start: COMPARISON_RANGE.start,
        end: COMPARISON_RANGE.end,
      },
      time_grain: "TIME_GRAIN_DAY",
      x: { field: "__time", type: "temporal" },
      y: { field: "total_bids", type: "quantitative" },
      color: "primary",
    },
    structured: common({
      title: "Total Bids over Time",
      time_grain: "TIME_GRAIN_DAY",
      data: DAYS.map((__time, i) => ({
        __time,
        total_bids: dailyBids[i] || 43800,
        total_bids_prev: Math.round(
          (dailyBids[i] || 43800) * (0.8 + rand() * 0.3),
        ),
      })),
      fields: pickFields("__time", "total_bids"),
      resolved_comparison_time_range: COMPARISON_RANGE,
    }),
  }),
);

const areaRows = [];
DAYS.forEach((__time, dayIndex) => {
  DEVICES.forEach((device, deviceIndex) => {
    // Leave two holes to exercise per-series zero-filling.
    if (
      (dayIndex === 3 && device === "Mobile") ||
      (dayIndex === 5 && device === "Tablet")
    )
      return;
    areaRows.push({
      __time,
      device,
      total_spend: round(1800 + deviceIndex * 600 + rand() * 900),
    });
  });
});
write(
  "area_chart",
  fixture({
    name: "Stacked area chart",
    description: "Daily spend split by device, with two missing buckets.",
    expect: "canvas",
    chartType: "area_chart",
    spec: {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      time_grain: "TIME_GRAIN_DAY",
      x: { field: "__time", type: "temporal" },
      y: { field: "total_spend", type: "quantitative" },
      color: { field: "device", type: "nominal", limit: 3 },
    },
    structured: common({
      title: "Total Spend over Time split by Device",
      time_grain: "TIME_GRAIN_DAY",
      data: areaRows,
      fields: pickFields("__time", "device", "total_spend"),
      domain_values: { device: DEVICES },
    }),
  }),
);

const stackedRows = [];
PUBLISHERS.forEach((publisher, p) => {
  DEVICES.forEach((device, d) => {
    stackedRows.push({
      publisher,
      device,
      impressions: int(20000, 90000) - p * 2500 - d * 1500,
    });
  });
});
for (const [name, chartType, label] of [
  ["stacked_bar", "stacked_bar", "Stacked bar chart"],
  [
    "stacked_bar_normalized",
    "stacked_bar_normalized",
    "Normalized stacked bar chart",
  ],
]) {
  write(
    name,
    fixture({
      name: label,
      description: "Impressions per publisher split by device.",
      expect: "canvas",
      chartType,
      spec: {
        metrics_view: "bids_metrics",
        time_range: timeRange,
        x: { field: "publisher", type: "nominal", limit: 10, sort: "-y" },
        y: { field: "impressions", type: "quantitative" },
        color: { field: "device", type: "nominal", limit: 10 },
      },
      structured: common({
        title: "Impressions per Publisher split by Device",
        data: stackedRows,
        fields: pickFields("publisher", "device", "impressions"),
        domain_values: { publisher: PUBLISHERS, device: DEVICES },
      }),
    }),
  );
}

const spendByPublisher = [48210.5, 36990.25, 25480.75, 18760.0, 9340.4];
write(
  "donut_chart",
  fixture({
    name: "Donut chart",
    description: "Spend share per publisher with a total in the center.",
    expect: "canvas",
    chartType: "donut_chart",
    spec: {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      measure: { field: "total_spend", type: "quantitative", showTotal: true },
      color: { field: "publisher", type: "nominal", limit: 5 },
      innerRadius: 50,
    },
    structured: common({
      title: "Total Spend split by Publisher",
      data: PUBLISHERS.map((publisher, i) => ({
        publisher,
        total_spend: spendByPublisher[i],
      })),
      fields: pickFields("publisher", "total_spend"),
      domain_values: { publisher: PUBLISHERS },
      totals: { total_spend: 151220.9 },
    }),
  }),
);

write(
  "pie_chart",
  fixture({
    name: "Pie chart",
    description: "Bid share per device.",
    expect: "canvas",
    chartType: "pie_chart",
    spec: {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      measure: { field: "total_bids", type: "quantitative" },
      color: { field: "device", type: "nominal", limit: 10 },
    },
    structured: common({
      title: "Total Bids split by Device",
      data: DEVICES.map((device, i) => ({
        device,
        total_bids: [312400, 201300, 48900][i],
      })),
      fields: pickFields("device", "total_bids"),
      domain_values: { device: DEVICES },
      totals: { total_bids: 562600 },
    }),
  }),
);

write(
  "funnel_chart",
  fixture({
    name: "Funnel chart",
    description: "Bids by campaign stage.",
    expect: "canvas",
    chartType: "funnel_chart",
    spec: {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      measure: { field: "total_bids", type: "quantitative" },
      stage: { field: "campaign_stage", type: "nominal", limit: 15 },
      mode: "width",
      percentMode: "top",
    },
    structured: common({
      title: "Total Bids funnel by Campaign Stage",
      data: STAGES.map((campaign_stage, i) => ({
        campaign_stage,
        total_bids: [562600, 148200, 61300, 22400][i],
      })),
      fields: pickFields("campaign_stage", "total_bids"),
      domain_values: { campaign_stage: STAGES },
    }),
  }),
);

const heatmapRows = [];
PUBLISHERS.forEach((publisher) => {
  DEVICES.forEach((device) => {
    heatmapRows.push({
      publisher,
      device,
      win_rate: round(0.08 + rand() * 0.3, 3),
    });
  });
});
write(
  "heatmap",
  fixture({
    name: "Heatmap",
    description:
      "Win rate per publisher and device. No domain_values: the widget falls back to first appearance.",
    expect: "canvas",
    chartType: "heatmap",
    spec: {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      x: { field: "publisher", type: "nominal", limit: 40 },
      y: { field: "device", type: "nominal", limit: 40 },
      color: { field: "win_rate", type: "quantitative" },
    },
    structured: common({
      title: "Win Rate by Publisher and Device",
      data: heatmapRows,
      fields: pickFields("publisher", "device", "win_rate"),
    }),
  }),
);

write(
  "combo_chart",
  fixture({
    name: "Combo chart",
    description: "Daily bids as bars with the win rate as a line.",
    expect: "canvas",
    chartType: "combo_chart",
    spec: {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      time_grain: "TIME_GRAIN_DAY",
      x: { field: "__time", type: "temporal" },
      y1: { field: "total_bids", type: "quantitative", mark: "bar" },
      y2: { field: "win_rate", type: "quantitative", mark: "line" },
      color: { field: "rill_measures", type: "nominal" },
    },
    structured: common({
      title: "Total Bids & Win Rate over Time",
      time_grain: "TIME_GRAIN_DAY",
      data: DAYS.map((__time, i) => ({
        __time,
        total_bids: dailyBids[i] || 43800,
        win_rate: round(0.12 + rand() * 0.1, 3),
      })),
      fields: pickFields("__time", "total_bids", "win_rate"),
      domain_values: { rill_measures: ["total_bids", "win_rate"] },
    }),
  }),
);

const scatterRows = [];
ADVERTISERS.forEach((advertiser_name, i) => {
  DEVICES.forEach((device) => {
    scatterRows.push({
      advertiser_name,
      device,
      total_bids: Math.round(barBids[i] * (0.2 + rand() * 0.3)),
      win_rate: round(0.05 + rand() * 0.35, 3),
      impressions: int(5000, 60000),
    });
  });
});
write(
  "scatter_plot",
  fixture({
    name: "Scatter plot",
    description:
      "Bids against win rate per advertiser and device, sized by impressions.",
    expect: "canvas",
    chartType: "scatter_plot",
    spec: {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      x: { field: "total_bids", type: "quantitative" },
      y: { field: "win_rate", type: "quantitative" },
      dimension: { field: "advertiser_name", type: "nominal", limit: 50 },
      color: { field: "device", type: "nominal", limit: 3 },
      size: { field: "impressions", type: "quantitative" },
    },
    structured: common({
      title:
        "Total Bids vs Win Rate for Advertiser Name split by Device sized by Impressions",
      data: scatterRows,
      fields: pickFields(
        "advertiser_name",
        "device",
        "total_bids",
        "win_rate",
        "impressions",
      ),
      domain_values: { device: DEVICES },
    }),
  }),
);

// `truncated` alone means a top-N cap applied; the x limit is the one the backend applied.
write(
  "truncated",
  fixture({
    name: "Truncated bar chart",
    description:
      "The server capped the x values and tells the model how to narrow the chart.",
    expect: "canvas",
    chartType: "bar_chart",
    spec: barSpec,
    structured: {
      ...barStructured,
      truncated: true,
      truncation_warning:
        "Only the first 50 of 120 advertisers are shown. Lower the x limit or add a filter to see the rest.",
    },
  }),
);

// Row truncation keeps the newest buckets: the first three days were cut, not empty.
write(
  "truncated_rows",
  fixture({
    name: "Row-truncated line chart",
    description:
      "The row limit cut the oldest days. The widget does not zero-fill before Sep 12 and says the most recent periods are shown.",
    expect: "canvas",
    chartType: "line_chart",
    spec: lineSpec,
    structured: common({
      title: "Total Bids over Time",
      time_grain: "TIME_GRAIN_DAY",
      data: DAYS.slice(3).map((__time, i) => ({
        __time,
        total_bids: dailyBids[i + 3] || 43800,
      })),
      fields: pickFields("__time", "total_bids"),
      truncated: true,
      rows_truncated: true,
      truncation_warning:
        "The chart data is incomplete: the data was limited to 4 rows. To see the complete picture, use a coarser time_grain, lower the field limits, or add filters.",
    }),
  }),
);

// A zero-row result still carries `data`, as an empty array; `domain_values` is omitempty and absent.
write(
  "empty",
  fixture({
    name: "Empty result",
    description:
      "The query returned no rows: `data` is empty and the server omits `domain_values`.",
    expect: "empty",
    chartType: "bar_chart",
    spec: barSpec,
    structured: common({
      title: "Total Bids per Advertiser Name",
      data: [],
      fields: pickFields("advertiser_name", "total_bids"),
    }),
  }),
);

write("no_data", {
  name: "Result without data",
  description:
    "The host did not forward the data section (e.g. Claude wrote a large result to a file). The widget shows an error with Open in Rill.",
  expect: "error",
  arguments: { chart_type: "bar_chart", spec: barSpec },
  result: {
    content: [
      { type: "text", text: "**Bar: Total Bids per Advertiser Name**" },
    ],
    structuredContent: {
      widget: "chart",
      chart_type: "bar_chart",
      spec: barSpec,
      message: "Chart created successfully: bar_chart",
      title: "Total Bids per Advertiser Name",
      open_url: OPEN_URL,
    },
  },
});

write("unsupported_widget", {
  name: "Unsupported widget mode",
  description:
    "A newer server sent a widget mode this bundle does not implement.",
  expect: "error",
  arguments: { type: "canvas", name: "overview" },
  result: {
    content: [{ type: "text", text: "Dashboard: Overview" }],
    structuredContent: {
      widget: "dashboard",
      type: "canvas",
      name: "overview",
      title: "Overview",
      open_url: "https://ui.rilldata.com/demo/bids/canvas/overview",
    },
  },
});

write("tool_error", {
  name: "Tool error",
  description: "The tool call itself failed.",
  expect: "error",
  arguments: {
    chart_type: "bar_chart",
    spec: { ...barSpec, metrics_view: "missing_metrics" },
  },
  result: {
    isError: true,
    content: [
      { type: "text", text: 'metrics view "missing_metrics" not found' },
    ],
  },
});

// ---------------------------------------------------------------------------------------------
// Flint results: `create_chart` called with Flint's `chart_spec` and Rill's data binding in the
// `data` slot (see .claude/plans/flint-implementation-contract.md). The server echoes the
// effective input (limits filled in) and derives the semantic types and display names.

/** Mirrors the server's derivation: format preset first, then the data type. */
function semanticTypeOf(field, grain) {
  if (field.kind === "time") {
    return (
      {
        TIME_GRAIN_YEAR: "Year",
        TIME_GRAIN_QUARTER: "YearQuarter",
        TIME_GRAIN_MONTH: "YearMonth",
        TIME_GRAIN_WEEK: "YearWeek",
        TIME_GRAIN_DAY: "Date",
      }[grain] ?? "DateTime"
    );
  }
  if (field.kind === "dimension") return "Category";
  if (
    field.format_preset === "currency_usd" ||
    field.format_preset === "currency_eur"
  )
    return "Amount";
  if (field.format_preset === "percentage") return "Percentage";
  if (field.format_d3?.includes("$")) return "Amount";
  if (field.format_d3?.includes("%")) return "Percentage";
  if (["INT8", "INT16", "INT32", "INT64", "INT128"].includes(field.type))
    return "Count";
  return "Number";
}

function semanticsFor(fields, grain, rows) {
  const semantic_types = {};
  const field_display_names = {};
  for (const [name, field] of Object.entries(fields)) {
    semantic_types[name] = semanticTypeOf(field, grain);
    field_display_names[name] =
      field.kind === "time" ? "Time" : field.display_name;
    const previous = `${name}_prev`;
    if (field.kind === "measure" && rows.some((row) => previous in row)) {
      semantic_types[previous] = semantic_types[name];
      field_display_names[previous] = `${field.display_name} (previous period)`;
    }
  }
  return { semantic_types, field_display_names };
}

const BINDING = { metrics_view: "bids_metrics", time_range: timeRange };

function flintFixture({
  name,
  description,
  expect,
  chartSpec,
  data = BINDING,
  structured,
}) {
  const rows = structured.data ?? [];
  const structuredContent = {
    widget: "chart",
    chart_type: chartSpec.chartType,
    input: { chart_spec: chartSpec, data },
    message: `Chart created successfully: ${chartSpec.chartType}`,
    ...structured,
    ...semanticsFor(structured.fields ?? {}, data.time_grain, rows),
  };
  return {
    name,
    description,
    expect,
    arguments: { chart_spec: chartSpec, data },
    result: {
      content: [
        {
          type: "text",
          text: textFallback(chartSpec.chartType, structuredContent),
        },
      ],
      structuredContent,
    },
  };
}

const flintBarSpec = {
  chartType: "Bar Chart",
  encodings: {
    x: {
      field: "advertiser_name",
      sortBy: "y",
      sortOrder: "descending",
      limit: 10,
    },
    y: "total_bids",
  },
};
const flintBarStructured = common({
  title: "Total Bids by Advertiser Name",
  data: ADVERTISERS.map((advertiser_name, i) => ({
    advertiser_name,
    total_bids: barBids[i],
  })),
  fields: pickFields("advertiser_name", "total_bids"),
  domain_values: { advertiser_name: ADVERTISERS },
});
write(
  "flint_bar_chart",
  flintFixture({
    name: "Flint bar chart",
    description:
      "Canonical Flint result: top advertisers by total bids, with the server's effective limit on x.",
    expect: "canvas",
    chartSpec: flintBarSpec,
    structured: flintBarStructured,
  }),
);

write(
  "flint_line_comparison",
  flintFixture({
    name: "Flint line chart with comparison",
    description:
      "Daily total bids with the previous week as `_prev` columns; the widget folds them into a second series.",
    expect: "canvas",
    chartSpec: {
      chartType: "Line Chart",
      encodings: { x: "__time", y: "total_bids" },
    },
    data: {
      ...BINDING,
      comparison_time_range: {
        start: COMPARISON_RANGE.start,
        end: COMPARISON_RANGE.end,
      },
      time_grain: "TIME_GRAIN_DAY",
    },
    structured: common({
      title: "Total Bids over Time",
      time_grain: "TIME_GRAIN_DAY",
      data: DAYS.map((__time, i) => ({
        __time,
        total_bids: dailyBids[i] || 43800,
        total_bids_prev: Math.round(
          (dailyBids[i] || 43800) * (0.8 + rand() * 0.3),
        ),
      })),
      fields: pickFields("__time", "total_bids"),
      resolved_comparison_time_range: COMPARISON_RANGE,
    }),
  }),
);

write(
  "flint_waterfall",
  flintFixture({
    name: "Flint waterfall chart",
    description:
      "Spend change per publisher; Flint colors the bars by sign and adds the totals.",
    expect: "canvas",
    chartSpec: {
      chartType: "Waterfall Chart",
      subtitle: "Week over week, USD",
      encodings: { x: { field: "publisher", limit: 10 }, y: "spend_change" },
      chartProperties: { totals: "both" },
    },
    structured: common({
      title: "Spend Change by Publisher",
      data: PUBLISHERS.map((publisher, i) => ({
        publisher,
        spend_change: [1240.5, -410.25, 655.0, -980.75, 310.4][i],
      })),
      fields: pickFields("publisher", "spend_change"),
      domain_values: { publisher: PUBLISHERS },
    }),
  }),
);

write(
  "flint_pie",
  flintFixture({
    name: "Flint donut chart",
    description: "Bid share per device as a Pie Chart with an inner radius.",
    expect: "canvas",
    chartSpec: {
      chartType: "Pie Chart",
      encodings: { size: "total_bids", color: { field: "device", limit: 10 } },
      chartProperties: { innerRadius: 55 },
    },
    structured: common({
      title: "Total Bids by Device",
      data: DEVICES.map((device, i) => ({
        device,
        total_bids: [312400, 201300, 48900][i],
      })),
      fields: pickFields("device", "total_bids"),
      domain_values: { device: DEVICES },
    }),
  }),
);

write(
  "flint_heatmap",
  flintFixture({
    name: "Flint heatmap",
    description: "Win rate per publisher and device.",
    expect: "canvas",
    chartSpec: {
      chartType: "Heatmap",
      encodings: {
        x: { field: "publisher", limit: 40 },
        y: { field: "device", limit: 40 },
        color: "win_rate",
      },
    },
    structured: common({
      title: "Win Rate by Publisher and Device",
      data: heatmapRows,
      fields: pickFields("publisher", "device", "win_rate"),
      domain_values: { publisher: PUBLISHERS, device: DEVICES },
    }),
  }),
);

const MONTHS = ["2023-07-01T00:00:00Z", "2023-08-01T00:00:00Z"];
const slopeRows = [];
ADVERTISERS.slice(0, 4).forEach((advertiser_name, i) => {
  MONTHS.forEach((__time, month) => {
    slopeRows.push({
      __time,
      advertiser_name,
      total_bids: barBids[i] + month * int(-30000, 30000),
    });
  });
});
write(
  "flint_slope",
  flintFixture({
    name: "Flint slope chart",
    description:
      "Two months of bids per advertiser at month grain (YearMonth semantics).",
    expect: "canvas",
    chartSpec: {
      chartType: "Slope Chart",
      encodings: {
        x: "__time",
        y: "total_bids",
        color: { field: "advertiser_name", limit: 10 },
      },
    },
    data: {
      metrics_view: "bids_metrics",
      time_range: {
        start: "2023-07-01T00:00:00Z",
        end: "2023-09-01T00:00:00Z",
      },
      time_grain: "TIME_GRAIN_MONTH",
    },
    structured: common({
      title: "Total Bids over Time by Advertiser Name",
      time_grain: "TIME_GRAIN_MONTH",
      resolved_time_range: {
        start: "2023-07-01T00:00:00Z",
        end: "2023-09-01T00:00:00Z",
      },
      data: slopeRows,
      fields: pickFields("__time", "advertiser_name", "total_bids"),
      domain_values: { advertiser_name: ADVERTISERS.slice(0, 4) },
    }),
  }),
);

write(
  "flint_grouped_bar",
  flintFixture({
    name: "Flint grouped bar chart",
    description:
      "Impressions per publisher with devices side by side (the `group` channel).",
    expect: "canvas",
    chartSpec: {
      chartType: "Grouped Bar Chart",
      encodings: {
        x: { field: "publisher", limit: 10 },
        y: "impressions",
        group: { field: "device", limit: 10 },
      },
    },
    structured: common({
      title: "Impressions by Publisher and Device",
      data: stackedRows,
      fields: pickFields("publisher", "device", "impressions"),
      domain_values: { publisher: PUBLISHERS, device: DEVICES },
    }),
  }),
);

write(
  "flint_empty",
  flintFixture({
    name: "Flint empty result",
    description: "A Flint result whose query returned no rows.",
    expect: "empty",
    chartSpec: flintBarSpec,
    structured: common({
      title: "Total Bids by Advertiser Name",
      data: [],
      fields: pickFields("advertiser_name", "total_bids"),
    }),
  }),
);
