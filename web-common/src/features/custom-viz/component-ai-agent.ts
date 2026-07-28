import type { Conversation } from "@rilldata/web-common/features/chat/core/conversation";
import { getConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager";
import { ToolName } from "@rilldata/web-common/features/chat/core/types";
import { developerChatActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { get } from "svelte/store";

/**
 * Dev-agent integration for standalone component (custom viz) files.
 * Mirrors the per-widget chart agent, but targets a component file path instead
 * of a canvas item. Conversations are tracked per file so follow-up prompts
 * continue in context.
 */

const fileConversations = new Map<string, Conversation>();

/**
 * The conversion instruction sent to the dev agent when a Vega-Lite gallery
 * example is imported (and by the workspace's retry action). It turns the
 * verbatim example into a reusable parameterized component.
 */
export const CONVERT_EXAMPLE_PROMPT = [
  "This component was imported from the Vega-Lite example gallery and still contains the example's inline sample data. Convert it into a reusable parameterized component. The file's only top-level YAML keys are type, display_name, params, and custom_chart (which contains metrics_sql and vega_spec); never put metrics_sql/vega_spec at the top level and never nest custom_chart under a renderer: key. Then:",
  "replace the inline sample data (data.values) with a metrics_sql query against the most suitable metrics view in this project;",
  "the metrics_sql MUST be a single trivial statement of the form SELECT {{ .params.<dim> }}, {{ .params.<measure> }} FROM {{ .params.metrics_view }} with an ORDER BY and LIMIT — Metrics SQL is a restricted dialect with NO CTEs, JOINs, subqueries, window functions, CASE expressions, or aggregate functions (measures are already aggregated);",
  "make the sort customizable: declare a required string param named order_by and write ORDER BY {{ .params.order_by }} DESC (the editor offers the bound dimensions/measures as its choices, defaulting to the measure); time-series charts instead order by the time dimension ascending, without an order_by param;",
  "declare a number param named limit (not required, with a default) and end the SQL with LIMIT {{ .params.limit }} so dashboards can tune the row count — choose the default from the chart's visual density: around 10 where each row is a large mark (bars, arcs, text), up to 100 for many small marks (point/circle/square scatterplots, rect heatmaps, tick plots), and up to 2000 for time series;",
  "every dimension/measure/time_dimension param referenced by the vega_spec encodings or by the ORDER BY must also appear in the SELECT list — a field can only be used if it is selected;",
  "declare a required param named metrics_view of type metrics_view, and one required param per encoded field typed dimension, measure, or time_dimension (never string, and no defaults);",
  "name the params after their role in the chart (x_axis, y_axis, color, size, ...), not after the example's column names, and describe each param;",
  'template the query and all Vega field references with {{ .params.<name> }} and bind the data as {"name": "query1"};',
  "remove Vega transforms that reshape or aggregate the dataset (filter, aggregate, joinaggregate, window, stack, fold, pivot, lookup, ...) and do not reproduce their derived values in SQL — simplify the chart to direct field encodings, dropping encodings that depend on them;",
  'KEEP row-local calculate transforms whose derived values the chart\'s geometry needs (diverging sign flips, ratios, differences, midpoints, absolute values): rewrite their datum references against the params, keep the "as" name and the encodings reading it, and expose any literal category value the expression compares against as a required string param (e.g. negate_when for the category plotted negative) — scalar params are injected as native Vega-Lite params, referenceable directly in expressions;',
  "preserve timeUnit encodings (timeUnit is an encoding property, not a transform): when several channels derive from ONE temporal field via different timeUnits (calendar heatmaps, punch cards), declare a SINGLE time_dimension param referenced from every such channel with its original timeUnit and type, select it once in the SQL, and order by it ascending with a time-series limit (up to 2000);",
  'match every encoding\'s "type" to its bound param — dimension → "nominal"/"ordinal", time_dimension → the example\'s temporal/ordinal type with its timeUnit, measure → "quantitative" — and remove "aggregate" from encodings since measures arrive pre-aggregated, except where a timeUnit groups several rows into one mark (keep the example\'s aggregate there);',
  'never "bin" a dimension param or type it "quantitative" (metrics-view dimensions are categorical strings, so binning them renders a blank chart) — if the example bins raw numeric axis fields, make those channels nominal/ordinal dimension params without bin and keep count/size channels on measure params;',
  "keep the remaining visual style (mark, colors, axes, legends) as close to the example as possible, copying the example's mark and config blocks verbatim without inventing config or sizing properties the example does not have.",
].join(" ");

export function clearComponentFileConversation(filePath: string): void {
  fileConversations.delete(filePath);
}

/**
 * Ask the dev agent to edit a component file, opening the chat sidebar.
 */
export function sendComponentFilePrompt(
  client: RuntimeClient,
  filePath: string,
  userPrompt: string,
): void {
  const conversationManager = getConversationManager(client, {
    conversationState: "browserStorage",
    agent: ToolName.DEVELOPER_AGENT,
    surface: "developer",
  });

  const existing = fileConversations.get(filePath);
  const currentConversation = get(conversationManager.getCurrentConversation());
  if (!existing || existing !== currentConversation) {
    conversationManager.enterNewConversationMode();
  }

  const fullPrompt = `Update the component file ${filePath} (a standalone custom viz with declared params; type: component). ${userPrompt}`;
  developerChatActions.startChat(fullPrompt);

  const conversation = get(conversationManager.getCurrentConversation());
  fileConversations.set(filePath, conversation);
}
