import {
  type ChatConfig,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types";
import type {
  RuntimeServiceCompleteBody,
  V1StructType,
} from "@rilldata/web-common/runtime-client";
import { derived, get, readable, writable, type Readable } from "svelte/store";
import type { CellState, NotebookStore } from "./query-store";

const emptyContext: Readable<Partial<RuntimeServiceCompleteBody>> = readable(
  {},
);

/** Set by QueryWorkspace when the notebook is created; read by the chat config. */
export const activeNotebook = writable<NotebookStore | null>(null);

/** Max sample rows to include in the prompt. */
const MAX_SAMPLE_ROWS = 20;
/** Max characters for the results section to avoid bloating the prompt. */
const MAX_RESULTS_CHARS = 4000;

/** Format a results summary (schema + sample rows) for the AI prompt. */
function formatResultsSummary(cell: CellState): string {
  const schema = cell.result?.schema;
  const data = cell.result?.data;
  if (!schema?.fields?.length || !data?.length) return "";

  const columns = schema.fields.map((f) => f.name ?? "?");
  const totalRows = data.length;
  const sample = data.slice(0, MAX_SAMPLE_ROWS);

  // Build a markdown table of sample rows
  const header = `| ${columns.join(" | ")} |`;
  const separator = `| ${columns.map(() => "---").join(" | ")} |`;
  const rows = sample.map(
    (row) =>
      `| ${columns.map((col) => truncateValue(row[col])).join(" | ")} |`,
  );

  let table = [header, separator, ...rows].join("\n");
  if (table.length > MAX_RESULTS_CHARS) {
    table = table.slice(0, MAX_RESULTS_CHARS) + "\n... (truncated)";
  }

  const rowNote =
    totalRows > MAX_SAMPLE_ROWS
      ? `Showing ${MAX_SAMPLE_ROWS} of ${totalRows} rows`
      : `${totalRows} row${totalRows === 1 ? "" : "s"}`;

  return `\n[Query results (${rowNote})]\n${table}\n`;
}

/** Column types summary for the AI. */
function formatSchemaTypes(schema: V1StructType): string {
  if (!schema.fields?.length) return "";
  const types = schema.fields.map(
    (f) => `${f.name ?? "?"}: ${f.type?.code ?? "unknown"}`,
  );
  return `\n[Column types: ${types.join(", ")}]\n`;
}

/** Truncate a cell value for display in the prompt. */
function truncateValue(val: unknown): string {
  if (val === null || val === undefined) return "NULL";
  const s = String(val);
  return s.length > 80 ? s.slice(0, 77) + "..." : s;
}

/** Build a prompt string with the focused cell's SQL context and results. */
export function buildSQLPrompt(): string {
  const nb = get(activeNotebook);
  if (!nb) return "";
  const state = get(nb);
  const cell = state.cells.find((c) => c.id === state.focusedCellId);
  if (!cell?.sql?.trim()) return "";
  const limitNote =
    cell.limit !== undefined ? `\nLimit: ${cell.limit} rows` : "";

  let prompt = `[Current SQL query (${cell.connector} connector)${limitNote}]\n\`\`\`sql\n${cell.sql}\n\`\`\`\n`;

  // Append schema types and sample results if available
  if (cell.result?.schema?.fields?.length) {
    prompt += formatSchemaTypes(cell.result.schema);
  }
  if (cell.result?.data?.length) {
    prompt += formatResultsSummary(cell);
  }
  if (cell.error) {
    prompt += `\n[Query error]\n${cell.error}\n`;
  }

  return prompt + "\n";
}

export function createQueryChatConfig(): ChatConfig {
  return {
    agent: ToolName.ANALYST_AGENT,
    additionalContextStoreGetter: () => emptyContext,
    promptTransformer: (prompt) => {
      const context = buildSQLPrompt();
      return context ? context + prompt : prompt;
    },
    contextIndicator: () =>
      derived(activeNotebook, ($nb) => {
        if (!$nb) return null;
        const state = get($nb);
        const cell = state.cells.find((c) => c.id === state.focusedCellId);
        if (!cell?.sql?.trim()) return null;
        const hasResults = (cell.result?.data?.length ?? 0) > 0;
        return hasResults
          ? `SQL query + results attached (${cell.connector})`
          : `SQL query attached (${cell.connector})`;
      }),
    emptyChatLabel: "Ask for help with your SQL",
    placeholder: "Ask about your SQL query...",
    minChatHeight: "min-h-[4rem]",
  };
}
