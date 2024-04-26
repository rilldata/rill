<script lang="ts">
  import { Search } from "@rilldata/web-common/components/search";
  import type { Table } from "@tanstack/svelte-table";
  import { getContext } from "svelte";
  import type { Readable } from "svelte/store";

  const table = getContext<Readable<Table<unknown>>>("table");

  // Search
  let filter = "";

  $: filterTable(filter);

  function filterTable(filter: string) {
    $table.setGlobalFilter(filter);
  }

  // Number of reports
  $: numReports = $table.getRowModel().rows.length;

  // Sort
  // function sortAlphabetically() {
  //   $table.setSorting([{ id: "monocolumn", desc: false }]);
  // }

  // function sortByMostRecentlyRun() {
  //   $table.setSorting([{ id: "lastRun", desc: true }]);
  // }

  // function sortByNextToRun() {
  //   $table.setSorting([{ id: "monocolumn", desc: false }]);
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

      <!-- filter menu button (future work) -->
      <!-- <Button on:click={() => console.log("open filter menu")} type="secondary">
    <span>Filter</span>
    <CaretDownIcon />
  </Button> -->

      <!-- Number of reports -->
      <span class="shrink-0"
        >{numReports} report{numReports !== 1 ? "s" : ""}</span
      >

      <!-- Sort button -->
      <!-- <WithTogglableFloatingElement active={openSortMenu}>
        <Button on:click={() => (openSortMenu = true)} type="secondary">
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
          <MenuItem on:select={sortAlphabetically}>Alphabetical</MenuItem>
          <MenuItem on:select={sortByMostRecentlyRun}
            >Most recently run</MenuItem
          >
          <MenuItem on:select={sortByNextToRun} disabled>Next to run</MenuItem>
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
