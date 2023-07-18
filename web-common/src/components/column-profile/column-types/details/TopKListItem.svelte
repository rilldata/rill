<script lang="ts">
  import BarAndLabel from "@rilldata/web-common/components/BarAndLabel.svelte";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";

  export let value: number; // should be between 0 and 1.
  export let color = "bg-blue-200 dark:bg-blue-600";
  export let isActive = false;
  export let showContext = false;

  /** compact mode is used in e.g. profiles */

  const dispatch = createEventDispatcher();

  let hovered = false;
  const onHover = () => {
    hovered = true;
    dispatch("focus");
  };
  const onLeave = () => {
    hovered = false;
    dispatch("blur");
  };
</script>

<button
  class="block flex flex-row w-full text-left transition-color"
  on:blur={onLeave}
  on:click
  on:focus={onHover}
  on:mouseleave={onLeave}
  on:mouseover={onHover}
  transition:slide|local={{ duration: 200 }}
>
  <BarAndLabel
    {color}
    justify={false}
    showBackground={false}
    showHover
    tweenParameters={{ duration: 200 }}
    {value}
  >
    <div
      class="grid leaderboard-entry items-center gap-x-3"
      style:height="18px"
    >
      <div
        class="justify-self-start text-left w-full text-ellipsis overflow-hidden whitespace-nowrap"
      >
        <div>
          <slot {isActive} name="title" />
        </div>
      </div>
      <div
        class="justify-self-end overflow-hidden ui-copy-number flex gap-x-4 items-baseline"
      >
        <slot {isActive} name="right" />
        {#if $$slots["context"] && showContext}
          <div
            class="text-xs text-gray-500 dark:text-gray-400"
            style:width="44px"
          >
            <slot {isActive} name="context" />
          </div>
        {/if}
      </div>
    </div>
  </BarAndLabel>
</button>

<style>
  .leaderboard-entry {
    grid-template-columns: auto max-content;
  }
</style>
