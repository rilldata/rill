<!-- @component
The dialog for creating or editing an ephemeral measure definition, shared by
the explore dashboards (EphemeralMeasureDialog) and the canvas inspector
(canvas/inspector/fields/EphemeralMeasureEditor). It owns the form state and
validation; the wrapper supplies the measures the expression may reference and
persists the definition.
-->
<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Chip from "@rilldata/web-common/components/chip/core/Chip.svelte";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import type {
    MetricsViewSpecMeasure,
    V1MetricsViewSpec,
  } from "@rilldata/web-common/runtime-client";
  import { parseMeasureExpression } from "./expression-parser";
  import { ephemeralFormatPresetOptions } from "./format-presets";
  import MeasureExpressionInput from "./MeasureExpressionInput.svelte";
  import type { EphemeralMeasureDef } from "./types";
  import {
    slugifyEphemeralMeasureName,
    validateEphemeralMeasureCount,
    validateEphemeralMeasureDef,
  } from "./validation";

  // null = creating a new definition.
  export let editingDef: EphemeralMeasureDef | null = null;
  export let metricsView: V1MetricsViewSpec | undefined;
  // The measures the expression may reference (see isReferenceableMeasure).
  export let referenceableMeasures: MetricsViewSpecMeasure[];
  // The definitions that already exist alongside this one.
  export let existingDefs: EphemeralMeasureDef[];
  export let onSave: (def: EphemeralMeasureDef) => void;
  // Set when editing: shows the delete button.
  export let onRemove: (() => void) | undefined = undefined;
  export let onClose: () => void;

  const FORMAT_PRESETS = ephemeralFormatPresetOptions();

  let displayName = editingDef?.displayName ?? "";
  let expression = editingDef?.expression ?? "";
  let description = editingDef?.description ?? "";
  let formatPreset: string = editingDef?.formatPreset ?? FormatPreset.HUMANIZE;
  let saveError: string | undefined = undefined;
  let expressionInput: MeasureExpressionInput;

  $: otherDefs = existingDefs.filter((d) => d.name !== editingDef?.name);
  $: knownMeasureNames = new Set(
    referenceableMeasures.map((mes) => mes.name as string),
  );
  $: reservedNames = new Set([
    ...(metricsView?.measures ?? []).map((mes) => mes.name as string),
    ...(metricsView?.dimensions ?? []).map(
      (d) => (d.name || d.column) as string,
    ),
    ...(metricsView?.timeDimension ? [metricsView.timeDimension] : []),
    ...otherDefs.map((d) => d.name),
  ]);
  // Labels of every other measure: a display name that repeats one gets a
  // numeric suffix on save, so two measures never look the same.
  $: otherDisplayNames = [
    ...(metricsView?.measures ?? []).map((mes) => mes.displayName || mes.name),
    ...otherDefs.map((d) => d.displayName),
  ].filter((label): label is string => !!label);

  // Live expression feedback (only once the user typed something).
  $: parsed = parseMeasureExpression(expression);
  $: unknownRef = parsed.refs.find((ref) => !knownMeasureNames.has(ref));
  $: expressionError =
    expression.trim() === ""
      ? undefined
      : (parsed.error?.message ??
        (unknownRef
          ? `"${unknownRef}" is not a measure in this dashboard`
          : undefined));

  function save() {
    const def: EphemeralMeasureDef = {
      name:
        editingDef?.name ??
        slugifyEphemeralMeasureName(displayName, reservedNames),
      displayName: getName(displayName.trim(), otherDisplayNames),
      expression: expression.trim(),
      ...(formatPreset !== FormatPreset.HUMANIZE ? { formatPreset } : {}),
      ...(description.trim() ? { description: description.trim() } : {}),
    };
    saveError =
      validateEphemeralMeasureDef(def, knownMeasureNames, reservedNames) ??
      (editingDef
        ? undefined
        : validateEphemeralMeasureCount(existingDefs.length));
    if (saveError) return;
    onSave(def);
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
        id="ephemeral-measure-name"
        label={m.dashboard_pivot_ephemeral_display_name_label()}
        placeholder={m.dashboard_pivot_ephemeral_display_name_placeholder()}
        claimFocusOnMount
      />

      <MeasureExpressionInput
        bind:this={expressionInput}
        bind:value={expression}
        id="ephemeral-measure-expression"
        measures={referenceableMeasures}
        errors={expressionError}
      />

      <Input
        bind:value={description}
        id="ephemeral-measure-description"
        label={m.dashboard_pivot_ephemeral_description_label()}
        placeholder={m.dashboard_pivot_ephemeral_description_placeholder()}
        optional
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
                on:click={() => expressionInput.insertMeasure(mes)}
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
      {#if onRemove}
        <div class="mr-auto">
          <Button type="secondary-destructive" onClick={onRemove}>
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
