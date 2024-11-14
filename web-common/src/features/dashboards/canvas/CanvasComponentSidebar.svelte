<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import ChartOptions from "@rilldata/web-common/features/dashboards/canvas/chart/ChartOptions.svelte";
  import ComponentOptions from "@rilldata/web-common/features/dashboards/canvas/ComponentOptions.svelte";
  import { ArrowLeft } from "lucide-svelte";
  import { slide } from "svelte/transition";

  let sidebarHeight = 0;
  let selectedComponent;

  $: heading = selectedComponent?.id
    ? selectedComponent?.title
    : "Add component";
</script>

<div
  class="sidebar"
  bind:clientHeight={sidebarHeight}
  transition:slide={{ axis: "x" }}
>
  <div class="heading">
    {#if selectedComponent}
      <Button
        type="subtle"
        class="inline-block"
        on:click={() => (selectedComponent = null)}
      >
        <ArrowLeft size="16px" />
      </Button>
    {/if}
    <h2>{heading}</h2>
  </div>

  {#if selectedComponent}
    <ChartOptions chartType={selectedComponent?.id} />
  {:else}
    <ComponentOptions bind:selectedComponent />
  {/if}
</div>

<style lang="postcss">
  .sidebar {
    @apply flex flex-col flex-none relative overflow-hidden;
    @apply h-full border-r z-0 w-60;
    transition-property: width;
    will-change: width;
    @apply select-none;
  }

  .heading {
    @apply flex gap-x-2;
    @apply p-2 text-lg font-semibold;
    @apply border-b border-slate-200;
  }
</style>
