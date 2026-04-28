import {
  parseGraphUrlParams,
  tokenForSeedString,
  tokenForKind,
} from "../navigation/seed-parser";
import type { ResourceStatusFilterValue } from "./types";

export const ISOLATED_STORAGE_KEY = "rill:graph:showIsolated";

export const STATUS_FILTER_OPTIONS: {
  label: string;
  value: ResourceStatusFilterValue;
}[] = [
  { label: "OK", value: "ok" },
  { label: "Pending", value: "pending" },
  { label: "Warning", value: "warning" },
  { label: "Errored", value: "errored" },
];

/**
 * Read the isolated resources preference from localStorage.
 */
export function readIsolatedPreference(): boolean {
  try {
    return localStorage.getItem(ISOLATED_STORAGE_KEY) === "true";
  } catch {
    return false;
  }
}

/**
 * Persist the isolated resources preference to localStorage.
 */
export function writeIsolatedPreference(value: boolean): void {
  try {
    localStorage.setItem(ISOLATED_STORAGE_KEY, String(value));
  } catch {
    // ignore
  }
}

/**
 * Derive graph seeds and active kind from URL parameters.
 */
export function deriveGraphState(url: URL) {
  const urlParams = parseGraphUrlParams(url);
  const derivedKindFromResource =
    urlParams.resources.length > 0
      ? tokenForSeedString(urlParams.resources[0])
      : null;
  const activeKind = urlParams.kind ?? derivedKindFromResource ?? "dashboards";
  const seeds = urlParams.kind
    ? [urlParams.kind]
    : urlParams.resources.length > 0
      ? urlParams.resources
      : [activeKind];
  const hasResourceParam = urlParams.resources.length > 0;
  const selectedGroupId = hasResourceParam ? urlParams.resources[0] : null;

  return {
    urlParams,
    activeKind,
    seeds,
    hasResourceParam,
    selectedGroupId,
  };
}

/**
 * Build the URL search params for a group selection change.
 */
export function buildGroupChangeParams(
  groupId: string,
  activeKind: string,
): URLSearchParams {
  const name = groupId.includes(":") ? groupId.split(":").pop() : groupId;
  const kindPart = groupId.includes(":")
    ? groupId.split(":").slice(0, -1).join(":")
    : null;
  const derivedKind = kindPart ? tokenForKind(kindPart) : null;
  const params = new URLSearchParams();
  params.set("kind", derivedKind ?? activeKind);
  if (name) params.set("resource", name);
  return params;
}
