<script lang="ts">
  import {
    mapExprToMeasureFilter,
    MeasureFilterEntry,
  } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { Chip } from "../../../../components/chip";
  import {
    measureChipColors,
    type ChipColors,
  } from "../../../../components/chip/chip-types";
  import type { V1Expression } from "../../../../runtime-client";
  import MeasureFilterBody from "./MeasureFilterBody.svelte";

  export let dimensionName: string;
  export let label: string | undefined = undefined;
  export let colors: ChipColors = measureChipColors;
  export let expr: V1Expression | undefined = undefined;

  let filter: MeasureFilterEntry | undefined;
  $: filter = mapExprToMeasureFilter(expr);
</script>

<Chip {...colors} extraRounded={false} {label} outline readOnly>
  <div class="mx-2" slot="body">
    <MeasureFilterBody {dimensionName} {filter} {label} readOnly />
  </div>
</Chip>
