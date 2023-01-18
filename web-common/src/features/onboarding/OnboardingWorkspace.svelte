<script lang="ts">
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import Metrics from "@rilldata/web-common/components/icons/Metrics.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import Source from "@rilldata/web-common/components/icons/Source.svelte";
  import WorkspaceContainer from "@rilldata/web-local/lib/components/workspace/core/WorkspaceContainer.svelte";
  import WorkspaceBody from "@rilldata/web-local/lib/components/workspace/core/WorkspaceBody.svelte";
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
    {
      heading: "Define your metrics and dimensions",
      description:
        "Define aggregate metrics and break out dimensions for your modeled data.",
      icon: Metrics,
    },
    {
      heading: "Explore your metrics dashboard",
      description:
        "Interactively explore line charts and leaderboards to uncover insights.",
      icon: Explore,
    },
  ];
</script>

<WorkspaceContainer top="0px" assetID="onboarding" inspector={false}>
  <WorkspaceBody bgClass="bg-gray-100">
    <div class="mt-10 p-2 place-content-center h-screen">
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
              <p>{step.description}</p>
            </div>
          </div>
        {/each}
      </div>
    </div>
  </WorkspaceBody>
</WorkspaceContainer>
