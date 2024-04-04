<script lang="ts">
  import HideLeftSidebar from "@rilldata/web-common/components/icons/HideLeftSidebar.svelte";
  import SurfaceView from "@rilldata/web-common/components/icons/SurfaceView.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { portal } from "@rilldata/web-common/lib/actions/portal";

  export let navWidth: number;
  export let navOpen: boolean;
  export let resizing: boolean;
  export let show = true;

  let active = false;

  $: label = navOpen ? "Close sidebar" : "Show sidebar";
</script>

<button
  class="text-gray-500"
  class:resizing
  class:opacity-0={!show}
  class:shift={!navOpen}
  style:left="{navWidth - 32}px"
  aria-label={label}
  data-tooltip
  on:click
  on:mousedown={() => {
    active = false;
  }}
  use:portal
>
  {#if navOpen}
    <HideLeftSidebar size="18px" />
  {:else}
    <SurfaceView size="16px" mode={"hamburger"} />
  {/if}
</button>

<style lang="postcss">
  button {
    @apply rounded flex justify-center items-center absolute;
    @apply w-6 h-6 mt-[13px];
    transition-property: left;
  }

  button:hover {
    @apply bg-gray-300;
  }

  button:not(.resizing) {
    transition-duration: 400ms;
    transition-timing-function: ease-in-out;
  }

  .shift {
    left: 8px !important;
  }
</style>
