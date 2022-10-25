<script lang="ts">
  import { Button } from "@rilldata/web-local/lib/components/button";
  import { Search } from "@rilldata/web-local/lib/components/search";

  import Cancel from "@rilldata/web-local/lib/components/icons/Cancel.svelte";
  import Check from "@rilldata/web-local/lib/components/icons/Check.svelte";
  import Spacer from "@rilldata/web-local/lib/components/icons/Spacer.svelte";

  import { Menu } from "@rilldata/web-local/lib/components/menu";
  import MenuItem from "@rilldata/web-local/lib/components/menu/core/MenuItem.svelte";
  import { createEventDispatcher } from "svelte";
  import Footer from "./Footer.svelte";
  import Switch from "@rilldata/web-local/lib/components/button/Switch.svelte";

  export let selectedValues: string[];
  export let searchedValues: string[] = [];
  export let excludeMode = false;

  let excludeToggle = excludeMode;
  $: if (excludeToggle != excludeMode) {
    onToggleHandler();
  }

  let searchText = "";

  const dispatch = createEventDispatcher();

  function onSearch() {
    dispatch("search", searchText);
  }

  function onToggleHandler() {
    dispatch("toggle");
  }

  /** On instantiation, only take the exact current selectedValues, so that
   * when the user unchecks a menu item, it still persists in the FilterMenu
   * until the user closes.
   */
  let candidateValues = [...selectedValues];
  let valuesToDisplay = [...candidateValues];

  $: if (searchText) {
    valuesToDisplay = [...searchedValues];
  } else valuesToDisplay = [...candidateValues];

  $: numSelectedNotInSearch = selectedValues.filter(
    (v) => !valuesToDisplay.includes(v)
  ).length;

  function toggleValue(value) {
    dispatch("apply", value);

    if (!candidateValues.includes(value)) {
      candidateValues = [...candidateValues, value];
    }
  }
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

  <Search bind:value={searchText} on:input={onSearch} />

  <!-- apply a wrapped flex element to ensure proper bottom spacing between body and footer -->
  <div class="flex flex-col flex-1 overflow-auto w-full pb-1">
    {#if valuesToDisplay.length}
      {#each valuesToDisplay as value}
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
              <Check />
            {:else if selectedValues.includes(value) && excludeMode}
              <Cancel />
            {:else}
              <Spacer />
            {/if}
          </svelte:fragment>
          {#if value?.length > 240}
            {value.slice(0, 240)}...
          {:else}
            {value}
          {/if}
        </MenuItem>
      {/each}
    {:else}
      <div class="mt-5 ui-copy-disabled text-center">no results</div>
    {/if}
  </div>
  <Footer>
    <span class="flex gap-x-1 items-center font-semibold ui-copy">
      <Switch bind:checked={excludeToggle} />
      {#if excludeMode}
        <Cancel /> Exclude
      {:else}
        <Check /> Include
      {/if}
    </span>
    {#if numSelectedNotInSearch}
      <div class="ui-label italic">
        {numSelectedNotInSearch} other value{numSelectedNotInSearch > 1
          ? "s"
          : ""} selected
      </div>
    {/if}
  </Footer>
</Menu>
