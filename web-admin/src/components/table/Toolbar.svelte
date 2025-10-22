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
  $: numResults = $table.getRowModel().rows.length;

  beforeNavigate(() => (filter = "")); // resets filter when changing projects
</script>

<div class="w-full flex items-center justify-between gap-x-4">
  <!-- Search bar -->
  <Search
    placeholder="Search"
    autofocus={false}
    bind:value={filter}
    background={false}
    rounded="lg"
  />
  <!-- Result count -->
  <span class="text-sm text-gray-500 shrink-0">
    {numResults} result{numResults !== 1 ? "s" : ""}
  </span>
</div>
