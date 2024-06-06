<script lang="ts">
  import { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { Chip } from "../../../../components/chip";
  import {
    measureChipColors,
    type ChipColors,
  } from "../../../../components/chip/chip-types";
  import MeasureFilterBody from "./MeasureFilterBody.svelte";

  export let dimensionName: string;
  export let label: string | undefined = undefined;
  export let colors: ChipColors = measureChipColors;
  export let filter: MeasureFilterEntry | undefined = undefined;
</script>

<Chip {...colors} extraRounded={false} {label} outline readOnly>
  <div class="mx-2" slot="body">
    <MeasureFilterBody {dimensionName} {filter} {label} readOnly />
  </div>
</Chip>
