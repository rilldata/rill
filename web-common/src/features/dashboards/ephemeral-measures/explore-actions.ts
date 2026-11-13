import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { V1ExploreSpec } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { notifyEphemeralMeasureChange } from "./notify";
import type { EphemeralMeasureDef } from "./types";

/**
 * Adds or updates an ephemeral measure on an explore's state and confirms it.
 */
export function saveExploreEphemeralMeasure(
  exploreName: string,
  def: EphemeralMeasureDef,
  isNew: boolean,
) {
  if (isNew) {
    metricsExplorerStore.addEphemeralMeasure(exploreName, def);
  } else {
    metricsExplorerStore.updateEphemeralMeasure(exploreName, def);
  }
  notifyEphemeralMeasureChange(isNew ? "created" : "updated", def.displayName);
}

/**
 * Deletes an ephemeral measure, and every use of it, from an explore's state
 * and confirms it. Every explore entry point that deletes one goes through
 * here: the dialog, the pivot sidebar and the pivot header's chips.
 */
export function deleteExploreEphemeralMeasure(
  exploreName: string,
  measureName: string,
  explore: V1ExploreSpec | undefined,
) {
  const def = get(metricsExplorerStore).entities[
    exploreName
  ]?.ephemeralMeasures?.find((d) => d.name === measureName);
  if (!def) return;
  metricsExplorerStore.removeEphemeralMeasure(
    exploreName,
    measureName,
    explore,
  );
  notifyEphemeralMeasureChange("deleted", def.displayName);
}
