<script lang="ts">
  import { setContext } from "svelte";
  import { writable } from "svelte/store";

  import FloatingElement from "../floating-element/FloatingElement.svelte";
  import Portal from "../Portal.svelte";
  export let location = "bottom";
  export let alignment = "middle";
  export let distance = 0;
  export let pad = 8;
  export let suppress = false;

  export let active = false;

  /** this passes down the dom element used for the "outside click" action.
   * Since this element is not strictly within the parent of the menu (which is in a Portal),
   * we will need to check to see if this element was also clicked before firing the outside click callback.
   */
  const triggerElementStore = writable(undefined);
  $: triggerElementStore.set(parent?.children?.[0]);
  setContext("rill:menu:menuTrigger", triggerElementStore);

  let parent;
</script>

<div class="contents" bind:this={parent}>
  <slot
    {active}
    handleClose={() => {
      active = false;
    }}
    toggleMenu={() => {
      active = !active;
    }}
  />
  {#if active && !suppress}
    <Portal>
      <div style="z-index: 50;">
        <FloatingElement
          target={parent}
          {location}
          {alignment}
          {distance}
          {pad}
        >
          <slot name="menu" />
        </FloatingElement>
      </div>
    </Portal>
  {/if}
</div>
