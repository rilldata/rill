<script lang="ts">
  import { Button } from "@rilldata/web-local/lib/components/button";
  import { Search } from "@rilldata/web-local/lib/components/search";

  import Check from "@rilldata/web-local/lib/components/icons/Check.svelte";
  import Spacer from "@rilldata/web-local/lib/components/icons/Spacer.svelte";

  import { Menu } from "@rilldata/web-local/lib/components/menu";
  import MenuItem from "@rilldata/web-local/lib/components/menu/core/MenuItem.svelte";
  import { createEventDispatcher, tick } from "svelte";
  import Footer from "./Footer.svelte";

  export let selectedValues: string[];
  export let searchedValues: string[] = [];
  let searchText = "";

  const dispatch = createEventDispatcher();

  function onCloseHandler() {
    dispatch("close");
  }

  function onSearch() {
    dispatch("search", searchText);
  }

  async function onApplyHandler() {
    dispatch("apply", candidateValues);
    await tick();
    onCloseHandler();
  }

  /** On instantiation, only take the exact current selectedValues, so that
   * when the user unchecks a menu item, it still persists in the FilterMenu
   * until the user closes.
   */
  let candidateValues = [];
  let valuesToDisplay = [...selectedValues];

  $: if (searchText) {
    valuesToDisplay = [...searchedValues];
  } else valuesToDisplay = [...selectedValues];

  $: numSelectedNotInSearch = selectedValues.filter(
    (v) => !valuesToDisplay.includes(v)
  ).length;

  function toggleValue(value) {
    if (candidateValues.includes(value)) {
      candidateValues = [
        ...candidateValues.filter((candidate) => candidate !== value),
      ];
    } else {
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
            {#if selectedValues.includes(value) !== candidateValues.includes(value)}
              <Check />
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
      <div class="mt-5 italic text-gray-500 text-center">no results</div>
    {/if}
  </div>
  <Footer>
    <Button
      type="secondary"
      compact
      disabled={!candidateValues.length}
      on:click={onApplyHandler}
    >
      <Check />
      <span class="font-semibold text-gray-800">Include</span>
    </Button>
    {#if numSelectedNotInSearch}
      <div class="text-gray-600 italic">
        {numSelectedNotInSearch} other value{numSelectedNotInSearch > 1
          ? "s"
          : ""} selected
      </div>
    {/if}
  </Footer>
</Menu>
