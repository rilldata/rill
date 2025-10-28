<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { Search } from "@rilldata/web-common/components/search";
  import type { Table } from "@tanstack/svelte-table";
  import { getContext } from "svelte";
  import type { Readable } from "svelte/store";

  const table = getContext<Readable<Table<unknown>>>("table");

  // Search
  let filter = "";

  function filterTable(filter: string) {
    $table.setGlobalFilter(filter);
  }

  $: filterTable(filter);

  beforeNavigate(() => (filter = "")); // resets filter when changing projects
</script>

<Search
  placeholder="Search"
  autofocus={false}
  bind:value={filter}
  background={false}
  rounded="lg"
/>
