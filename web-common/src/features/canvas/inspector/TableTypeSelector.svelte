<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import type { PivotCanvasComponent } from "../components/pivot";

  export let component: PivotCanvasComponent;

  $: ({ specStore } = component);

  $: selected = "columns" in $specStore ? 0 : 1;
</script>

<div class="section">
  <InputLabel small label={m.canvas_table_type()} id="table-type-selector" />
  <FieldSwitcher
    small
    expand
    fields={[m.dashboard_flat(), m.dashboard_pivot()]}
    {selected}
    onClick={(index) => {
      if (index === 0) {
        selected = 0;
        component.updateTableType("table");
      } else if (index === 1) {
        selected = 1;
        component.updateTableType("pivot");
      }
    }}
  />
</div>

<style lang="postcss">
  .section {
    @apply px-5 flex flex-col gap-y-2 pt-2;
    @apply border-t;
  }
</style>
