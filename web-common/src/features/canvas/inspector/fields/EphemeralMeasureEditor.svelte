<!-- @component
Creates or edits an ephemeral measure in a canvas component's `adhoc_measures`.
It closes and confirms only once the canvas file is saved; a failed save keeps
the dialog open with an error.
-->
<script lang="ts">
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import type { ComponentSpec } from "@rilldata/web-common/features/canvas/components/types";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import {
    ephemeralDefsToSpecs,
    ephemeralSpecsToDefs,
    removeMeasureFromComponentSpec,
    type EphemeralMeasureSpec,
  } from "@rilldata/web-common/features/dashboards/ephemeral-measures/canvas";
  import EphemeralMeasureForm from "@rilldata/web-common/features/dashboards/ephemeral-measures/EphemeralMeasureForm.svelte";
  import { notifyEphemeralMeasureChange } from "@rilldata/web-common/features/dashboards/ephemeral-measures/notify";
  import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
  import { isReferenceableMeasure } from "@rilldata/web-common/features/dashboards/ephemeral-measures/validation";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { AllKeys } from "../types";

  export let component: BaseCanvasComponent;
  export let canvasName: string;
  export let metricName: string;
  // null = creating a new definition.
  export let editingDef: EphemeralMeasureDef | null = null;
  export let onClose: () => void;
  // Invoked with the new definition's name and display name after a successful
  // create, so the opener can add it to its field selection.
  export let onCreated:
    | ((name: string, displayName: string) => void)
    | undefined = undefined;

  const client = useRuntimeClient();

  // The spec key holding the definitions on every component that supports them.
  const EPHEMERAL_SPEC_KEY = "adhoc_measures" as AllKeys<ComponentSpec>;

  $: ctx = getCanvasStore(canvasName, client.instanceId);
  $: metricsViewStore =
    ctx.canvasEntity.metricsView.getMetricsViewFromName(metricName);
  $: metricsViewSpec = $metricsViewStore.metricsView;

  $: specStore = component.specStore;
  $: defs =
    ephemeralSpecsToDefs(
      ($specStore as { adhoc_measures?: EphemeralMeasureSpec[] })
        .adhoc_measures,
    ) ?? [];

  $: referenceableMeasures = (metricsViewSpec?.measures ?? []).filter(
    (mes) => mes.name && isReferenceableMeasure(mes),
  );

  // The raw YAML list, re-read at write time so a save never clobbers entries
  // added or reordered elsewhere (code editor, another refetch) while the
  // editor was open. Unknown or partial entries are preserved verbatim.
  function currentRawSpecs(): EphemeralMeasureSpec[] {
    const raw = ($specStore as { adhoc_measures?: EphemeralMeasureSpec[] })
      .adhoc_measures;
    return Array.isArray(raw) ? [...raw] : [];
  }

  // Resolves to whether the definition was saved.
  async function save(def: EphemeralMeasureDef): Promise<boolean> {
    const isNew = !editingDef;
    const rawSpecs = currentRawSpecs();
    const [spec] = ephemeralDefsToSpecs([def]);
    const index = rawSpecs.findIndex((entry) => entry?.name === def.name);
    if (index >= 0) {
      rawSpecs[index] = spec;
    } else {
      rawSpecs.push(spec);
    }
    const saved = await component.updateProperty(
      EPHEMERAL_SPEC_KEY,
      rawSpecs.length ? rawSpecs : undefined,
    );
    if (!saved) return false;
    // Selecting the new measure is a second file write; starting it only once
    // the definition is saved keeps the two writes from racing.
    if (isNew) onCreated?.(def.name, def.displayName);
    notifyEphemeralMeasureChange(
      isNew ? "created" : "updated",
      def.displayName,
    );
    onClose();
    return true;
  }

  // Resolves to whether the deletion was saved.
  async function remove(): Promise<boolean> {
    if (!editingDef) return false;
    const { name, displayName } = editingDef;
    const remaining = currentRawSpecs().filter((entry) => entry?.name !== name);
    // One write removes the definition along with its uses in the component's
    // field selections, chart field configs and per-measure formatting, so the
    // component never references a deleted measure.
    const saved = await (
      component as unknown as BaseCanvasComponent<Record<string, unknown>>
    ).updateProperties({
      ...removeMeasureFromComponentSpec(
        $specStore as unknown as Record<string, unknown>,
        name,
      ),
      [EPHEMERAL_SPEC_KEY]: remaining.length ? remaining : undefined,
    });
    if (!saved) return false;
    notifyEphemeralMeasureChange("deleted", displayName);
    onClose();
    return true;
  }
</script>

<EphemeralMeasureForm
  {editingDef}
  metricsView={metricsViewSpec}
  {referenceableMeasures}
  existingDefs={defs}
  onSave={save}
  onRemove={editingDef ? remove : undefined}
  {onClose}
/>
