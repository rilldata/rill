import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { Scalar, YAMLSeq } from "yaml";

type YamlDocumentWithTags = {
  get(key: string): unknown;
  set(key: string, value: unknown): void;
  delete(key: string): unknown;
  createNode(value: unknown): unknown;
};

function readYamlString(node: unknown): string {
  if (typeof node === "string") return node.trim();

  if (node instanceof Scalar) {
    return typeof node.value === "string" ? node.value.trim() : "";
  }

  if (node && typeof node === "object" && "value" in node) {
    const value = (node as { value: unknown }).value;
    return typeof value === "string" ? value.trim() : "";
  }

  return "";
}

export function normalizeTags(tags: string[] | undefined): string[] {
  const seen = new Set<string>();
  const normalized: string[] = [];

  for (const tag of tags ?? []) {
    const value = tag.trim();
    if (!value || seen.has(value)) continue;
    seen.add(value);
    normalized.push(value);
  }

  return normalized;
}

export function readYamlTags(node: unknown): string[] {
  if (!(node instanceof YAMLSeq)) return [];

  return normalizeTags(node.items.map(readYamlString));
}

export function readRootYamlTags(document: { get(key: string): unknown }) {
  return readYamlTags(document.get("tags"));
}

export function setRootYamlTags(
  document: YamlDocumentWithTags,
  tags: string[],
) {
  const normalizedTags = normalizeTags(tags);

  if (normalizedTags.length) {
    document.set("tags", document.createNode(normalizedTags));
  } else {
    document.delete("tags");
  }
}

export function getResourceTagSuggestions(
  resources: V1Resource[] | undefined,
  ...extraTags: Array<string[] | undefined>
) {
  return buildTagSuggestions(
    resources?.flatMap((resource) => resource.meta?.tags ?? []),
    ...extraTags,
  );
}

export function buildTagSuggestions(
  ...sources: Array<string[] | undefined>
): string[] {
  return normalizeTags(sources.flatMap((source) => source ?? [])).sort((a, b) =>
    a.localeCompare(b),
  );
}
