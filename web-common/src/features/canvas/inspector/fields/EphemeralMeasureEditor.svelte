<!-- @component
Creates or edits an ephemeral measure in a canvas component's `adhoc_measures`.
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

  function writeRawSpecs(rawSpecs: EphemeralMeasureSpec[]) {
    component.updateProperty(
      EPHEMERAL_SPEC_KEY,
      rawSpecs.length ? rawSpecs : undefined,
    );
  }

  function save(def: EphemeralMeasureDef) {
    const rawSpecs = currentRawSpecs();
    const [spec] = ephemeralDefsToSpecs([def]);
    const index = rawSpecs.findIndex((entry) => entry?.name === def.name);
    if (index >= 0) {
      rawSpecs[index] = spec;
    } else {
      rawSpecs.push(spec);
    }
    writeRawSpecs(rawSpecs);
    if (!editingDef) onCreated?.(def.name, def.displayName);
    onClose();
  }

  function remove() {
    if (!editingDef) return;
    const name = editingDef.name;
    writeRawSpecs(currentRawSpecs().filter((entry) => entry?.name !== name));
    // Also remove it from the component's field selections, including chart
    // field configs and per-measure formatting, so the component stays valid.
    const changes = removeMeasureFromComponentSpec(
      $specStore as unknown as Record<string, unknown>,
      name,
    );
    const untyped = component as unknown as BaseCanvasComponent<
      Record<string, unknown>
    >;
    for (const [key, value] of Object.entries(changes)) {
      untyped.updateProperty(key, value);
    }
    onClose();
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
