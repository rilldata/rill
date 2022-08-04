<script lang="ts">
  import {
    createEventDispatcher,
    getContext,
    onDestroy,
    onMount,
  } from "svelte";
  import type { Writable } from "svelte/store";

  export let icon = false;
  export let disabled = false;

  const dispatch = createEventDispatcher();

  const dark = getContext("rill:menu:dark");
  const onSelect: () => void = getContext("rill:menu:onSelect");
  const menuItems: Writable<any> = getContext("rill:menu:menuItems");
  const currentItem: Writable<any> = getContext("rill:menu:currentItem");

  let itemID;
  onMount(() => {
    // add to the menu's ids. This will enable us to use keybindings.
    itemID = $menuItems.length;
    $menuItems = [...$menuItems, { id: itemID, disabled }];
    if ($currentItem === undefined) {
      $currentItem = itemID;
    }
  });

  onDestroy(() => {
    $menuItems = [...$menuItems.filter(({ id }) => id !== itemID)];
  });

  let element;

  $: active = itemID === $currentItem;

  // if the element is the active one,
  // let's move the focus on it.
  // An element can be the focus if
  // (1) the mouse moves over it,
  // (2) the user tabs to it,
  // (3) the user uses the keyboard arrows
  $: if (active && element && !disabled) {
    element.focus();
  } else {
    if (element) {
      element.blur();
    }
  }

  let selected = false;
  let hovered = false;

  function onFocus() {
    if (!disabled) {
      $currentItem = itemID;
      hovered = true;
    }
  }

  function onBlur() {
    if (!disabled) {
      $currentItem = undefined;
      hovered = false;
    }
  }

  $: textColor = dark
    ? `${disabled ? "text-gray-400" : "text-white focus:bg-gray-600"}`
    : `${disabled ? "text-gray-600" : "text-gray-900 focus:bg-gray-200"}`;
</script>

<button
  bind:this={element}
  role="menuitem"
  style="--tw-ring-color: transparent"
  class="
        text-left 
        py-1
        px-3
        focus:outline-none
        active:outline-none
        grid
        content-start
        items-start
        gap-x-3
        justify-items-stretch
        {textColor}
    "
  style:grid-template-columns="{icon ? "max-content" : ""} auto max-content"
  class:selected
  class:cursor-not-allowed={disabled}
  aria-disabled={disabled}
  on:mouseover={onFocus}
  on:mouseleave={onBlur}
  on:focus={onFocus}
  on:blur={() => {
    if (!disabled) {
      hovered = false;
    }
  }}
  on:click|stopPropagation={() => {
    if (!disabled) {
      selected = true;
      dispatch("select");
      setTimeout(() => {
        onSelect();
      }, 100);
    }
  }}
>
  {#if icon}
    <div
      style:height="18px"
      class="grid place-content-center"
      style:opacity=".8"
    >
      <slot name="icon" />
    </div>
  {/if}
  <div class="text-left">
    <div>
      <slot {hovered} />
    </div>
    <div class="text-gray-400 italic" style:font-size="11px">
      <slot name="description" />
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
