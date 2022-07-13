<script lang="ts">
  import Source from "$lib/components/icons/Source.svelte";
  import Model from "$lib/components/icons/Model.svelte";
  // import Database from "$lib/components/icons/Database.svelte";
  // import SurfaceView from "$lib/components/icons/SurfaceView.svelte";

  const steps = [
    {
      heading: "Import your data source",
      description:
        "Add to your sources by clicking on the + icon, or by dragging a csv or parquet file to this window.",
      icon: Source,
    },
    {
      heading: "Model your sources into one big table",
      description:
        "Build intuition about your sources and use SQL to model them into an analytics-ready resource.",
      icon: Model,
    },
    // ACTIVATE WHEN WE RELEASE METRICS EXPLORE
    // {
    //   heading: "define your metrics and dimensions",
    //   description:
    //     "Define interesting aggregate metrics and break out dimensions for your modeled data.",
    //   icon: SurfaceView,
    // },
    // {
    //   heading: "explore your metrics dashboard",
    //   description:
    //     "Interactively explore line charts and leaderboards to uncover insights.",
    //   icon: Database,
    // },
  ];
</script>

<div class="mt-10 p-2 place-content-center">
  <div class="text-center">
    <div class="font-bold">Getting started</div>
    <p>Building data intuition at every step of analysis</p>
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
