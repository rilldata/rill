<script lang="ts">
  import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
  import {
    conditionalFormatSpecToMeasureFormatting,
    type PivotCanvasComponent,
    type PivotSpec,
    type TableSpec,
  } from "@rilldata/web-common/features/canvas/components/pivot";
  import {
    fieldConfigs,
    fieldNames,
    mergeFieldList,
    type ColumnSettings,
    type PivotFieldEntry,
    type PivotFieldListKey,
  } from "@rilldata/web-common/features/canvas/components/pivot/field-config";
  import type { PivotMeasureFormatting } from "@rilldata/web-common/features/dashboards/pivot/types";
  import type { AllKeys, FieldType } from "../types";
  import MultiFieldInput from "./MultiFieldInput.svelte";

  export let component: PivotCanvasComponent;
  export let canvasName: string;
  export let metricName: string;
  export let label: string;
  // The spec key holding the selected fields ("columns", "measures",
  // "row_dimensions" or "col_dimensions").
  export let id: string;
  // Entries are field names or objects carrying per-column overrides.
  export let selectedItems: PivotFieldEntry[] = [];
  export let types: FieldType[];
  export let ephemeralMeasures: EphemeralMeasureDef[] | undefined = undefined;

  $: spec = component.specStore;
  $: listKey = id as PivotFieldListKey;
  $: measureFormatting = conditionalFormatSpecToMeasureFormatting(
    $spec.conditional_format,
  );
  $: columnSettings = {
    listKey,
    configs: fieldConfigs($spec),
    adhocNames: new Set(
      ($spec.adhoc_measures ?? []).map((measure) => measure.name),
    ),
    onChange: (name, patch) => component.setFieldConfig(listKey, name, patch),
  } satisfies ColumnSettings;
</script>

<MultiFieldInput
  {canvasName}
  {metricName}
  {label}
  {id}
  selectedItems={fieldNames(selectedItems)}
  {types}
  {ephemeralMeasures}
  {component}
  onMultiSelect={(items) =>
    component.updateProperty(
      id as AllKeys<PivotSpec | TableSpec>,
      // Surviving entries keep their objects, so overrides follow reorders.
      mergeFieldList(selectedItems, items),
    )}
  {measureFormatting}
  setMeasureFormatting={(
    measureName: string,
    fmt: PivotMeasureFormatting | null,
  ) => component.setMeasureFormatting(measureName, fmt)}
  {columnSettings}
/>
