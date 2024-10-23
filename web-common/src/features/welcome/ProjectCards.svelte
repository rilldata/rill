<script lang="ts">
  import Subheading from "@rilldata/web-common/components/typography/Subheading.svelte";
  import Card from "../../components/card/Card.svelte";
  import CardDescription from "../../components/card/CardDescription.svelte";
  import CardTitle from "../../components/card/CardTitle.svelte";
  import { behaviourEvent } from "../../metrics/initMetrics";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../metrics/service/MetricsTypes";
  import { createRuntimeServiceUnpackExample } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { EMPTY_PROJECT_TITLE } from "./constants";
  import AddCircleOutline from "@rilldata/web-common/components/icons/AddCircleOutline.svelte";
  import { createRuntimeServiceUnpackEmpty } from "../../runtime-client";
  import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
  import { goto } from "$app/navigation";

  const unpackExampleProject = createRuntimeServiceUnpackExample();
  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  const EXAMPLES = [
    {
      name: "rill-cost-monitoring",
      title: "Cost Monitoring",
      description: "Monitoring cloud infrastructure",
      image: "/img/welcome-bg-cost-monitoring.png",
      firstPage: "/files/dashboards/metrics_margin_model_metrics_explore.yaml",
    },
    {
      name: "rill-openrtb-prog-ads",
      title: "OpenRTB Programmatic Ads",
      description: "Real-time Bidding (RTB) advertising",
      image: "/img/welcome-bg-openrtb.png",
      firstPage:
        "/files/dashboards/auction_data_model_metrics_explore_row_policies.yaml",
    },
    {
      name: "rill-github-analytics",
      title: "Github Analytics",
      description: "A Git project's commit activity",
      image: "/img/welcome-bg-github-analytics.png",
      firstPage: "/files/dashboards/duckdb_commits_model_metrics_explore.yaml",
    },
  ];

  let selectedProjectName: string | null = null;

  $: ({ instanceId } = $runtime);

  $: ({ mutateAsync: unpackExample } = $unpackExampleProject);
  $: ({ mutateAsync: unpackEmpty } = $unpackEmptyProject);

  async function unpackProject(example?: (typeof EXAMPLES)[number]) {
    selectedProjectName = example ? example.name : EMPTY_PROJECT_TITLE;

    await behaviourEvent?.fireSplashEvent(
      example
        ? BehaviourEventAction.ExampleAdd
        : BehaviourEventAction.ProjectEmpty,
      BehaviourEventMedium.Card,
      MetricsEventSpace.Workspace,
      example?.name,
    );

    const mutationFunction = example ? unpackExample : unpackEmpty;
    const key = example ? "name" : "title";

    try {
      await mutationFunction({
        instanceId,
        data: {
          [key]: selectedProjectName,
          force: true,
        },
      });

      await asyncWait(400);

      if (example?.firstPage) await goto(example.firstPage);
    } catch {
      selectedProjectName = null;
    }
  }
</script>

<section class="flex flex-col items-center gap-y-5">
  <Subheading>Or jump right into a project.</Subheading>
  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
    <Card
      disabled={!!selectedProjectName}
      isLoading={selectedProjectName === EMPTY_PROJECT_TITLE}
      on:click={() => unpackProject()}
    >
      <AddCircleOutline size="2em" className="text-slate-600" />
      <CardTitle position="middle">Start with an empty project</CardTitle>
    </Card>

    {#each EXAMPLES as example (example.name)}
      <Card
        imageUrl={example.image}
        disabled={!!selectedProjectName}
        isLoading={selectedProjectName === example.name}
        on:click={async () => {
          await unpackProject(example);
        }}
      >
        <CardTitle>{example.title}</CardTitle>
        <CardDescription>{example.description}</CardDescription>
      </Card>
    {/each}
  </div>
</section>
