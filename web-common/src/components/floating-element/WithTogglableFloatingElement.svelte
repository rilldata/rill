<script lang="ts">
  import { createEventDispatcher, setContext } from "svelte";
  import { writable } from "svelte/store";
  import { FloatingElement } from "./index";
  import type { FloatingElementRelationship } from "./types";

  export let location = "bottom";
  export let alignment = "middle";
  export let relationship: FloatingElementRelationship = "parent";
  export let distance = 0;
  export let pad = 8;
  export let suppress = false;
  export let active = false;
  export let inline = false;
  export let overflowFlipY = false;
  export let mousePos = { x: 0, y: 0 };

  /** this passes down the dom element used for the "outside click" action.
   * Since this element is not strictly within the parent of the menu (which is in a Portal),
   * we will need to check to see if this element was also clicked before firing the outside click callback.
   */
  const triggerElementStore = writable<Element | undefined>(undefined);
  $: triggerElementStore.set(parent?.children?.[0]);
  setContext("rill:menu:menuTrigger", triggerElementStore);

  const dispatch = createEventDispatcher();
  $: {
    if (active) dispatch("open");
    if (!active) dispatch("close");
  }

  let parent: HTMLDivElement | null = null;

  function handleClose() {
    active = false;
  }

  function toggleFloatingElement() {
    active = !active;
  }
</script>

<div class:inline bind:this={parent}>
  <slot {active} {handleClose} {toggleFloatingElement} />
  {#if parent && active && !suppress}
    <FloatingElement
      target={parent}
      {relationship}
      {location}
      {alignment}
      {distance}
      {pad}
      {overflowFlipY}
      {mousePos}
    >
      <slot
        name="floating-element"
        {active}
        {handleClose}
        {toggleFloatingElement}
      />
    </FloatingElement>
  {/if}
</div>
