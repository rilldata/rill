<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import Cancel from "../../icons/Cancel.svelte";
  import Check from "../../icons/Check.svelte";
  import Spacer from "../../icons/Spacer.svelte";
  import { Menu, MenuItem } from "../../menu";
  import { Search } from "../../search";
  import Footer from "./Footer.svelte";
  import Button from "../../button/Button.svelte";

  export let excludeMode: boolean;
  export let selectedValues: string[];
  export let allValues: string[] | null = [];

  let searchText = "";

  const dispatch = createEventDispatcher();

  function onSearch() {
    dispatch("search", searchText);
  }

  function toggleValue(value: string) {
    dispatch("apply", value);
  }

  function toggleSelectAll() {
    allValues?.forEach((value) => {
      if (!allSelected && selectedValues.includes(value)) return;

      toggleValue(value);
    });
  }

  $: allSelected =
    selectedValues?.length && allValues?.length === selectedValues.length;
</script>

<Menu
  paddingTop={1}
  paddingBottom={0}
  rounded={false}
  focusOnMount={false}
  maxWidth="480px"
  minHeight="150px"
  maxHeight="400px"
  on:escape
  on:click-outside
>
  <!-- the min-height is set to have about 3 entries in it -->
  <div class="px-3 py-2">
    <Search
      bind:value={searchText}
      on:input={onSearch}
      label="Search list"
      showBorderOnFocus={false}
    />
  </div>

  <!-- apply a wrapped flex element to ensure proper bottom spacing between body and footer -->
  <div class="flex flex-col flex-1 overflow-auto w-full pb-1">
    {#if allValues?.length}
      {#each allValues.sort() as value}
        <MenuItem
          icon
          animateSelect={false}
          focusOnMount={false}
          on:select={() => {
            toggleValue(value);
          }}
        >
          <svelte:fragment slot="icon">
            {#if selectedValues.includes(value) && !excludeMode}
              <Check size="20px" color="#15141A" />
            {:else if selectedValues.includes(value) && excludeMode}
              <Cancel size="20px" color="#15141A" />
            {:else}
              <Spacer size="20px" />
            {/if}
          </svelte:fragment>
          <span
            class:ui-copy-disabled={selectedValues.includes(value) &&
              excludeMode}
          >
            {#if value?.length > 240}
              {value.slice(0, 240)}...
            {:else}
              {value}
            {/if}
          </span>
        </MenuItem>
      {/each}
    {:else}
      <div class="mt-5 ui-copy-disabled text-center">no results</div>
    {/if}
  </div>
  <Footer>
    <Button type="text" on:click={toggleSelectAll}>
      {#if allSelected}
        Deselect all
      {:else}
        Select all
      {/if}
    </Button>

    <Button type="secondary" on:click={() => dispatch("toggle")}>
      {#if excludeMode}
        Include
      {:else}
        Exclude
      {/if}
    </Button>
  </Footer>
</Menu>
