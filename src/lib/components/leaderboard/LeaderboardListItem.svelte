<script lang="ts">
  import { slide } from "svelte/transition";
  import BarAndLabel from "$lib/components/BarAndLabel.svelte";
  export let value: number; // should be between 0 and 1.
  export let color = "bg-blue-200";
  export let isActive = false;

  let hovered = false;
  const onHover = () => {
    hovered = true;
  };
  const onLeave = () => {
    hovered = false;
  };
</script>

<button
  transition:slide|local={{ duration: 200 }}
  on:mouseover={onHover}
  on:mouseleave={onLeave}
  on:focus={onHover}
  on:blur={onLeave}
  on:click
  class="
        block 
        w-full 
        text-left 
        hover:bg-gray-100 
        transition-color"
>
  <BarAndLabel
    {color}
    {value}
    showBackground={false}
    tweenParameters={{ duration: 200 }}
    justify={false}
  >
    <div
      class="grid leaderboard-entry items-center gap-x-3"
      style:height="22px"
    >
      <div
        class="justify-self-start text-left w-full text-ellipsis overflow-hidden whitespace-nowrap"
      >
        <div class:font-bold={isActive}>
          <slot name="title" />
        </div>
      </div>
      <div class="justify-self-end" class:font-bold={isActive}>
        <slot name="right" />
      </div>
    </div>
  </BarAndLabel>
</button>

<style>
  .leaderboard-entry {
    grid-template-columns: auto max-content;
  }
</style>
