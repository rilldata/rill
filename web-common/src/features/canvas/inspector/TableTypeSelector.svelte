<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type { CanvasComponentType } from "@rilldata/web-common/features/canvas/components/types";
  import type { CanvasComponentObj } from "@rilldata/web-common/features/canvas/components/util";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";

  export let component: CanvasComponentObj;
  export let componentType: CanvasComponentType;
  export let metricsViewName: string;

  const ctx = getCanvasStateManagers();
  const { getMetricsViewFromName } = ctx.canvasEntity.spec;

  $: selected = componentType === "table" ? 0 : 1;
  $: metricsView = getMetricsViewFromName(metricsViewName);

  async function onChange(tableType: "pivot" | "table") {
    if (tableType !== componentType) {
      component.updateTableType(tableType, $metricsView);
    }
  }
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
        onChange("table");
      } else if (field === "Pivot") {
        selected = 1;
        onChange("pivot");
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
