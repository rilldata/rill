<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { Search } from "@rilldata/web-common/components/search";
  import type { Table } from "@tanstack/svelte-table";
  import { getContext } from "svelte";
  import type { Readable } from "svelte/store";

  const table = getContext("table") as Readable<Table<unknown>>;

  // Search
  let filter = "";

  function filterTable(filter: string) {
    $table.setGlobalFilter(filter);
  }

  $: filterTable(filter);

  beforeNavigate(() => (filter = "")); // resets filter when changing projects

  // Number of dashboards
  $: numDashboards = $table.getRowModel().rows.length;

  // Sort
  // function sortByTitle() {
  //   $table.setSorting([{ id: "title", desc: false }]);
  // }

  // function sortByName() {
  //   $table.setSorting([{ id: "name", desc: false }]);
  // }

  // function sortByLastRefreshTime() {
  //   $table.setSorting([{ id: "lastRefreshed", desc: true }]);
  // }

  // let openSortMenu = false;
  // function closeSortMenu() {
  //   openSortMenu = false;
  // }
</script>

<thead>
  <tr>
    <td
      class="pl-2 pr-4 py-2 max-w-[800px] flex items-center gap-x-2 bg-slate-100"
    >
      <!-- Search bar -->
      <div class="px-2 grow">
        <Search placeholder="Search" autofocus={false} bind:value={filter} />
      </div>

      <!-- Spacer -->
      <div class="grow" />

      <!-- Number of dashboards -->
      <span>{numDashboards} dashboard{numDashboards !== 1 ? "s" : ""}</span>

      <!-- Sort button -->
      <!-- <WithTogglableFloatingElement
        active={openSortMenu}
        distance={4}
        alignment="end"
      >
        <Button
          on:click={() => (openSortMenu = !openSortMenu)}
          type="secondary"
        >
          <span>Sort</span>
          <CaretDownIcon />
        </Button>
        <Menu
          slot="floating-element"
          minWidth="0px"
          on:item-select={closeSortMenu}
          on:click-outside={closeSortMenu}
          on:escape={closeSortMenu}
        >
          <MenuItem on:select={sortByTitle}>Alphabetical by title</MenuItem>
          <MenuItem on:select={sortByName}>Alphabetical by URL</MenuItem>
          <MenuItem on:select={sortByLastRefreshTime}
            >Most recently refreshed</MenuItem
          >
        </Menu>
      </WithTogglableFloatingElement> -->
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
