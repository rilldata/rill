<script lang="ts">
  import ParquetIcon from "$lib/components/icons/Parquet.svelte";
  import Code from "$lib/components/icons/Code.svelte";
  import Database from "$lib/components/icons/Database.svelte";
  import SurfaceView from "$lib/components/icons/SurfaceView.svelte";

  const steps = [
    {
      heading: "import your data source",
      description:
        "Add sources by dragging and dropping your csv or parquet file to this window.",
      icon: SurfaceView,
    },
    {
      heading: "transform your source into one big table",
      description:
        "Leverage column profiling on sources to create one big table that has already been joined, normalized, and optimized for insights that scale.",
      icon: Database,
    },
    {
      heading: "define your metrics and dimensions",
      description:
        "Create scalable metrics definitions to represent the analytical logic used for aggregation and decomposition of transformed data.",
      icon: Code,
    },
    {
      heading: "explore your metrics dashboard",
      description:
        "Interactively explore line charts and leaderboards representing transformed sources, their metrics definitions, and dimensions.",
      icon: ParquetIcon,
    },
  ];
</script>

<div class="mt-10 p-2 place-content-center">
  <div class="text-center">
    <div class="font-bold">getting started</div>
    <p>building data intuition at every step of analysis</p>
  </div>
  <div class="p-5 pt-2">
    {#each steps as step (step.heading)}
      <div
        class="flex items-center p-6 mt-3 bg-gray-50 rounded-lg border border-gray-200"
      >
        <div>
          <svelte:component this={step.icon} color="grey" size="3em" />
        </div>
        <div class="ml-5">
          <h5 class="font-bold">{step.heading}</h5>
          <p class="italic">{step.description}</p>
        </div>
      </div>
    {/each}
  </div>
</div>
