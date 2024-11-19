<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import ChartOptions from "@rilldata/web-common/features/dashboards/canvas/chart/ChartOptions.svelte";
  import ComponentOptions from "@rilldata/web-common/features/dashboards/canvas/ComponentOptions.svelte";
  import { Inspector } from "@rilldata/web-common/layout/workspace";
  import { ArrowLeft } from "lucide-svelte";

  let selectedComponent;

  $: heading = selectedComponent?.id
    ? selectedComponent?.title
    : "Add component";
</script>

<Inspector filePath="canvas_path" resizable={false} fixedWidth={320}>
  <div class="sidebar">
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
      <h1>{heading}</h1>
    </div>
    <div class="sidebar-body">
      {#if selectedComponent}
        <ChartOptions chartType={selectedComponent?.id} />
      {:else}
        <ComponentOptions bind:selectedComponent />
      {/if}
    </div>
    <footer
      class="flex flex-col gap-y-2 mt-auto border-t px-5 py-3 w-full text-sm text-gray-500"
    >
      <p>Checkout the docs for more information</p>
    </footer>
  </div>
</Inspector>

<style lang="postcss">
  .sidebar {
    @apply size-full w-full bg-background;
    @apply flex-none flex flex-col select-none rounded-[2px];
    transition-property: width;
    will-change: width;
  }

  .sidebar-body {
    @apply w-full h-full;
    @apply overflow-y-auto overflow-x-visible;
  }
  .heading {
    @apply flex gap-x-2;
    @apply p-2 text-lg font-semibold;
    @apply border-b border-slate-200;
  }
</style>
