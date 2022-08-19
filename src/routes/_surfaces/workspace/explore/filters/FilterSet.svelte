<script>
  import WithTogglableFloatingElement from "$lib/components/floating-element/WithTogglableFloatingElement.svelte";
  import Check from "$lib/components/icons/Check.svelte";
  import Close from "$lib/components/icons/Close.svelte";
  import Spacer from "$lib/components/icons/Spacer.svelte";
  import { Divider, Menu } from "$lib/components/menu";
  import MenuHeader from "$lib/components/menu/core/MenuHeader.svelte";
  import MenuItem from "$lib/components/menu/core/MenuItem.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";
  import { createEventDispatcher } from "svelte";
  import { fly, slide } from "svelte/transition";
  export let name;
  export let selectedValues;
  const dispatch = createEventDispatcher();

  $: visibleValues = selectedValues.slice(0, 3);
  let active = false;

  let currentlyVisibleValuesInMenu = selectedValues;
  /** update only if menu is closed */
  $: if (!active) currentlyVisibleValuesInMenu = selectedValues;
</script>

<WithTogglableFloatingElement
  let:toggleFloatingElement
  bind:active
  distance={8}
>
  <div>
    <Tooltip
      location="bottom"
      alignment="start"
      distance={8}
      activeDelay={60}
      suppress={active}
    >
      <button
        on:click={() => {
          toggleFloatingElement();
        }}
        transition:slide={{ duration: 200 }}
        class="
      grid gap-x-3 items-center px-2 py-1 rounded cursor-pointer
      {!active ? 'hover:bg-blue-50' : ''}
      {active ? 'bg-blue-100' : ''}
    "
        style:grid-template-columns="max-content max-content max-content"
      >
        <button on:click|stopPropagation={() => dispatch("remove-filters")}>
          <Close />
        </button>
        <div
          class="font-bold text-ellipsis overflow-hidden whitespace-nowrap"
          style:max-width="160px"
        >
          {name}
        </div>
        <div class="flex flex-wrap gap-x-3 gap-y-1">
          {#each visibleValues as value, i (i)}
            <div
              class="text-ellipsis overflow-hidden whitespace-nowrap"
              style:max-width={selectedValues.length === 1 ? "240px" : "120px"}
            >
              {value}{#if i < visibleValues.length - 1}, {/if}
            </div>
          {/each}
          {#if selectedValues.length > 3}
            {@const whatsLeft = selectedValues.length - 3}
            <div class="italic">
              + {whatsLeft} other{#if whatsLeft !== 1}s{/if}
            </div>
          {/if}
        </div>
      </button>
      <div
        slot="tooltip-content"
        transition:fly|local={{ duration: 100, y: 4 }}
      >
        <TooltipContent maxWidth="400px">
          <TooltipTitle>
            <svelte:fragment slot="name">{name}</svelte:fragment>
            <svelte:fragment slot="description">dimension</svelte:fragment>
          </TooltipTitle>
          click to edit the filters in this dimension
        </TooltipContent>
      </div>
    </Tooltip>
  </div>
  <Menu
    maxWidth="480px"
    slot="floating-element"
    on:escape={toggleFloatingElement}
    on:click-outside={toggleFloatingElement}
  >
    <MenuHeader>
      <svelte:fragment slot="title">Filters</svelte:fragment>
      <svelte:fragment slot="right">
        <button
          class="hover:bg-gray-100  grid place-items-center"
          style:width="24px"
          style:height="24px"
          on:click={toggleFloatingElement}
        >
          <Close size="16px" /></button
        >
      </svelte:fragment>
    </MenuHeader>
    <Divider />
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
</WithTogglableFloatingElement>
