import type { V1Resource } from "@rilldata/web-common/runtime-client";
import {
  isModelPartitioned,
  isModelIncremental,
  hasModelErroredPartitions,
} from "./utils";

export type ModelAction =
  | "describe"
  | "viewLogs"
  | "viewPartitions"
  | "refreshErrored"
  | "fullRefresh"
  | "incrementalRefresh";

/**
 * Returns the list of available actions for a model resource based on its state.
 * - Always: describe, viewLogs, fullRefresh
 * - If partitioned: viewPartitions
 * - If has errored partitions: refreshErrored
 * - If incremental: incrementalRefresh
 */
export function getAvailableModelActions(
  resource: V1Resource | undefined,
): ModelAction[] {
  if (!resource) return [];

  const actions: ModelAction[] = ["describe", "viewLogs"];

  if (isModelPartitioned(resource)) {
    actions.push("viewPartitions");
  }

  if (hasModelErroredPartitions(resource)) {
    actions.push("refreshErrored");
  }

  actions.push("fullRefresh");

  if (isModelIncremental(resource)) {
    actions.push("incrementalRefresh");
  }

  return actions;
}
