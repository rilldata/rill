import { writable } from "svelte/store";
import type { EphemeralMeasureDef } from "./types";

// State of the ephemeral measure dialog. null = closed;
// `def` present = editing an existing definition, otherwise creating a new one.
export const ephemeralMeasureDialog = writable<{
  def?: EphemeralMeasureDef;
} | null>(null);
