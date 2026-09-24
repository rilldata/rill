<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Chip from "@rilldata/web-common/components/chip/core/Chip.svelte";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import MeasureExpressionInput from "@rilldata/web-common/features/dashboards/ephemeral-measures/MeasureExpressionInput.svelte";
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import type { ComponentSpec } from "@rilldata/web-common/features/canvas/components/types";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import {
    ephemeralDefsToSpecs,
    ephemeralSpecsToDefs,
    removeMeasureFromComponentSpec,
    type EphemeralMeasureSpec,
  } from "@rilldata/web-common/features/dashboards/ephemeral-measures/canvas";
  import {
    formatMeasureRef,
    parseMeasureExpression,
  } from "@rilldata/web-common/features/dashboards/ephemeral-measures/expression-parser";
  import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
  import {
    isReferenceableMeasure,
    slugifyEphemeralMeasureName,
    validateEphemeralMeasureCount,
    validateEphemeralMeasureDef,
  } from "@rilldata/web-common/features/dashboards/ephemeral-measures/validation";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { ephemeralFormatPresetOptions } from "@rilldata/web-common/features/dashboards/ephemeral-measures/format-presets";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
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

  const FORMAT_PRESETS = ephemeralFormatPresetOptions();

  let displayName = editingDef?.displayName ?? "";
  let expression = editingDef?.expression ?? "";
  let formatPreset: string = editingDef?.formatPreset ?? FormatPreset.HUMANIZE;
  let saveError: string | undefined = undefined;

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
  $: knownMeasureNames = new Set(
    referenceableMeasures.map((mes) => mes.name as string),
  );

  $: reservedNames = new Set([
    ...(metricsViewSpec?.measures ?? []).map((mes) => mes.name as string),
    ...(metricsViewSpec?.dimensions ?? []).map(
      (d) => (d.name || d.column) as string,
    ),
    ...(metricsViewSpec?.timeDimension ? [metricsViewSpec.timeDimension] : []),
    ...defs.filter((d) => d.name !== editingDef?.name).map((d) => d.name),
  ]);

  $: parsed = parseMeasureExpression(expression);
  $: unknownRef = parsed.refs.find((ref) => !knownMeasureNames.has(ref));
  $: expressionError =
    expression.trim() === ""
      ? undefined
      : (parsed.error?.message ??
        (unknownRef
          ? `"${unknownRef}" is not a measure in this metrics view`
          : undefined));

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

  function insertMeasure(name: string) {
    const token = formatMeasureRef(name);
    expression = expression === "" ? token : `${expression} ${token}`;
  }

  function save() {
    const def: EphemeralMeasureDef = {
      name:
        editingDef?.name ??
        slugifyEphemeralMeasureName(displayName, reservedNames),
      displayName: displayName.trim(),
      expression: expression.trim(),
      ...(formatPreset !== FormatPreset.HUMANIZE ? { formatPreset } : {}),
    };
    saveError =
      validateEphemeralMeasureDef(def, knownMeasureNames, reservedNames) ??
      (editingDef ? undefined : validateEphemeralMeasureCount(defs.length));
    if (saveError) return;

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

<Dialog.Root
  open
  onOpenChange={(open) => {
    if (!open) onClose();
  }}
>
  <Dialog.Content class="w-[480px] max-h-[90vh] flex flex-col">
    <Dialog.Header>
      <Dialog.Title>
        {editingDef
          ? m.dashboard_pivot_ephemeral_edit_title()
          : m.dashboard_pivot_ephemeral_new_title()}
      </Dialog.Title>
    </Dialog.Header>

    <!-- Body scrolls if the dialog still exceeds the viewport; the small gutter
         keeps the inputs' focus ring from being clipped by the overflow. -->
    <div class="flex flex-col gap-y-4 min-h-0 overflow-y-auto -mx-1 px-1">
      <Input
        bind:value={displayName}
        id="canvas-ephemeral-measure-name"
        label={m.dashboard_pivot_ephemeral_display_name_label()}
        placeholder={m.dashboard_pivot_ephemeral_display_name_placeholder()}
        claimFocusOnMount
      />

      <MeasureExpressionInput
        bind:value={expression}
        id="canvas-ephemeral-measure-expression"
        measures={referenceableMeasures}
        errors={expressionError}
      />

      {#if referenceableMeasures.length}
        <div class="flex flex-col gap-y-1.5">
          <span class="text-xs text-fg-secondary">
            {m.dashboard_pivot_ephemeral_insert_measure()}
          </span>
          <!-- Cap the chip list so dashboards with hundreds of measures scroll
               inside the dialog instead of pushing the footer off-screen. -->
          <div class="flex flex-wrap gap-1 max-h-40 overflow-y-auto">
            {#each referenceableMeasures as mes (mes.name)}
              <button
                type="button"
                on:click={() => insertMeasure(mes.name ?? "")}
              >
                <Chip type="measure" label={mes.displayName || mes.name}>
                  <span slot="body" class="text-xs">
                    {mes.displayName || mes.name}
                  </span>
                </Chip>
              </button>
            {/each}
          </div>
        </div>
      {/if}

      <label class="flex flex-col gap-y-1">
        <span class="text-sm font-medium text-fg-primary">
          {m.dashboard_pivot_ephemeral_format_label()}
        </span>
        <select
          bind:value={formatPreset}
          class="h-8 rounded-sm border bg-surface px-2 text-xs"
        >
          {#each FORMAT_PRESETS as preset (preset.value)}
            <option value={preset.value}>{preset.label}</option>
          {/each}
        </select>
      </label>

      {#if saveError}
        <p class="text-xs text-destructive">{saveError}</p>
      {/if}
    </div>

    <Dialog.Footer>
      {#if editingDef}
        <div class="mr-auto">
          <Button type="secondary-destructive" onClick={remove}>
            {m.common_delete()}
          </Button>
        </div>
      {/if}
      <Button type="secondary" onClick={onClose}>
        {m.common_cancel()}
      </Button>
      <Button
        type="primary"
        disabled={!displayName.trim() ||
          !expression.trim() ||
          !!expressionError}
        onClick={save}
      >
        {m.dashboard_pivot_ephemeral_save()}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
