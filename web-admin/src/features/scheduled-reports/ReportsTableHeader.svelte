<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Menu from "@rilldata/web-common/components/menu/core/Menu.svelte";
  import MenuItem from "@rilldata/web-common/components/menu/core/MenuItem.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import type { Table } from "@tanstack/table-core/src/types";
  import { getContext } from "svelte";
  import { get, Readable } from "svelte/store";

  const table = getContext("table") as Readable<Table<unknown>>;

  let filter = "";

  $: filterTable(table, filter);

  function filterTable(table: Readable<Table<unknown>>, filter: string) {
    get(table).setGlobalFilter(filter);
  }

  function sortAlphabetically() {
    get(table).setSorting([{ id: "monocolumn", desc: false }]);
  }

  function sortByMostRecentlyRun() {
    get(table).setSorting([{ id: "lastRun", desc: true }]);
  }

  function sortByNextToRun() {
    get(table).setSorting([{ id: "monocolumn", desc: false }]);
  }

  let openSortMenu = false;
  function closeSortMenu() {
    openSortMenu = false;
  }
</script>

<thead class="bg-slate-100">
  <tr>
    <div class="p-2 max-w-[800px] flex items-center gap-x-2">
      <div class="grow" />

      <!-- search bar -->
      <Search placeholder="Search" bind:value={filter} />

      <!-- filter menu button (future work) -->
      <!-- <Button on:click={() => console.log("open filter menu")} type="secondary">
    <span>Filter</span>
    <CaretDownIcon />
  </Button> -->

      <!-- sort menu button -->
      <WithTogglableFloatingElement active={openSortMenu}>
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
      </WithTogglableFloatingElement>
    </div>
  </tr>
</thead>
