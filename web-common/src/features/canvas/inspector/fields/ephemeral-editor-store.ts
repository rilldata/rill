import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
import { writable } from "svelte/store";

export type CanvasEphemeralMeasureEditorState = {
  // Component whose `adhoc_measures` the editor reads and writes.
  component: BaseCanvasComponent;
  canvasName: string;
  metricName: string;
  // null = creating a new definition.
  editingDef: EphemeralMeasureDef | null;
  // Invoked after a successful create so the opening field can select the
  // new measure.
  onCreated?: (name: string, displayName: string) => void;
};

// State of the canvas ephemeral measure editor, mirroring the explore's
// `ephemeralMeasureDialog`. Field inputs set it; the inspector root mounts the
// editor once and reacts to it. null = closed. Each open replaces the state
// wholesale, so a field never receives another field's `onCreated`.
export const canvasEphemeralMeasureEditor =
  writable<CanvasEphemeralMeasureEditorState | null>(null);
