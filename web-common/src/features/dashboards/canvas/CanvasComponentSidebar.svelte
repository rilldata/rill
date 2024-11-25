<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import {
    chartConfig,
    updateChartType,
  } from "@rilldata/web-common/features/dashboards/canvas/chart/configStore";
  import { Inspector } from "@rilldata/web-common/layout/workspace";
  import { ArrowLeft } from "lucide-svelte";
  import ComponentOptions from "./ComponentOptions.svelte";

  $: heading = "Add component";
  $: selectedComponent = $chartConfig.chartType;
</script>

<Inspector filePath="canvas_path" resizable={false} fixedWidth={320}>
  <div class="sidebar">
    <div class="heading">
      {#if selectedComponent}
        <Button
          type="subtle"
          class="inline-block"
          on:click={() => {
            updateChartType(null);
            heading = "Add component";
          }}
        >
          <ArrowLeft size="16px" />
        </Button>
      {/if}
      <h1>{heading}</h1>
    </div>
    <div class="sidebar-body">
      <ComponentOptions
        on:select={(e) => {
          heading = e.detail?.title;
          updateChartType(e.detail?.id);
        }}
      />
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
