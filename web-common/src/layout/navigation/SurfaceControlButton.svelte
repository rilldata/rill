<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import HideSidebar from "@rilldata/web-common/components/icons/HideSidebar.svelte";
  import SurfaceView from "@rilldata/web-common/components/icons/SurfaceView.svelte";

  export let navWidth: number;
  export let navOpen: boolean;
  export let resizing: boolean;
  export let show = true;
  export let onClick: () => void;

  $: label = navOpen ? "Close sidebar" : "Show sidebar";
</script>

<span
  class="text-gray-500"
  class:resizing
  class:opacity-0={!show}
  class:shift={!navOpen}
  style:left="{navWidth - 32}px"
  aria-label={label}
  title={label}
>
  <Button
    type={navOpen ? "secondary" : "ghost"}
    gray={!navOpen}
    selected={navOpen}
    square
    on:click={onClick}
  >
    {#if navOpen}
      <HideSidebar side="left" open={navOpen} size="18px" />
    {:else}
      <SurfaceView size="16px" mode={"hamburger"} />
    {/if}
  </Button>
</span>

<style lang="postcss">
  span {
    @apply rounded flex justify-center items-center absolute;
    @apply z-50;
    @apply w-6 h-6 mt-[10px];
    transition-property: left;
  }

  span:hover {
    @apply bg-gray-300;
  }

  span:not(.resizing) {
    transition-duration: 300ms;
    transition-timing-function: ease-in-out;
  }

  .shift {
    left: 12px !important;
  }
</style>
