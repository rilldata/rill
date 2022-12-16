<script lang="ts">
  import {
    createEventDispatcher,
    getContext,
    onDestroy,
    onMount,
  } from "svelte";
  import type { Writable } from "svelte/store";

  export let icon = false;
  export let role = "menuitem";
  export let selected = false;
  export let animateSelect = true;
  export let focusOnMount = true;
  export let disabled = false;
  export let propogateSelect = true; // if true, clicks will fire the `rill:menu:onSelect` function

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
    if (focusOnMount && $currentItem === undefined) {
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

  let justClicked = false;
  export let focused = false;

  /** accessibility requirements */
  let ariaProperties;
  $: if (role === "menuitem") {
    ariaProperties = {
      role,
    };
  } else if (role === "option") {
    ariaProperties = {
      role,
      ["aria-selected"]: selected,
    };
  }
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

  function handleClick() {
    if (disabled) return;
    // set justClicked to true if this item isn't already selected.
    // this happens when these menu items are multi-selectable
    if (!selected) {
      // only animate if animateSelect is true, which it is by default.
      if (animateSelect) justClicked = true;
      // fire an event to change anything before any selection animation occurs.
      dispatch("before-select");
    }
    // pre-select
    setTimeout(
      () => {
        dispatch("select");
        if (propogateSelect) onSelect();
        justClicked = false;
      },
      animateSelect ? 150 : 0
    );
  }

  $: textColor = dark
    ? `${
        disabled ? "text-gray-500" : "focus:bg-gray-600 dark:focus:bg-gray-600"
      }`
    : `${
        disabled ? "text-gray-500" : "focus:bg-gray-200 dark:focus:bg-gray-600"
      }`;
</script>

<button
  bind:this={element}
  {...ariaProperties}
  class:dark
  class:surface-focus={hovered}
  style="--tw-ring-color: transparent; --flicker-color:{dark
    ? 'rgb(75, 85, 99)'
    : 'rgb(235, 235, 235)'}"
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
        ui-copy
        {textColor}
    "
  style:grid-template-columns="{icon ? "max-content" : ""} auto max-content"
  class:recently-clicked={justClicked}
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
  on:click|stopPropagation={handleClick}
>
  {#if icon}
    <div
      style:height="18px"
      class="grid place-content-center ui-copy-icon dark:text-white"
      style:opacity=".8"
    >
      <slot name="icon" />
    </div>
  {/if}
  <div
    class:ui-copy={!disabled}
    class:ui-copy-disabled={disabled}
    class="text-left"
  >
    <div>
      <slot {focused} />
    </div>
    <div
      class:ui-copy-muted={!disabled}
      class:ui-copy-disabled-faint={disabled}
      style:font-size="11px"
    >
      <slot name="description" />
    </div>
  </div>
  <div class="text-right ui-copy-muted">
    <slot name="right" {focused} />
  </div>
</button>

<style>
  .recently-clicked {
    animation: flicker 150ms;
    animation-iteration-count: 1;
  }

  @keyframes flicker {
    0%,
    100% {
      background-color: rgba(255, 255, 255, 0);
    }
    50% {
      background-color: var(--flicker-color);
    }
  }
</style>
