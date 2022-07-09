<script>
  import {
    createEventDispatcher,
    getContext,
    onDestroy,
    onMount,
  } from "svelte";
  import Spacer from "$lib/components/icons/Spacer.svelte";

  export let icon = false;

  const dispatch = createEventDispatcher();

  const onSelect = getContext("rill:menu:onSelect");
  const menuItems = getContext("rill:menu:menuItems");
  const currentItem = getContext("rill:menu:currentItem");

  let itemID;
  onMount(() => {
    // add to the menu's ids. This will enable us to use keybindings.
    itemID = $menuItems.length;
    $menuItems = [...$menuItems, itemID];
    if ($currentItem === undefined) {
      $currentItem = itemID;
    }
  });

  onDestroy(() => {
    $menuItems = [...$menuItems.filter((id) => id !== itemID)];
  });

  let element;

  $: active = itemID === $currentItem;

  // if the element is the active one,
  // let's move the focus on it.
  // An element can be the focus if
  // (1) the mouse moves over it,
  // (2) the user tabs to it,
  // (3) the user uses the keyboard arrows
  $: if (active && element) {
    element.focus();
  } else {
    if (element) {
      element.blur();
    }
  }

  let selected = false;

  let hovered = false;
</script>

<button
  bind:this={element}
  role="menuitem"
  style="--tw-ring-color: transparent"
  class="
        text-left 
        py-1
        px-3
        text-black
        focus:bg-gray-200
        focus:outline-none
        active:outline-none
        gap-x-4
        grid
        justify-items-stretch
    "
  style:grid-template-columns="auto max-content"
  class:selected
  on:mouseover={() => {
    $currentItem = itemID;
    hovered = true;
  }}
  on:mouseleave={() => {
    hovered = false;
  }}
  on:focus={() => {
    $currentItem = itemID;
  }}
  on:click={() => {
    selected = true;
    dispatch("select");
    setTimeout(() => {
      onSelect();
    }, 100);
  }}
>
  <div
    class="grid gap-x-2"
    style:grid-template-columns="{icon ? "max-content" : ""} auto"
  >
    {#if icon}
      <div class="self-center">
        <slot name="icon">
          <!-- <Spacer /> -->
        </slot>
      </div>
    {/if}
    <div class="text-left">
      <slot {hovered} />
    </div>
  </div>
  <div class="text-right text-gray-400">
    <slot name="right" {hovered} />
  </div>
</button>

<style>
  .selected {
    animation: flicker 75ms;
    animation-iteration-count: 1;
  }

  @keyframes flicker {
    0%,
    100% {
      background-color: rgb(75, 85, 99);
    }
    50% {
      background-color: rgba(255, 255, 255, 0);
    }
  }
</style>
