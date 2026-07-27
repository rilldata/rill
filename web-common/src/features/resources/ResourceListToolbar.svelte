<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { Search } from "@rilldata/web-common/components/search";
  import type { Table } from "tanstack-table-8-svelte-5";
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
  placeholder={m.resource_search_placeholder()}
  autofocus={false}
  bind:value={filter}
  rounded="lg"
/>
