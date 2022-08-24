<script>
  import { Button } from "$lib/components/button";

  import Check from "$lib/components/icons/Check.svelte";
  import Close from "$lib/components/icons/Close.svelte";
  import Spacer from "$lib/components/icons/Spacer.svelte";

  import { Menu } from "$lib/components/menu";
  import Divider from "$lib/components/menu/core/Divider.svelte";
  import MenuHeader from "$lib/components/menu/core/MenuHeader.svelte";
  import MenuItem from "$lib/components/menu/core/MenuItem.svelte";
  import { createEventDispatcher, tick } from "svelte";
  import Footer from "./Footer.svelte";

  export let selectedValues;

  const dispatch = createEventDispatcher();

  function onCloseHandler() {
    dispatch("close");
  }

  async function onApplyHandler() {
    dispatch(
      "apply",
      /** get the original filters that are not left in the candidates */
      currentlySelectedValues.filter(
        (value) => !candidateValues.includes(value)
      )
    );
    await tick();
    onCloseHandler();
  }

  /** On instantiation, only take the exact current selectedValues, so that
   * when the user unchecks a menu item, it still persists in the FilterMenu
   * until the user closes.
   */
  let currentlySelectedValues = [...selectedValues];
  let candidateValues = [...selectedValues];

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
  maxWidth="480px"
  on:escape
  on:click-outside
>
  <!-- the min-height is set to have about 3 entries in it -->

  <MenuHeader>
    <svelte:fragment slot="title">Filters</svelte:fragment>
    <svelte:fragment slot="right">
      <button
        class="hover:bg-gray-100  grid place-items-center"
        style:width="24px"
        style:height="24px"
        on:click={onCloseHandler}
      >
        <Close size="16px" /></button
      >
    </svelte:fragment>
  </MenuHeader>

  <Divider marginTop={1} marginBottom={1} />
  <!-- apply a wrapped flex element to ensure proper bottom spacing between body and footer -->
  <div class="flex flex-col w-full pb-1">
    {#each currentlySelectedValues as value}
      <MenuItem
        icon
        {value}
        on:select={() => {
          toggleValue(value);
        }}
      >
        <svelte:fragment slot="icon">
          {#if candidateValues.includes(value)}
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
  </div>
  <Footer>
    <Button type="text" compact on:click={onCloseHandler}>Cancel</Button>
    <Button
      type="primary"
      compact
      disabled={currentlySelectedValues.every((value) =>
        candidateValues.includes(value)
      )}
      on:click={onApplyHandler}>Apply</Button
    >
  </Footer>
</Menu>
