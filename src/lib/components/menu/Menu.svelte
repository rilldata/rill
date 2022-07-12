<script context="module">
  // only one at a time
  const globalActiveMenu = writable(undefined);
</script>

<script lang="ts">
  import { guidGenerator } from "$lib/util/guid";
  import { createEventDispatcher, onMount, setContext } from "svelte";
  import { writable } from "svelte/store";
  import { fade } from "svelte/transition";

  export let dark = false;
  setContext("rill:menu:dark", dark);

  const dispatch = createEventDispatcher();

  const menuID = guidGenerator();

  let key;
  function handleKeydown(event) {
    key = event.key;

    if (key === "Escape") {
      dispatch("escape");
    }

    if (key === "ArrowDown") {
      $currentItem =
        $currentItem !== undefined
          ? Math.min($currentItem + 1, $menuItems.length - 1)
          : 0;
    }
    if (key === "ArrowUp") {
      $currentItem =
        $currentItem !== undefined ? Math.max($currentItem - 1, 0) : 0;
    }
  }

  function onSelect() {
    dispatch("item-select");
  }

  const menuItems = writable([]);
  const currentItem = writable(undefined);

  setContext("rill:menu:onSelect", onSelect);
  setContext("rill:menu:menuItems", menuItems);
  setContext("rill:menu:currentItem", currentItem);

  // once open, we should select the first menu item.
  onMount(() => {
    $currentItem = 0;
    $globalActiveMenu = menuID;
  });

  // This will effectively close any additional menus that might be open.
  $: if ($globalActiveMenu !== menuID) {
    dispatch("escape");
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div
  transition:fade|local={{ duration: 35 }}
  on:mouseleave={() => {
    $currentItem = undefined;
  }}
  class="
        py-2 
        w-max 
        rounded 
        flex 
        flex-col
        outline-none
        {dark
    ? 'bg-gray-800 border-none shadow'
    : 'bg-white border border-gray-300 shadow-md'}
        "
  style:outline="none"
  style:min-width="300px"
  tabindex="0"
  role="menu"
>
  <slot />
</div>
