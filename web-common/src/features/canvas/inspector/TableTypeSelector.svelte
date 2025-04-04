<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { PivotCanvasComponent } from "../components/pivot";

  export let component: PivotCanvasComponent;
  export let metricsViewName: string;
  export let canvasName: string;

  $: ctx = getCanvasStore(canvasName);
  $: ({ getMetricsViewFromName } = ctx.canvasEntity.spec);

  $: ({ specStore } = component);

  $: selected = "columns" in $specStore ? 0 : 1;
  $: metricsView = getMetricsViewFromName(metricsViewName);
</script>

<div class="section">
  <InputLabel small label="Table type" id="table-type-selector" />
  <FieldSwitcher
    small
    expand
    fields={["Flat", "Pivot"]}
    {selected}
    onClick={(_, field) => {
      if (field === "Flat") {
        selected = 0;
        component.updateTableType("table", $metricsView);
      } else if (field === "Pivot") {
        selected = 1;
        component.updateTableType("pivot", $metricsView);
      }
    }}
  />
</div>

<style lang="postcss">
  .section {
    @apply px-5 flex flex-col gap-y-2 pt-2;
    @apply border-t border-gray-200;
  }
</style>
