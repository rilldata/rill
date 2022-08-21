<script>
  import Check from "$lib/components/icons/Check.svelte";
  import Close from "$lib/components/icons/Close.svelte";
  import Spacer from "$lib/components/icons/Spacer.svelte";

  import { Menu } from "$lib/components/menu";
  import Divider from "$lib/components/menu/core/Divider.svelte";
  import MenuHeader from "$lib/components/menu/core/MenuHeader.svelte";
  import MenuItem from "$lib/components/menu/core/MenuItem.svelte";
  import { createEventDispatcher } from "svelte";

  export let selectedValues;

  const dispatch = createEventDispatcher();

  /** On instantiation, only take the exact current selectedValues, so that
   * when the user unchecks a menu item, it still persists in the FilterMenu
   * until the user closes.
   */
  let currentlyVisibleValuesInMenu = [...selectedValues];
</script>

<Menu
  paddingTop={1}
  paddingBottom={1}
  rounded={false}
  maxWidth="480px"
  on:escape
  on:click-outside
>
  <MenuHeader>
    <svelte:fragment slot="title">Filters</svelte:fragment>
    <svelte:fragment slot="right">
      <button
        class="hover:bg-gray-100  grid place-items-center"
        style:width="24px"
        style:height="24px"
        on:click={() => {
          dispatch("close");
        }}
      >
        <Close size="16px" /></button
      >
    </svelte:fragment>
  </MenuHeader>
  <Divider marginTop={1} marginBottom={1} />
  {#each currentlyVisibleValuesInMenu as value}
    <MenuItem
      icon
      {value}
      on:select={() => {
        dispatch("select", value);
      }}
    >
      <svelte:fragment slot="icon">
        {#if selectedValues.includes(value)}
          <Check />
        {:else}
          <Spacer />
        {/if}
      </svelte:fragment>
      {#if value.length > 240}
        {value.slice(0, 240)}...
      {:else}
        {value}
      {/if}
    </MenuItem>
  {/each}
</Menu>
