import { ListValue, Value, type JsonValue } from "@bufbuild/protobuf";
import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { BaseChart } from "@rilldata/web-common/features/canvas/components/charts/BaseChart";
import type { CanvasChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import {
  CanvasWorkbookExportRequest,
  WorkbookBlock,
  WorkbookColumn,
  WorkbookColumnType,
  WorkbookDashboardSheet,
  WorkbookDataSheet,
  WorkbookGeometry,
  WorkbookImageBlock,
  WorkbookKpiBlock,
  WorkbookKpiValue,
  WorkbookMeta,
  WorkbookNoteBlock,
  WorkbookTableBlock,
  WorkbookTextBlock,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import { marked, type Token, type Tokens } from "marked";
import { get } from "svelte/store";
import type { MeasuredBlock } from "../pdf/capture";
import { hasNativeChart, mapChart } from "./chart-mapping";
import type { WorkbookHeader } from "./header";
import { excelNumberFormat, PERCENT_FORMAT } from "./number-format";
import type {
  ComponentExportPayload,
  CustomChartExportResult,
  ExportColumn,
  ExportTable,
  KpiExportValue,
  LeaderboardExportTable,
} from "./payload";
import { tableExportTable } from "./table-rows";
import { excelDateFormat, localIsoInZone, timeColumnType } from "./time-format";
import type { XlsxChartMode } from "./types";

export interface CollectInput {
  canvasName: string;
  canvasEntity: CanvasEntity;
  // Component cards of the export render, measured; tab label bands excluded.
  blocks: MeasuredBlock[];
  // PNG snapshots by block id, for the blocks that need one (see needsImage).
  images: Map<string, Uint8Array<ArrayBuffer>>;
  chartMode: XlsxChartMode;
  header: WorkbookHeader;
  generatedAt: string;
  dashboardUrl: string;
  // Budget for the rows a table loads for its data sheet; see table-rows.ts.
  timeoutMs?: number;
}

// Vertical gap between canvas rows on a dashboard sheet, in CSS pixels: two
// grid rows, so stacked blocks read as separate items.
export const ROW_GAP_PX = 40;

// Whether a component is exported as a PNG snapshot: charts in image mode and
// chart types without a native Excel equivalent always; custom charts, maps and
// images always; components without a payload (e.g. referenced components) as
// a fallback, except static ones that would be captured as blank cards.
export function needsImage(
  component: BaseCanvasComponent,
  payload: ComponentExportPayload | undefined,
  chartMode: XlsxChartMode,
): boolean {
  if (!payload) {
    return component.type !== "markdown" && component.type !== "image";
  }
  switch (payload.kind) {
    case "chart":
      return chartMode === "image" || !hasNativeChart(payload.chartType);
    // In image mode a KPI block is a snapshot too, so its sparklines come out as Rill draws them.
    case "kpi":
      return chartMode === "image";
    case "customChart":
    case "map":
    case "imageUrl":
      return true;
    default:
      return false;
  }
}

// Builds the workbook request: one dashboard sheet for the canvas's own rows
// plus one per exported tab, each mirroring the export render's geometry with
// the rows stacked (so a tab group leaves no gap), and the data sheets behind
// every chart, table, leaderboard and KPI. Async because a flat table may load
// the rows its data sheet needs.
export async function buildWorkbookRequest(
  input: CollectInput,
): Promise<CanvasWorkbookExportRequest> {
  const components = get(input.canvasEntity.componentsStore);
  const dataSheets: WorkbookDataSheet[] = [];
  const kpis: KpiRows = { rows: [], formats: [] };
  const usedIds = new Set<string>();

  const sheetFor = async (blocks: MeasuredBlock[], title: string) => {
    const positions = stackRows(blocks);
    const sheet = new WorkbookDashboardSheet({ title });
    for (const measured of blocks) {
      const component = components.get(measured.id);
      const position = positions.get(measured.id) ?? {
        x: measured.xPx,
        y: measured.yPx,
      };
      const geometry = new WorkbookGeometry({
        xPx: Math.max(0, Math.round(position.x)),
        yPx: Math.max(0, Math.round(position.y)),
        widthPx: Math.max(1, Math.round(measured.widthPx)),
        heightPx: Math.max(1, Math.round(measured.heightPx)),
      });
      const block = component
        ? await blockFor(component, input, dataSheets, kpis, usedIds)
        : { title: "", subtitle: "", content: note("No valid component") };
      sheet.blocks.push(
        new WorkbookBlock({
          component: measured.id,
          title: block.title,
          subtitle: block.subtitle,
          geometry,
          content: block.content,
        }),
      );
    }
    return sheet;
  };

  const { main, tabs } = groupBySheet(input.blocks);
  const tabTitles = exportedTabTitles(input.canvasEntity);
  const dashboardSheets = [await sheetFor(main, input.header.title)];
  const tabNotes: string[] = [];
  for (const [key, blocks] of tabs) {
    const title = tabTitles.get(key) ?? key;
    tabNotes.push(title);
    dashboardSheets.push(await sheetFor(blocks, title));
  }
  if (tabNotes.length) {
    dashboardSheets[0].notes.push(
      `Tabs: ${tabNotes.join(", ")} (see the sheets of the same name)`,
    );
  }

  if (kpis.rows.length) {
    // One row per KPI, so the value columns have no format of their own: each
    // row takes its measure's.
    dataSheets.push(
      toDataSheet(uniqueId("kpis", usedIds), "KPIs", {
        columns: [
          {
            name: "measure",
            displayName: "Measure",
            type: WorkbookColumnType.STRING,
          },
          {
            name: "value",
            displayName: "Value",
            type: WorkbookColumnType.NUMBER,
          },
          {
            name: "comparison",
            displayName: "Comparison period",
            type: WorkbookColumnType.STRING,
          },
          {
            name: "comparison_value",
            displayName: "Comparison value",
            type: WorkbookColumnType.NUMBER,
          },
          {
            name: "delta",
            displayName: "Change",
            type: WorkbookColumnType.NUMBER,
          },
          {
            name: "delta_percent",
            displayName: "% change",
            type: WorkbookColumnType.NUMBER,
            numberFormat: PERCENT_FORMAT,
          },
        ],
        rows: kpis.rows,
        rowNumberFormats: kpis.formats,
      }),
    );
  }

  return new CanvasWorkbookExportRequest({
    canvas: input.canvasName,
    meta: new WorkbookMeta({
      title: input.header.title,
      generatedAt: input.generatedAt,
      dashboardUrl: input.dashboardUrl,
      timeRangeText: input.header.timeRangeText,
      comparisonText: input.header.comparisonText,
      timeZone: input.header.timeZone,
      filterLines: input.header.filterLines,
    }),
    dashboardSheets,
    dataSheets,
  });
}

type SheetKey = string;
const MAIN_SHEET: SheetKey = "";

function sheetKey(block: MeasuredBlock): SheetKey {
  return block.tabGroup && block.tab
    ? `${block.tabGroup}\u0000${block.tab}`
    : MAIN_SHEET;
}

// Splits the blocks into the main sheet's and one group per tab, in document order.
export function groupBySheet(blocks: MeasuredBlock[]): {
  main: MeasuredBlock[];
  tabs: Map<SheetKey, MeasuredBlock[]>;
} {
  const main: MeasuredBlock[] = [];
  const tabs = new Map<SheetKey, MeasuredBlock[]>();
  for (const block of blocks) {
    if (block.isTabLabel) continue;
    const key = sheetKey(block);
    if (key === MAIN_SHEET) {
      main.push(block);
      continue;
    }
    const group = tabs.get(key) ?? [];
    group.push(block);
    tabs.set(key, group);
  }
  return { main, tabs };
}

// Lays the rows of one sheet out top to bottom with a fixed gap, keeping each
// block's offset within its row, so the sheet has no gap where other sheets'
// content (a tab group) sat in the export render.
export function stackRows(
  blocks: MeasuredBlock[],
  gapPx = ROW_GAP_PX,
): Map<string, { x: number; y: number }> {
  const rows = new Map<number, MeasuredBlock[]>();
  for (const block of blocks) {
    const row = rows.get(block.rowIndex) ?? [];
    row.push(block);
    rows.set(block.rowIndex, row);
  }
  const positions = new Map<string, { x: number; y: number }>();
  let cursor = 0;
  for (const rowIndex of Array.from(rows.keys()).sort((a, b) => a - b)) {
    const row = rows.get(rowIndex)!;
    const top = Math.min(...row.map((b) => b.yPx));
    const bottom = Math.max(...row.map((b) => b.yPx + b.heightPx));
    for (const block of row) {
      positions.set(block.id, { x: block.xPx, y: cursor + (block.yPx - top) });
    }
    cursor += bottom - top + gapPx;
  }
  return positions;
}

// The display names of the tabs, keyed like sheetKey, in strip order.
function exportedTabTitles(canvasEntity: CanvasEntity): Map<SheetKey, string> {
  const titles = new Map<SheetKey, string>();
  for (const block of get(canvasEntity.layout)) {
    if (block.kind !== "tab-group") continue;
    for (const tab of get(block.group.tabs)) {
      titles.set(
        `${block.group.name}\u0000${tab.name}`,
        tab.displayName || tab.name,
      );
    }
  }
  return titles;
}

// The rows of the KPIs data sheet, with the number format of each row's measure.
interface KpiRows {
  rows: unknown[][];
  formats: string[];
}

interface BlockContent {
  title: string;
  // The component's description.
  subtitle: string;
  content: WorkbookBlock["content"];
}

async function blockFor(
  component: BaseCanvasComponent,
  input: CollectInput,
  dataSheets: WorkbookDataSheet[],
  kpis: KpiRows,
  usedIds: Set<string>,
): Promise<BlockContent> {
  const payload = get(component.exportPayload);
  const spec = get(component.specStore) as
    | { title?: string; description?: string }
    | undefined;
  const image = input.images.get(component.id);
  const timeZone = input.header.timeZone || "UTC";
  let title = spec?.title?.trim() ?? "";
  const subtitle = spec?.description?.trim() ?? "";

  const addSheet = (
    suffix: string,
    sheetTitle: string,
    table: ExportTable,
  ): string => {
    const id = uniqueId(
      suffix ? `${component.id}-${suffix}` : component.id,
      usedIds,
    );
    dataSheets.push(toDataSheet(id, sheetTitle || component.id, table));
    return id;
  };
  const imageOr = (fallback: string): WorkbookBlock["content"] =>
    image
      ? { case: "image", value: new WorkbookImageBlock({ png: image }) }
      : note(fallback);

  if (!payload) {
    return { title, subtitle, content: imageOr("No data") };
  }

  switch (payload.kind) {
    case "chart": {
      if (!title) {
        title =
          (component as BaseChart<CanvasChartSpec>).chartTitle?.(
            payload.data.fields,
          ) ?? "";
      }
      const timeGrain = get(component.timeAndFilterStore)?.timeGrain;
      const { table, chart } = mapChart({
        chartType: payload.chartType,
        spec: payload.spec,
        data: payload.data,
        timeGrain,
        timeZone,
      });
      const sheetId = addSheet("", title, table);
      if (chart && !needsImage(component, payload, input.chartMode)) {
        chart.dataSheetId = sheetId;
        return { title, subtitle, content: { case: "chart", value: chart } };
      }
      return { title, subtitle, content: imageOr("No data") };
    }
    case "kpi": {
      for (const value of payload.values) {
        kpis.rows.push([
          value.label,
          value.value,
          value.comparisonLabel ?? null,
          value.comparisonValue,
          value.delta,
          value.deltaPercent,
        ]);
        kpis.formats.push(
          excelNumberFormat(value.measure, [
            value.value,
            value.comparisonValue,
            value.delta,
          ]),
        );
      }
      const trend = kpiTrendTable(payload.values, timeZone);
      const trendDataSheetId = trend
        ? addSheet("trend", `${title || component.id} trend`, trend)
        : "";
      if (needsImage(component, payload, input.chartMode)) {
        return { title, subtitle, content: imageOr("No data") };
      }
      return {
        title,
        subtitle,
        content: {
          case: "kpi",
          value: new WorkbookKpiBlock({
            values: payload.values.map((value, i) =>
              kpiValue(
                value,
                trend && value.trend?.length ? kpiTrendColumn(i) : "",
              ),
            ),
            trendDataSheetId,
          }),
        },
      };
    }
    case "table": {
      const table = await tableExportTable(
        component,
        payload.state,
        payload.config,
        input.timeoutMs,
      );
      const sheetId = addSheet("", title, table);
      return {
        title,
        subtitle,
        content: {
          case: "table",
          value: new WorkbookTableBlock({ dataSheetIds: [sheetId] }),
        },
      };
    }
    case "leaderboard": {
      const ids = payload.tables.map((table) =>
        addSheet(
          table.dimension.name ?? "",
          [title, table.dimension.displayName || table.dimension.name]
            .filter(Boolean)
            .join(" - "),
          leaderboardTable(table),
        ),
      );
      if (!ids.length) return { title, subtitle, content: note("No data") };
      return {
        title,
        subtitle,
        content: {
          case: "table",
          value: new WorkbookTableBlock({ dataSheetIds: ids }),
        },
      };
    }
    case "customChart": {
      payload.results.forEach((result, i) => {
        const table = customChartTable(result, timeZone);
        if (!table) return;
        const suffix = payload.results.length > 1 ? `query ${i + 1}` : "";
        addSheet(
          suffix,
          [title || component.id, suffix].filter(Boolean).join(" "),
          table,
        );
      });
      return { title, subtitle, content: imageOr("No data") };
    }
    case "map": {
      addSheet("", title, mapTable(payload));
      return { title, subtitle, content: imageOr("No data") };
    }
    case "text":
      return {
        title,
        subtitle,
        content: {
          case: "text",
          value: new WorkbookTextBlock({ text: stripMarkdown(payload.text) }),
        },
      };
    case "imageUrl":
      // The snapshot is the image as the canvas shows it, with the URL as its
      // alt text; when the snapshot failed, the block names the URL instead.
      if (image) {
        return {
          title: title || payload.url,
          subtitle,
          content: {
            case: "image",
            value: new WorkbookImageBlock({ png: image }),
          },
        };
      }
      return {
        title,
        subtitle,
        content: {
          case: "text",
          value: new WorkbookTextBlock({ text: `Image: ${payload.url}` }),
        },
      };
    case "error":
      return { title, subtitle, content: note(payload.message || "Error") };
  }
}

function note(text: string): WorkbookBlock["content"] {
  return { case: "note", value: new WorkbookNoteBlock({ text }) };
}

function kpiValue(
  value: KpiExportValue,
  trendColumn: string,
): WorkbookKpiValue {
  return new WorkbookKpiValue({
    label: value.label,
    value: Value.fromJson(jsonCell(value.value)),
    numberFormat: excelNumberFormat(value.measure, [
      value.value,
      value.comparisonValue,
      value.delta,
    ]),
    comparisonLabel: value.comparisonLabel ?? "",
    comparisonValue: Value.fromJson(jsonCell(value.comparisonValue)),
    delta: Value.fromJson(jsonCell(value.delta)),
    deltaPercent: Value.fromJson(jsonCell(value.deltaPercent)),
    trendColumn,
  });
}

function kpiTrendColumn(index: number): string {
  return `kpi_${index}`;
}

// The time series behind a KPI block's sparklines: one row per time bucket
// across every KPI that has a trend, one column per KPI. Undefined when none has one.
export function kpiTrendTable(
  values: KpiExportValue[],
  timeZone: string,
): ExportTable | undefined {
  const withTrend = values
    .map((value, index) => ({ value, index }))
    .filter(({ value }) => value.trend?.length);
  if (!withTrend.length) return undefined;

  const grain = withTrend[0].value.timeGrain;
  const columns: ExportColumn[] = [
    {
      name: "time",
      displayName: "Time",
      type: timeColumnType(grain),
      numberFormat: excelDateFormat(grain),
    },
    ...withTrend.map(({ value, index }) => ({
      name: kpiTrendColumn(index),
      displayName: value.label,
      type: WorkbookColumnType.NUMBER,
      numberFormat: excelNumberFormat(
        value.measure,
        value.trend!.map((point) => point.value),
      ),
    })),
  ];

  const byTime = new Map<string, (number | null)[]>();
  withTrend.forEach(({ value }, column) => {
    for (const point of value.trend!) {
      let row = byTime.get(point.time);
      if (!row) {
        row = withTrend.map(() => null);
        byTime.set(point.time, row);
      }
      row[column] = point.value;
    }
  });
  const rows = Array.from(byTime.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([time, cells]) => [localIsoInZone(time, timeZone), ...cells]);
  return { columns, rows };
}

// The rows of one leaderboard dimension: the dimension, each measure and, when
// the leaderboard compares periods, the previous value and changes per measure.
export function leaderboardTable(table: LeaderboardExportTable): ExportTable {
  const columns: ExportColumn[] = [
    {
      name: "dimension",
      displayName:
        table.dimension.displayName || table.dimension.name || "Value",
      type: WorkbookColumnType.STRING,
    },
  ];
  const measureNames = table.measures.map((m) => m.name ?? "");
  for (const measure of table.measures) {
    const name = measure.name ?? "";
    columns.push({
      name,
      displayName: measure.displayName || name,
      type: WorkbookColumnType.NUMBER,
      numberFormat: excelNumberFormat(
        measure,
        table.rows.map((r) => r.values[name]),
      ),
    });
  }
  const compared = measureNames.filter((name) =>
    table.rows.some((row) => row.prevValues[name] != null),
  );
  for (const name of compared) {
    const measure = table.measures.find((m) => m.name === name);
    const label = measure?.displayName || name;
    const format = excelNumberFormat(
      measure,
      table.rows.map((r) => r.values[name]),
    );
    columns.push(
      {
        name: `${name}_prev`,
        displayName: `${label} (previous)`,
        type: WorkbookColumnType.NUMBER,
        numberFormat: format,
      },
      {
        name: `${name}_delta`,
        displayName: `${label} change`,
        type: WorkbookColumnType.NUMBER,
        numberFormat: format,
      },
      {
        name: `${name}_delta_perc`,
        displayName: `${label} % change`,
        type: WorkbookColumnType.NUMBER,
        numberFormat: PERCENT_FORMAT,
      },
    );
  }
  const rows = table.rows.map((row) => [
    row.dimensionValue,
    ...measureNames.map((name) => row.values[name] ?? null),
    ...compared.flatMap((name) => [
      row.prevValues[name] ?? null,
      row.deltaAbs[name] ?? null,
      row.deltaRels[name] ?? null,
    ]),
  ]);
  return { columns, rows };
}

// The result of one metrics SQL query, typed from its schema.
export function customChartTable(
  result: CustomChartExportResult,
  timeZone: string,
): ExportTable | undefined {
  if (!result.schema?.length) return undefined;
  const columns: ExportColumn[] = result.schema.map((field) => ({
    name: field.name,
    displayName: field.name,
    type: columnTypeForCode(field.type),
  }));
  const rows = (result.data ?? []).map((row) =>
    columns.map((column) => {
      const value = row[column.name];
      return column.type === WorkbookColumnType.DATE ||
        column.type === WorkbookColumnType.TIMESTAMP
        ? localIsoInZone(value, timeZone)
        : (value ?? null);
    }),
  );
  return { columns, rows };
}

function columnTypeForCode(code: string | undefined): WorkbookColumnType {
  if (!code) return WorkbookColumnType.STRING;
  if (/^CODE_(U?INT|FLOAT|DECIMAL)/.test(code))
    return WorkbookColumnType.NUMBER;
  if (code === "CODE_BOOL") return WorkbookColumnType.BOOL;
  if (code === "CODE_TIMESTAMP") return WorkbookColumnType.TIMESTAMP;
  if (code === "CODE_DATE") return WorkbookColumnType.DATE;
  return WorkbookColumnType.STRING;
}

function mapTable(
  payload: Extract<ComponentExportPayload, { kind: "map" }>,
): ExportTable {
  const columns: ExportColumn[] = [
    ...payload.dimensions.map((dimension) => ({
      name: dimension.name ?? "",
      displayName: dimension.displayName || dimension.name || "",
      type: WorkbookColumnType.STRING,
    })),
    ...payload.measures.map((measure) => ({
      name: measure.name ?? "",
      displayName: measure.displayName || measure.name || "",
      type: WorkbookColumnType.NUMBER,
      numberFormat: excelNumberFormat(
        measure,
        payload.rows.map((r) => r[measure.name ?? ""]),
      ),
    })),
  ];
  const rows = payload.rows.map((row) =>
    columns.map((column) => row[column.name] ?? null),
  );
  return { columns, rows };
}

// A data sheet from a table. Column names must be unique on a sheet, and a
// chart can put one measure on two channels (a combo chart's y1 and y2, a
// scatter sized by its own x, a map coloured and sized by one measure), so a
// column named like an earlier one is dropped along with its cells.
export function toDataSheet(
  id: string,
  title: string,
  table: ExportTable,
): WorkbookDataSheet {
  const seen = new Set<string>();
  const kept: number[] = [];
  table.columns.forEach((column, i) => {
    if (seen.has(column.name)) return;
    seen.add(column.name);
    kept.push(i);
  });
  return new WorkbookDataSheet({
    id,
    title,
    columns: kept.map(
      (i) =>
        new WorkbookColumn({
          name: table.columns[i].name,
          displayName: table.columns[i].displayName,
          type: table.columns[i].type,
          numberFormat: table.columns[i].numberFormat ?? "",
        }),
    ),
    rows: table.rows.map((row) =>
      ListValue.fromJson(kept.map((i) => jsonCell(row[i]))),
    ),
    totals: table.totals
      ? ListValue.fromJson(kept.map((i) => jsonCell(table.totals![i])))
      : undefined,
    totalsLast: table.totalsLast ?? false,
    rowNumberFormats: table.rowNumberFormats ?? [],
    moreRows: table.moreRows ?? false,
  });
}

function uniqueId(base: string, used: Set<string>): string {
  let id = base;
  for (let i = 2; used.has(id); i++) id = `${base}-${i}`;
  used.add(id);
  return id;
}

// A cell value as JSON for the request: numbers, strings and booleans as is,
// dates as ISO strings, anything else serialized (or null when it can't be), and null for no value.
export function jsonCell(value: unknown): JsonValue {
  if (value === undefined || value === null) return null;
  if (typeof value === "number") return Number.isFinite(value) ? value : null;
  if (typeof value === "string" || typeof value === "boolean") return value;
  if (typeof value === "bigint") return Number(value);
  if (value instanceof Date) return value.toISOString();
  try {
    return JSON.stringify(value);
  } catch {
    return null;
  }
}

// Reduces markdown to the plain text a cell shows, by walking the token tree of
// the parser the markdown component renders with: headings, emphasis, code,
// links and images keep their text, list items get a marker and raw HTML is
// dropped. Blocks keep a blank line between them where the source had one.
export function stripMarkdown(markdown: string): string {
  return blocksText(marked.lexer(markdown)).trim();
}

function blocksText(tokens: Token[]): string {
  let text = "";
  let separator = "";
  for (const token of tokens) {
    if (token.type === "space") {
      separator = "\n\n";
      continue;
    }
    const block = blockText(token);
    if (block === undefined) continue;
    if (text) text += separator;
    text += block;
    separator = token.raw.endsWith("\n\n") ? "\n\n" : "\n";
  }
  return text;
}

// The text of a block-level token, or undefined for one that shows no text.
function blockText(token: Token): string | undefined {
  switch (token.type) {
    case "html":
    case "hr":
    case "def":
      return undefined;
    case "blockquote":
      return blocksText((token as Tokens.Blockquote).tokens);
    case "list": {
      const list = token as Tokens.List;
      const start = Number(list.start) || 1;
      return list.items
        .map(
          (item, i) =>
            (list.ordered ? `${start + i}. ` : "• ") + blocksText(item.tokens),
        )
        .join("\n");
    }
    case "code":
      return (token as Tokens.Code).text;
    case "table": {
      const table = token as Tokens.Table;
      return [table.header, ...table.rows]
        .map((cells) =>
          cells.map((cell) => inlineText(cell.tokens)).join(" | "),
        )
        .join("\n");
    }
    default:
      return inlineText([token]);
  }
}

function inlineText(tokens: Token[]): string {
  let text = "";
  for (const token of tokens) {
    const { tokens: children, text: own } = token as {
      tokens?: Token[];
      text?: string;
    };
    switch (token.type) {
      case "html":
        break;
      case "br":
        text += "\n";
        break;
      case "codespan":
        text += own ?? "";
        break;
      default:
        text += children ? inlineText(children) : decodeEntities(own ?? "");
    }
  }
  return text;
}

const HTML_ENTITIES: Record<string, string> = {
  amp: "&",
  lt: "<",
  gt: ">",
  quot: '"',
  apos: "'",
  nbsp: " ",
};

// Markdown text may spell characters as HTML entities; a cell shows the characters.
function decodeEntities(text: string): string {
  return text.replace(
    /&(#\d+|#x[0-9a-f]+|[a-z]+);/gi,
    (entity, name: string) => {
      if (name[0] !== "#") return HTML_ENTITIES[name.toLowerCase()] ?? entity;
      const code =
        name[1].toLowerCase() === "x"
          ? parseInt(name.slice(2), 16)
          : parseInt(name.slice(1), 10);
      return code > 0 && code <= 0x10ffff ? String.fromCodePoint(code) : entity;
    },
  );
}
