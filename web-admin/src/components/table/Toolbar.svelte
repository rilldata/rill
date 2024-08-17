<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { Search } from "@rilldata/web-common/components/search";
  import { getContext } from "svelte";

  const table = getContext("table");

  // Search
  let filter = "";

  function filterTable(filter: string) {
    $table.setGlobalFilter(filter);
  }

  $: filterTable(filter);

  beforeNavigate(() => (filter = "")); // resets filter when changing projects
</script>

<div class="w-full max-w-[800px] flex items-center">
  <!-- Search bar -->
  <Search
    placeholder="Search"
    autofocus={false}
    bind:value={filter}
    background={false}
  />

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
</div>
