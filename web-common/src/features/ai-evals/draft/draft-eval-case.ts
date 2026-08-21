import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  runtimeServiceGetFile,
  runtimeServicePutFile,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { parseDocument } from "yaml";

// EvalCaseDraft is a case drafted from a chat exchange, to be reviewed and saved into an eval file.
export interface EvalCaseDraft {
  name: string;
  question: string;
  expectedAnswer?: string;
  notes?: string;
  metricsView?: string;
  measures?: string[];
  dimensions?: string[];
}

// draftEvalCaseFromExchange extracts an eval case draft from a conversation exchange:
// the user's prompt, the agent's final answer (as the starting point for the expected answer,
// which the human should edit to assert the truth), and the union of metrics views, measures,
// and dimensions the agent queried in that exchange.
export function draftEvalCaseFromExchange(
  messages: V1Message[],
  targetMessageId: string,
): EvalCaseDraft | null {
  const targetIdx = messages.findIndex((m) => m.id === targetMessageId);
  if (targetIdx === -1) return null;
  const target = messages[targetIdx];

  // Find the router call that started the exchange (the target's parent)
  const parentIdx = messages.findIndex((m) => m.id === target.parentId);
  let question = "";
  if (parentIdx !== -1) {
    try {
      const args = JSON.parse(messages[parentIdx].contentData ?? "{}");
      question = args?.prompt ?? "";
    } catch {
      // Fall through to an empty question the user can fill in
    }
  }

  let answer = "";
  try {
    const result = JSON.parse(target.contentData ?? "{}");
    answer = result?.response ?? extractTextContent(target);
  } catch {
    answer = extractTextContent(target);
  }

  // Union of query expectations from the exchange's query_metrics_view calls
  const start = parentIdx === -1 ? 0 : parentIdx;
  const metricsViews = new Set<string>();
  const measures = new Set<string>();
  const dimensions = new Set<string>();
  for (const msg of messages.slice(start, targetIdx + 1)) {
    if (msg.type !== "call" || msg.tool !== "query_metrics_view") continue;
    try {
      const args = JSON.parse(msg.contentData ?? "{}");
      if (typeof args?.metrics_view === "string") {
        metricsViews.add(args.metrics_view);
      }
      for (const m of args?.measures ?? []) {
        if (typeof m?.name === "string") measures.add(m.name);
      }
      for (const d of args?.dimensions ?? []) {
        if (typeof d?.name === "string") dimensions.add(d.name);
      }
    } catch {
      // Skip malformed tool calls
    }
  }

  return {
    name: slugifyEvalCaseName(question),
    question,
    expectedAnswer: answer,
    // Only include a metrics view expectation when the exchange queried exactly one
    metricsView: metricsViews.size === 1 ? [...metricsViews][0] : undefined,
    measures: measures.size > 0 ? [...measures] : undefined,
    dimensions: dimensions.size > 0 ? [...dimensions] : undefined,
  };
}

// listEvalFiles returns the file path of each AIEval resource in the project.
export function listEvalFiles(): { name: string; path: string }[] {
  return fileArtifacts
    .getNamesForKind(ResourceKind.AIEval)
    .map((name) => {
      const artifact = fileArtifacts.findFileArtifact(
        ResourceKind.AIEval,
        name,
      );
      return artifact ? { name, path: artifact.path } : null;
    })
    .filter((f): f is { name: string; path: string } => f !== null);
}

// appendEvalCase appends a case to an existing eval file, preserving comments and key order
// by editing the YAML AST rather than re-serializing.
export async function appendEvalCase(
  client: RuntimeClient,
  filePath: string,
  draft: EvalCaseDraft,
): Promise<void> {
  const file = await runtimeServiceGetFile(client, { path: filePath });
  const doc = parseDocument(file.blob ?? "");

  const existingNames: string[] = [];
  const cases = doc.get("cases");
  if (cases && typeof cases === "object" && "items" in cases) {
    for (const item of (cases as { items: unknown[] }).items) {
      const name = (item as { get?: (key: string) => unknown })?.get?.("name");
      if (typeof name === "string") existingNames.push(name);
    }
  }

  doc.addIn(["cases"], buildCaseNode(draft, existingNames));
  await runtimeServicePutFile(client, {
    path: filePath,
    blob: doc.toString(),
  });
}

// buildEvalFileContents renders a new eval file containing the draft case.
// The cases list must be built from a JS value (not parsed from a `cases: []` template):
// a parsed flow-style seq would force the whole case into flow style, where a multiline
// expected answer silently re-parses as a nested map and breaks the file.
export function buildEvalFileContents(draft: EvalCaseDraft): string {
  const doc = parseDocument("type: eval\n");
  doc.set("cases", [buildCaseNode(draft, [])]);
  return doc.toString();
}

// createEvalFileWithCase creates a new eval file containing the draft case and returns its path.
export async function createEvalFileWithCase(
  client: RuntimeClient,
  draft: EvalCaseDraft,
): Promise<string> {
  const allNames = fileArtifacts.getNamesForKind(ResourceKind.AIEval);
  const fileName = getName("eval", allNames);
  const path = `evals/${fileName}.yaml`;

  await runtimeServicePutFile(client, {
    path,
    blob: buildEvalFileContents(draft),
    create: true,
    createOnly: true,
  });
  return `/${path}`;
}

function buildCaseNode(draft: EvalCaseDraft, existingNames: string[]) {
  const expect: Record<string, unknown> = {};
  if (draft.expectedAnswer) expect.answer = draft.expectedAnswer;
  if (draft.metricsView) expect.metrics_view = draft.metricsView;
  if (draft.measures?.length) expect.measures = draft.measures;
  if (draft.dimensions?.length) expect.dimensions = draft.dimensions;

  const node: Record<string, unknown> = {
    name: getName(draft.name || "case", existingNames),
    question: draft.question,
  };
  if (draft.notes) node.notes = draft.notes;
  node.expect = expect;
  return node;
}

export function slugifyEvalCaseName(question: string): string {
  const slug = question
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "_")
    .replace(/^_+|_+$/g, "")
    .slice(0, 40)
    .replace(/_+$/g, "");
  return slug || "case";
}

function extractTextContent(message: V1Message): string {
  return (message.content ?? [])
    .map((block) => block.text ?? "")
    .filter(Boolean)
    .join("\n");
}
