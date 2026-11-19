// Generates runtime/ai/chartquery/flint_catalog.json from the installed flint-chart package:
// the Vega-Lite chart types the MCP `create_chart` tool accepts, with the channels each one takes
// and its tunable properties. The Go side embeds the file to validate chart specs and to describe
// the chart types to the model, so the catalog cannot drift from the compiler the widget renders with.
// Run from the repo root after bumping the flint-chart dependency:
//
//   npm run generate:catalog -w web-mcp-app
//
// The output is committed.
import { createRequire } from "node:module";
import { readFileSync, writeFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { vlAllTemplateDefs } from "flint-chart/vegalite";

// The package does not export its package.json, so it is read next to the resolved entry point.
const require = createRequire(import.meta.url);
const packageRoot = resolve(dirname(require.resolve("flint-chart/vegalite")), "../..");
const { version } = JSON.parse(
  readFileSync(resolve(packageRoot, "package.json"), "utf8"),
);

const outPath = fileURLToPath(
  new URL("../../runtime/ai/chartquery/flint_catalog.json", import.meta.url),
);

// Chart types that need data a metrics view query cannot provide: geographic coordinates
// or region ids, and a single metric tile the canvas already has a native widget for.
const EXCLUDED = new Set(["Map", "Choropleth", "KPI Card"]);

function property(def) {
  const out = { key: def.key, label: def.label ?? def.key, type: def.type };
  if (def.type === "continuous") {
    if (typeof def.min === "number") out.min = def.min;
    if (typeof def.max === "number") out.max = def.max;
    if (typeof def.step === "number") out.step = def.step;
  }
  if (def.type === "discrete" && Array.isArray(def.options)) {
    out.options = def.options.map((option) => ({
      value: option.value,
      label: option.label ?? String(option.value),
    }));
  }
  if (def.defaultValue !== undefined) out.default = def.defaultValue;
  return out;
}

const chartTypes = vlAllTemplateDefs
  .filter((def) => !EXCLUDED.has(def.chart))
  .map((def) => ({
    name: def.chart,
    channels: [...def.channels],
    properties: (def.properties ?? []).map(property),
  }));

const catalog = { flint_version: version, chart_types: chartTypes };
writeFileSync(outPath, `${JSON.stringify(catalog, null, 2)}\n`);
console.log(
  `wrote ${chartTypes.length} chart types (flint-chart ${version}) to ${outPath}`,
);
