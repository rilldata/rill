<script lang="ts">
  import type { Table } from "@tanstack/svelte-table";
  import { getContext } from "svelte";
  import type { Readable } from "svelte/store";

  const table = getContext<Readable<Table<unknown>>>("table");

  // Number of alerts
  $: numAlerts = $table.getRowModel().rows.length;
</script>

<thead>
  <tr>
    <td class="pl-[17px] pr-4 py-2 bg-slate-100">
      <!-- Number of runs -->
      Last {numAlerts} check{numAlerts !== 1 ? "s" : ""}
    </td>
  </tr>
</thead>

<!--
Rounded table corners are tricky:
- `border-radius` does not apply to table elements when `border-collapse` is `collapse`.
- You can only apply `border-radius` to <td>, not <tr> or <table>.
-->
<style lang="postcss">
  thead tr td {
    @apply border-y;
  }
  thead tr td:first-child {
    @apply border-l rounded-tl-sm;
  }
  thead tr td:last-child {
    @apply border-r rounded-tr-sm;
  }
</style>
