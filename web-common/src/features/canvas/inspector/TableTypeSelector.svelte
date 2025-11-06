<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type { PivotCanvasComponent } from "../components/pivot";

  export let component: PivotCanvasComponent;

  $: ({ specStore } = component);

  $: selected = "columns" in $specStore ? 0 : 1;
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
        component.updateTableType("table");
      } else if (field === "Pivot") {
        selected = 1;
        component.updateTableType("pivot");
      }
    }}
  />
</div>

<style lang="postcss">
  @reference "tailwindcss";

  @reference "tailwindcss";

  .section {
    @apply px-5 flex flex-col gap-y-2 pt-2;
    @apply border-t;
  }
</style>
