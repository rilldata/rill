<script lang="ts">
  import {
    conditionalFormatSpecToMeasureFormatting,
    type PivotCanvasComponent,
    type PivotSpec,
    type TableSpec,
  } from "@rilldata/web-common/features/canvas/components/pivot";
  import type { PivotMeasureFormatting } from "@rilldata/web-common/features/dashboards/pivot/types";
  import type { AllKeys, FieldType } from "../types";
  import MultiFieldInput from "./MultiFieldInput.svelte";

  export let component: PivotCanvasComponent;
  export let canvasName: string;
  export let metricName: string;
  export let label: string;
  // The spec key holding the selected fields ("measures" or "columns").
  export let id: string;
  export let selectedItems: string[] = [];
  export let types: FieldType[];

  $: spec = component.specStore;
  $: measureFormatting = conditionalFormatSpecToMeasureFormatting(
    $spec.conditional_format,
  );
</script>

<MultiFieldInput
  {canvasName}
  {metricName}
  {label}
  {id}
  {selectedItems}
  {types}
  onMultiSelect={(items) =>
    component.updateProperty(id as AllKeys<PivotSpec | TableSpec>, items)}
  {measureFormatting}
  setMeasureFormatting={(
    measureName: string,
    fmt: PivotMeasureFormatting | null,
  ) => component.setMeasureFormatting(measureName, fmt)}
/>
