# Pie Chart "Other" Grouping Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add dynamic "Other" slice grouping and custom tooltips with percentages to Canvas pie/donut charts.

**Architecture:** Data transformation groups small slices into "Other" between the API response and Vega-Lite spec generation. A custom tooltip formatter replaces native Vega tooltips to show percentages and an "Other" breakdown. The "Other" slice gets conditional Vega-Lite styling (muted fill, dashed stroke).

**Tech Stack:** TypeScript, Svelte 4, Vega-Lite, Tailwind CSS, TanStack Query

---

## File Structure

**New files:**
- `web-common/src/features/components/charts/circular/other-grouping.ts` — `computeVisibleSlices` algorithm + `OtherGroupingResult` type
- `web-common/src/features/components/charts/circular/pie-tooltip-formatter.ts` — custom tooltip HTML formatter for regular and "Other" slices

**Modified files:**
- `web-common/src/features/components/charts/types.ts` — add `showOther` to `NominalFieldConfig`
- `web-common/src/features/components/charts/circular/CircularChartProvider.ts` — integrate grouping, store `otherItems` and `total` metadata
- `web-common/src/features/components/charts/circular/pie.ts` — accept pre-grouped data, conditional "Other" styling, remove native tooltip
- `web-common/src/features/components/charts/Chart.svelte` — pass custom `tooltipFormatter` for pie/donut charts
- `web-common/src/features/canvas/components/charts/variants/CircularChart.ts` — expose `showOther` in `chartInputParams`

---

### Task 1: Add `showOther` to NominalFieldConfig

**Files:**
- Modify: `web-common/src/features/components/charts/types.ts:144-153`

- [ ] **Step 1: Add `showOther` property to `NominalFieldConfig`**

In `web-common/src/features/components/charts/types.ts`, add `showOther` to the `NominalFieldConfig` interface:

```typescript
interface NominalFieldConfig {
  sort?: ChartSortDirection;
  limit?: number;
  showNull?: boolean;
  showOther?: boolean;
  labelAngle?: number;
  legendOrientation?: ChartLegend;
  colorMapping?: ColorMapping;
  /** Explicit dimension values to use (skips topN query) */
  values?: string[];
}
```

- [ ] **Step 2: Commit**

```bash
git add web-common/src/features/components/charts/types.ts
git commit -m "feat: add showOther property to NominalFieldConfig"
```

---

### Task 2: Create the `computeVisibleSlices` algorithm

**Files:**
- Create: `web-common/src/features/components/charts/circular/other-grouping.ts`

- [ ] **Step 1: Create `other-grouping.ts` with types and algorithm**

Create `web-common/src/features/components/charts/circular/other-grouping.ts`:

```typescript
export const OTHER_LABEL = "Other";
export const OTHER_FLAG_FIELD = "__isOther";
const MIN_SLICES_FOR_GROUPING = 6;
const DEFAULT_MAX_OTHER_PERCENT = 0.2;
const DEFAULT_HARD_CAP = 10;
const MAX_OTHER_TOOLTIP_ITEMS = 5;

export interface OtherSliceItem {
  label: string;
  value: number;
}

export interface OtherGroupingResult {
  /** Data rows for Vega-Lite; "Other" row has __isOther: true */
  visibleData: Record<string, unknown>[];
  /** Items grouped into "Other", sorted by value desc; null if no "Other" slice */
  otherItems: OtherSliceItem[] | null;
  /** Total value across ALL items (for percentage calculation) */
  total: number;
}

/**
 * Groups small pie chart slices into an "Other" aggregate.
 *
 * @param data - Raw data rows from the metrics view query
 * @param colorField - The dimension field name used for slice labels
 * @param measureField - The measure field name used for slice values
 * @param options.explicitLimit - If set by the editor in YAML, bypasses dynamic algorithm
 * @param options.showOther - If false, truncate at limit with no "Other" (default: true)
 * @param options.maxOtherPercent - Max fraction of total for "Other" before adding more slices (default: 0.2)
 * @param options.hardCap - Max visible slices before forcing "Other" (default: 10)
 */
export function computeVisibleSlices(
  data: Record<string, unknown>[],
  colorField: string,
  measureField: string,
  options: {
    explicitLimit?: number;
    showOther?: boolean;
    maxOtherPercent?: number;
    hardCap?: number;
  } = {},
): OtherGroupingResult {
  const {
    showOther = true,
    maxOtherPercent = DEFAULT_MAX_OTHER_PERCENT,
    hardCap = DEFAULT_HARD_CAP,
  } = options;

  // Sort by measure value descending
  const sorted = [...data].sort((a, b) => {
    const aVal = Number(a[measureField]) || 0;
    const bVal = Number(b[measureField]) || 0;
    return bVal - aVal;
  });

  const total = sorted.reduce(
    (sum, d) => sum + (Number(d[measureField]) || 0),
    0,
  );

  // If too few items or showOther is false with no explicit limit, return all
  if (sorted.length <= MIN_SLICES_FOR_GROUPING || !showOther) {
    const limit = options.explicitLimit;
    if (!showOther && limit !== undefined && limit < sorted.length) {
      // Truncate without "Other"
      return {
        visibleData: sorted.slice(0, limit),
        otherItems: null,
        total,
      };
    }
    return {
      visibleData: sorted,
      otherItems: null,
      total,
    };
  }

  // Determine how many slices to show
  let visibleCount: number;

  if (options.explicitLimit !== undefined) {
    // Editor set an explicit limit; use it directly
    visibleCount = Math.min(options.explicitLimit, sorted.length);
  } else {
    // Dynamic threshold algorithm
    visibleCount = 0;
    let visibleSum = 0;

    for (const item of sorted) {
      if (visibleCount >= hardCap) break;
      visibleCount++;
      visibleSum += Number(item[measureField]) || 0;
      const remaining = total - visibleSum;
      if (total > 0 && remaining / total <= maxOtherPercent) break;
    }
  }

  // If all items are visible, no "Other" needed
  if (visibleCount >= sorted.length) {
    return {
      visibleData: sorted,
      otherItems: null,
      total,
    };
  }

  const visibleRows = sorted.slice(0, visibleCount);
  const otherRows = sorted.slice(visibleCount);

  const otherItems: OtherSliceItem[] = otherRows.map((row) => ({
    label: String(row[colorField] ?? ""),
    value: Number(row[measureField]) || 0,
  }));

  const otherValue = otherItems.reduce((sum, item) => sum + item.value, 0);

  // Create the "Other" data row
  const otherDataRow: Record<string, unknown> = {
    [colorField]: OTHER_LABEL,
    [measureField]: otherValue,
    [OTHER_FLAG_FIELD]: true,
  };

  return {
    visibleData: [...visibleRows, otherDataRow],
    otherItems,
    total,
  };
}

export { MAX_OTHER_TOOLTIP_ITEMS };
```

- [ ] **Step 2: Commit**

```bash
git add web-common/src/features/components/charts/circular/other-grouping.ts
git commit -m "feat: add computeVisibleSlices algorithm for Other grouping"
```

---

### Task 3: Create the custom pie tooltip formatter

**Files:**
- Create: `web-common/src/features/components/charts/circular/pie-tooltip-formatter.ts`

- [ ] **Step 1: Create `pie-tooltip-formatter.ts`**

Create `web-common/src/features/components/charts/circular/pie-tooltip-formatter.ts`:

```typescript
import type { VLTooltipFormatter } from "@rilldata/web-common/components/vega/types";
import {
  MAX_OTHER_TOOLTIP_ITEMS,
  OTHER_FLAG_FIELD,
  OTHER_LABEL,
  type OtherSliceItem,
} from "./other-grouping";

/**
 * Creates a tooltip formatter for pie/donut charts that shows
 * slice name, formatted value, and percentage. For the "Other" slice,
 * shows a mini-leaderboard breakdown of grouped items.
 */
export function createPieTooltipFormatter(options: {
  colorField: string;
  measureField: string;
  total: number;
  otherItems: OtherSliceItem[] | null;
  /** Color map from dimension value → resolved CSS color string */
  colorMap: Map<string, string>;
  /** Format a measure value for display (e.g., "$12,450") */
  formatValue: (value: number) => string;
  /** Resolved CSS color for the muted token (for "Other" dot) */
  mutedColor: string;
}): VLTooltipFormatter {
  const {
    colorField,
    measureField,
    total,
    otherItems,
    colorMap,
    formatValue,
    mutedColor,
  } = options;

  return (value: unknown): string => {
    if (!value || typeof value !== "object") return "";
    const datum = value as Record<string, unknown>;

    const isOther = datum[OTHER_FLAG_FIELD] === true;
    const label = String(datum[colorField] ?? "");
    const measureValue = Number(datum[measureField]) || 0;
    const pct = total > 0 ? ((measureValue / total) * 100).toFixed(1) : "0.0";

    if (isOther && otherItems) {
      return renderOtherTooltip(
        measureValue,
        pct,
        otherItems,
        total,
        formatValue,
        mutedColor,
      );
    }

    const color = colorMap.get(label) || "#888";
    return renderSliceTooltip(label, measureValue, pct, color, formatValue);
  };
}

function renderSliceTooltip(
  label: string,
  value: number,
  pct: string,
  color: string,
  formatValue: (v: number) => string,
): string {
  const dot = `<span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:${color};margin-right:6px;vertical-align:middle;"></span>`;
  const formatted = formatValue(value);
  return `<div style="display:flex;align-items:center;gap:4px;white-space:nowrap;">${dot}<span style="font-weight:500;">${escapeHtml(label)}</span><span style="opacity:0.7;"> · ${formatted} · ${pct}%</span></div>`;
}

function renderOtherTooltip(
  otherTotal: number,
  pct: string,
  items: OtherSliceItem[],
  grandTotal: number,
  formatValue: (v: number) => string,
  mutedColor: string,
): string {
  const dot = `<span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:${mutedColor};margin-right:6px;vertical-align:middle;"></span>`;
  const formatted = formatValue(otherTotal);

  // Header
  let html = `<div style="display:flex;align-items:center;gap:4px;white-space:nowrap;margin-bottom:6px;">${dot}<span style="font-weight:500;">${OTHER_LABEL}</span><span style="opacity:0.7;"> · ${formatted} · ${pct}%</span></div>`;

  // Divider
  html += `<div style="border-bottom:1px solid var(--border, #e5e7eb);margin-bottom:6px;"></div>`;

  // Breakdown rows (max 5)
  const visibleItems = items.slice(0, MAX_OTHER_TOOLTIP_ITEMS);
  html += `<table style="width:100%;border-collapse:collapse;">`;
  for (const item of visibleItems) {
    const itemPct =
      grandTotal > 0 ? ((item.value / grandTotal) * 100).toFixed(1) : "0.0";
    const itemFormatted = formatValue(item.value);
    html += `<tr>`;
    html += `<td style="text-align:left;padding:1px 8px 1px 0;max-width:160px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">${escapeHtml(item.label)}</td>`;
    html += `<td style="text-align:right;padding:1px 4px;white-space:nowrap;font-variant-numeric:tabular-nums;">${itemFormatted}</td>`;
    html += `<td style="text-align:right;padding:1px 0 1px 4px;white-space:nowrap;font-variant-numeric:tabular-nums;opacity:0.7;">${itemPct}%</td>`;
    html += `</tr>`;
  }
  html += `</table>`;

  // Footer "and N more"
  const remaining = items.length - MAX_OTHER_TOOLTIP_ITEMS;
  if (remaining > 0) {
    html += `<div style="font-size:11px;opacity:0.6;font-style:italic;padding-top:4px;">and ${remaining} more</div>`;
  }

  return html;
}

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}
```

- [ ] **Step 2: Commit**

```bash
git add web-common/src/features/components/charts/circular/pie-tooltip-formatter.ts
git commit -m "feat: add custom pie chart tooltip formatter with Other breakdown"
```

---

### Task 4: Update pie.ts to handle "Other" slice styling

**Files:**
- Modify: `web-common/src/features/components/charts/circular/pie.ts`

- [ ] **Step 1: Add conditional color and stroke for "Other" slice, remove native tooltip**

Update `generateVLPieChartSpec` to accept an optional `OtherGroupingResult` and apply conditional styling. The function signature and arc layer encoding need changes:

Add import at top:

```typescript
import { OTHER_FLAG_FIELD } from "./other-grouping";
```

Replace the existing `tooltip` and `color` encoding setup in the arc layer. The key changes:

1. Remove `tooltip` from the arc encoding (we use a custom formatter instead)
2. Add conditional `stroke` and `strokeDash` for "Other" slice
3. Add conditional `strokeWidth` for "Other" slice

Update the `arcLayer` construction (replace the existing one):

```typescript
  const arcLayer: LayerSpec<Field> | UnitSpec<Field> = {
    mark: {
      type: "arc",
      padAngle: 0.01,
      innerRadius: getInnerRadius(config.innerRadius),
      stroke: { expr: `datum.${OTHER_FLAG_FIELD} ? '${resolvedBorderColor}' : null` },
      strokeWidth: { expr: `datum.${OTHER_FLAG_FIELD} ? 1 : 0` },
      strokeDash: { expr: `datum.${OTHER_FLAG_FIELD} ? [4, 3] : [0, 0]` },
    },
    encoding: {
      theta,
      color,
      order,
      tooltip: { value: null },
    },
  };
```

Note: `resolvedBorderColor` needs to be derived from the data's dark mode flag. Add before the arcLayer:

```typescript
  const resolvedBorderColor = data.isDarkMode ? "#374151" : "#e5e7eb";
```

Also, to make "Other" use a muted fill color, update the `createColorEncoding` call result. After `const color = createColorEncoding(config.color, data);`, we need to modify the color scale to include the "Other" entry. This is handled by the domain values from `getChartDomainValues()` which now includes `OTHER_LABEL`, and the color mapping from `getColorMappingForChart` in `Chart.svelte`. We need to ensure the "Other" entry gets a muted color.

The simplest approach: in `CircularChartProvider.getChartDomainValues()`, we already add `OTHER_LABEL` to the domain. The color for "Other" will be assigned by the palette. To override it to use the muted color, we add a color mapping entry for "Other" in the `customColorValues` flow. Actually, the cleaner approach is to handle this in the spec generator by adding "Other" to the color scale domain and range explicitly.

After `const color = createColorEncoding(config.color, data);`, add:

```typescript
  // Override "Other" slice color to use muted fill
  const hasOther = data.data?.some((d) => d[OTHER_FLAG_FIELD] === true);
  if (hasOther && color.scale && "domain" in color.scale && Array.isArray(color.scale.domain)) {
    const mutedColor = data.isDarkMode ? "#374151" : "#e5e7eb";
    if (!color.scale.domain.includes(OTHER_LABEL)) {
      color.scale.domain.push(OTHER_LABEL);
    }
    if (Array.isArray(color.scale.range) && !color.scale.range.includes(mutedColor)) {
      color.scale.range.push(mutedColor);
    }
  }
```

Add the `OTHER_LABEL` import alongside `OTHER_FLAG_FIELD`:

```typescript
import { OTHER_FLAG_FIELD, OTHER_LABEL } from "./other-grouping";
```

- [ ] **Step 2: Commit**

```bash
git add web-common/src/features/components/charts/circular/pie.ts
git commit -m "feat: add Other slice conditional styling in Vega-Lite spec"
```

---

### Task 5: Wire up data grouping and tooltip in Chart.svelte

**Files:**
- Modify: `web-common/src/features/components/charts/Chart.svelte`

- [ ] **Step 1: Apply "Other" grouping to data and pass tooltip formatter for pie/donut charts**

In `Chart.svelte`, the chart data flows through `$chartData` → `chartDataWithTheme` → `generateSpec()` → VegaLiteRenderer. We need to intercept BEFORE `generateSpec` so the color domain includes "Other".

Add imports at the top of the `<script>` block:

```typescript
import { computeVisibleSlices, OTHER_LABEL, type OtherGroupingResult } from "./circular/other-grouping";
import { createPieTooltipFormatter } from "./circular/pie-tooltip-formatter";
import type { VLTooltipFormatter } from "@rilldata/web-common/components/vega/types";
```

Add these reactive blocks BEFORE the existing `$: spec = generateSpec(...)` line (around line 81):

```typescript
  // For pie/donut charts: apply "Other" grouping before spec generation
  $: isPieOrDonut = chartType === "pie_chart" || chartType === "donut_chart";

  $: otherGrouping = isPieOrDonut
    ? applyPieGrouping(chartSpec, chartDataWithTheme.data)
    : null;

  // Inject grouped data + "Other" domain value into chartData for spec generation
  $: chartDataForSpec = otherGrouping
    ? {
        ...chartDataWithTheme,
        data: otherGrouping.visibleData,
        domainValues: injectOtherIntoDomain(
          chartDataWithTheme.domainValues,
          chartSpec,
          otherGrouping,
        ),
      }
    : chartDataWithTheme;
```

Then change the existing `spec` line to use `chartDataForSpec`:

```typescript
  $: spec = generateSpec(chartType, chartSpec, chartDataForSpec);
```

Add the tooltip formatter reactive block AFTER `measureFormatters` and `colorMapping` are defined:

```typescript
  $: pieTooltipFormatter = isPieOrDonut && otherGrouping
    ? buildPieTooltipFormatter(chartSpec, otherGrouping, colorMapping, measureFormatters, isThemeModeDark)
    : undefined;
```

Add the helper functions:

```typescript
  function applyPieGrouping(
    spec: CanvasChartSpec,
    rawData: Record<string, unknown>[],
  ): OtherGroupingResult | null {
    const colorField = "color" in spec ? spec.color?.field : undefined;
    const measureField = "measure" in spec ? spec.measure?.field : undefined;
    if (!colorField || !measureField) return null;

    const showOther = "color" in spec ? spec.color?.showOther !== false : true;
    const explicitLimit = "color" in spec ? spec.color?.limit : undefined;

    return computeVisibleSlices(rawData, colorField, measureField, {
      explicitLimit,
      showOther,
    });
  }

  function injectOtherIntoDomain(
    domainValues: Record<string, string[] | number[] | undefined> | undefined,
    spec: CanvasChartSpec,
    grouping: OtherGroupingResult,
  ): Record<string, string[] | number[] | undefined> | undefined {
    if (!domainValues || !grouping.otherItems) return domainValues;
    const colorField = "color" in spec ? spec.color?.field : undefined;
    if (!colorField || !domainValues[colorField]) return domainValues;

    const existing = domainValues[colorField] as string[];
    const visibleLabels = new Set(
      grouping.visibleData
        .filter((d) => !d.__isOther)
        .map((d) => String(d[colorField])),
    );
    const filtered = existing.filter((v) => visibleLabels.has(v));
    filtered.push(OTHER_LABEL);

    return {
      ...domainValues,
      [colorField]: filtered,
    };
  }

  function buildPieTooltipFormatter(
    spec: CanvasChartSpec,
    grouping: OtherGroupingResult,
    colorMappingArr: ColorMapping,
    formatters: Record<string, (value: number | null | undefined) => string>,
    isDark: boolean,
  ): VLTooltipFormatter | undefined {
    const colorField = "color" in spec ? spec.color?.field : undefined;
    const measureField = "measure" in spec ? spec.measure?.field : undefined;
    if (!colorField || !measureField) return undefined;

    const colorMap = new Map<string, string>(
      (colorMappingArr ?? []).map((m) => [m.value, m.color]),
    );

    const measureFormatter = formatters[sanitizeFieldName(measureField)];
    const formatValue = (v: number) =>
      measureFormatter ? measureFormatter(v) : String(v);

    const mutedColor = isDark ? "#374151" : "#e5e7eb";

    return createPieTooltipFormatter({
      colorField,
      measureField,
      total: grouping.total,
      otherItems: grouping.otherItems,
      colorMap,
      formatValue,
      mutedColor,
    });
  }
```

Update the VegaLiteRenderer section (around line 219-231) to pass grouped data and tooltip formatter:

```svelte
  <VegaLiteRenderer
    bind:viewVL={view}
    canvasDashboard={isCanvas}
    data={{ "metrics-view": chartDataForSpec.data }}
    {themeMode}
    {spec}
    {colorMapping}
    {signalListeners}
    renderer="canvas"
    {expressionFunctions}
    {hasComparison}
    tooltipFormatter={pieTooltipFormatter}
    config={getRillTheme(isThemeModeDark, theme)}
  />
```

Also update the VegaRenderer section (brush mode) similarly:

```svelte
  <VegaRenderer
    bind:view
    data={{ "metrics-view": chartDataForSpec.data }}
    ...
  />
```

- [ ] **Step 2: Commit**

```bash
git add web-common/src/features/components/charts/Chart.svelte
git commit -m "feat: wire up Other grouping and custom tooltip for pie/donut charts"
```

---

### Task 6: Expose `showOther` in CircularChart canvas component

**Files:**
- Modify: `web-common/src/features/canvas/components/charts/variants/CircularChart.ts`

- [ ] **Step 1: Add `showOther` to chartInputParams and default spec**

In `CircularChartComponent.chartInputParams`, the `color` input already has a `limitSelector`. We don't need a separate UI control for `showOther` — it defaults to `true` and can be set via YAML. No UI change needed.

However, update `newComponentSpec` to include `showOther: true` in the default color config so it's explicit:

In the `newComponentSpec` static method, update the return value's `color` field:

```typescript
    return {
      metrics_view: metricsViewName,
      innerRadius: 50,
      color: {
        type: "nominal",
        field: randomDimension,
        limit: DEFAULT_COLOR_LIMIT,
        sort: DEFAULT_SORT,
        showOther: true,
      },
      measure: {
        type: "quantitative",
        field: randomMeasure,
        showTotal: true,
      },
    };
```

- [ ] **Step 2: Commit**

```bash
git add web-common/src/features/canvas/components/charts/variants/CircularChart.ts
git commit -m "feat: add showOther default to CircularChart canvas component"
```

---

### Task 7: Test manually and fix edge cases

- [ ] **Step 1: Build the frontend to check for TypeScript errors**

```bash
cd /Users/eokuma/rill && npm run build -w web-common
```

Fix any type errors that arise. Common issues to watch for:
- `CanvasChartSpec` union type may not have `color` or `measure` on all variants — the `"color" in spec` checks handle this
- The `VLTooltipFormatter` type expects `(value: any) => string` — our formatter matches this
- Vega-Lite `mark` properties like `stroke`, `strokeWidth`, `strokeDash` may need type assertions when using `expr` expressions

- [ ] **Step 2: Fix any build errors and commit**

```bash
git add -u
git commit -m "fix: resolve build errors in pie chart Other grouping"
```

---

### Task 8: Final integration check

- [ ] **Step 1: Verify the complete data flow works end-to-end**

Review the full flow one more time:
1. `CircularChartProvider` fetches top 20 items from API (unchanged)
2. `Chart.svelte` receives data, detects pie/donut type
3. `computeVisibleSlices()` groups data → `OtherGroupingResult`
4. Domain values are patched to include "Other"
5. `generateSpec()` builds Vega spec with "Other" in color domain + muted fill + dashed stroke
6. `createPieTooltipFormatter()` builds custom formatter with breakdown data
7. VegaLiteRenderer renders with grouped data + custom tooltip

- [ ] **Step 2: Run frontend linting**

```bash
cd /Users/eokuma/rill && npm run quality
```

- [ ] **Step 3: Fix any lint issues and commit**

```bash
git add -u
git commit -m "chore: fix lint issues in pie chart Other grouping"
```
