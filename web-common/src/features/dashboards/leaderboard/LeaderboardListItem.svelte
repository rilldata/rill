<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import { createEventDispatcher } from "svelte";
  import { fly, slide } from "svelte/transition";
  import BarAndLabel from "../../../components/BarAndLabel.svelte";

  export let value: number; // should be between 0 and 1.
  export let color = "bg-blue-200 dark:bg-blue-600";
  export let isActive = false;
  export let excluded = false;
  export let showIcon = true;
  export let showContext = false;

  /** compact mode is used in e.g. profiles */
  export let compact = false;

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
  /** used for overly-large bar values */
  let zigZag =
    "M" +
    Array.from({ length: 7 })
      .map((_, i) => {
        return `${15 - 4 * (i % 2)} ${1.7 * (i * 2)}`;
      })
      .join(" L");

  $: height = compact ? "18px" : "22px";
</script>

<button
  transition:slide|local={{ duration: 200 }}
  on:mouseover={onHover}
  on:mouseleave={onLeave}
  on:focus={onHover}
  on:blur={onLeave}
  on:click
  class="block flex flex-row w-full text-left transition-color"
>
  {#if showIcon}
    <div style:width="22px" style:height class="grid place-items-center">
      {#if isActive && !excluded}
        <Check size="20px" />
      {:else if isActive && excluded}
        <Cancel size="20px" />
      {:else}
        <Spacer />
      {/if}
    </div>
  {/if}
  <BarAndLabel
    {color}
    {value}
    showHover
    showBackground={false}
    tweenParameters={{ duration: 200 }}
    justify={false}
  >
    <div class="grid leaderboard-entry items-center gap-x-3" style:height>
      <div
        class="justify-self-start text-left w-full text-ellipsis overflow-hidden whitespace-nowrap"
      >
        <div>
          <slot name="title" {isActive} />
        </div>
      </div>
      <div
        class="justify-self-end overflow-hidden ui-copy-number flex gap-x-2 items-baseline"
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
<!-- if the value is greater than 100%, we should add this little serration -->
{#if value > 1.001}
  <div transition:fly={{ duration: 200, x: 20 }}>
    <svg
      style="
      position:absolute;
      right: 0px;
      transform: translateY(-{height});
    "
      width="15"
      height="22"
    >
      <path d={zigZag} fill="white" />
    </svg>
  </div>
{/if}

<style>
  .leaderboard-entry {
    grid-template-columns: auto max-content;
  }
</style>
