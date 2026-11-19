// Generates the fixture files under harness/fixtures/. Each fixture is one tool call:
// the `create_chart` arguments (Flint's `chart_spec` with Rill's data binding in the `data` slot)
// and the result the host forwards to the widget, shaped like `CreateChartResult` in
// runtime/ai/create_chart.go: the server echoes the effective input (limits filled in) and derives
// the semantic types and display names. `bar_chart.json` is the canonical example.
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
  win_rate: {
    name: "win_rate",
    kind: "measure",
    display_name: "Win Rate",
    type: "FLOAT64",
    format_preset: "percentage",
    lower_is_better: false,
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

function textFallback(chartType, structured) {
  const {
    title,
    data = [],
    fields = {},
    resolved_time_range: range,
  } = structured;
  const header = `**${chartType}: ${title}**${
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

const timeRange = { start: RANGE.start, end: RANGE.end };
const BINDING = { metrics_view: "bids_metrics", time_range: timeRange };

function fixture({
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

// Canonical bar chart: top advertisers by total bids, with the server's effective limit on x.
const barBids = [128430, 99210, 87650, 64320, 51200, 40880, 28100, 19750];
const barSpec = {
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
const barStructured = common({
  title: "Total Bids by Advertiser Name",
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
      "Canonical result: top advertisers by total bids, with the server's effective limit on x.",
    expect: "canvas",
    chartSpec: barSpec,
    structured: barStructured,
  }),
);

// Daily series with one missing bucket (Sep 12) that the widget zero-fills.
const dailyBids = [41200, 45800, 39900, 0, 47300, 52100, 36400];
const lineSpec = {
  chartType: "Line Chart",
  encodings: { x: "__time", y: "total_bids" },
};
const lineBinding = { ...BINDING, time_grain: "TIME_GRAIN_DAY" };
write(
  "line_chart",
  fixture({
    name: "Line chart",
    description:
      "Daily total bids; Sep 12 is missing from the rows and gets zero-filled.",
    expect: "canvas",
    chartSpec: lineSpec,
    data: lineBinding,
    structured: common({
      title: "Total Bids over Time",
      time_grain: "TIME_GRAIN_DAY",
      data: DAYS.flatMap((__time, i) =>
        dailyBids[i] ? [{ __time, total_bids: dailyBids[i] }] : [],
      ),
      fields: pickFields("__time", "total_bids"),
    }),
  }),
);

write(
  "line_comparison",
  fixture({
    name: "Line chart with comparison",
    description:
      "Daily total bids with the previous week as `_prev` columns; the widget folds them into a second series.",
    expect: "canvas",
    chartSpec: lineSpec,
    data: {
      ...lineBinding,
      comparison_time_range: {
        start: COMPARISON_RANGE.start,
        end: COMPARISON_RANGE.end,
      },
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

write(
  "waterfall",
  fixture({
    name: "Waterfall chart",
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
  "pie",
  fixture({
    name: "Donut chart",
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
  "heatmap",
  fixture({
    name: "Heatmap",
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
  "slope",
  fixture({
    name: "Slope chart",
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
  "grouped_bar",
  fixture({
    name: "Grouped bar chart",
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

// `truncated` alone means a top-N cap applied; the x limit is the one the backend applied.
write(
  "truncated",
  fixture({
    name: "Truncated bar chart",
    description:
      "The server capped the x values and tells the model how to narrow the chart.",
    expect: "canvas",
    chartSpec: barSpec,
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
    chartSpec: lineSpec,
    data: lineBinding,
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
    chartSpec: barSpec,
    structured: common({
      title: "Total Bids by Advertiser Name",
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
  arguments: { chart_spec: barSpec, data: BINDING },
  result: {
    content: [
      { type: "text", text: "**Bar Chart: Total Bids by Advertiser Name**" },
    ],
    structuredContent: {
      widget: "chart",
      chart_type: "Bar Chart",
      input: { chart_spec: barSpec, data: BINDING },
      message: "Chart created successfully: Bar Chart",
      title: "Total Bids by Advertiser Name",
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
    chart_spec: barSpec,
    data: { ...BINDING, metrics_view: "missing_metrics" },
  },
  result: {
    isError: true,
    content: [
      { type: "text", text: 'metrics view "missing_metrics" not found' },
    ],
  },
});
