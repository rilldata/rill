<script lang="ts">
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { Chip } from "../../../../components/chip";

  import MeasureFilterBody from "./MeasureFilterBody.svelte";

  export let dimensionName: string;
  export let label: string | undefined = undefined;

  export let filter: MeasureFilterEntry | undefined = undefined;
</script>

<Chip type="measure" {label} readOnly>
  <MeasureFilterBody slot="body" {dimensionName} {filter} {label} />
</Chip>
