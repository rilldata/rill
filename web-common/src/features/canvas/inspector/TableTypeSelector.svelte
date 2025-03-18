<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type { CanvasComponentType } from "@rilldata/web-common/features/canvas/components/types";
  import type { CanvasComponentObj } from "@rilldata/web-common/features/canvas/components/util";

  export let component: CanvasComponentObj;
  export let componentType: CanvasComponentType;

  $: selected = componentType === "table" ? 0 : 1;

  function onChange(tableType: "pivot" | "table") {
    if (tableType !== componentType) {
      component.updateTableType(tableType);
    }
  }
</script>

<div class="section">
  <InputLabel small label="Table type" id="table-type-selector" />
  <FieldSwitcher
    small
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
