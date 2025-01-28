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

  const unpackExampleProject = createRuntimeServiceUnpackExample();

  const EXAMPLES = [
    {
      name: "rill-cost-monitoring",
      title: "Cost Monitoring",
      description: "Monitoring cloud infrastructure",
      image: "/img/welcome-bg-cost-monitoring.png",
    },
    {
      name: "rill-openrtb-prog-ads",
      title: "OpenRTB Programmatic Ads",
      description: "Real-time Bidding (RTB) advertising",
      image: "/img/welcome-bg-openrtb.png",
    },
    {
      name: "rill-github-analytics",
      title: "Github Analytics",
      description: "A Git project's commit activity",
      image: "/img/welcome-bg-github-analytics.png",
    },
  ];

  let selectedProjectName: string | null = null;

  $: ({ instanceId } = $runtime);

  $: ({ mutateAsync: unpackExample } = $unpackExampleProject);

  async function unpackProject(example: (typeof EXAMPLES)[number]) {
    selectedProjectName = example.name;

    await behaviourEvent?.fireSplashEvent(
      BehaviourEventAction.ExampleAdd,
      BehaviourEventMedium.Card,
      MetricsEventSpace.Workspace,
      selectedProjectName,
    );

    try {
      await unpackExample({
        instanceId,
        data: {
          name: selectedProjectName,
          force: true,
        },
      });

      setTimeout(() => {
        if (window.location.search.includes("redirect=true"))
          window.location.reload();
      }, 5000);
    } catch {
      selectedProjectName = null;
    }
  }
</script>

<section class="flex flex-col items-center gap-y-5">
  <Subheading>Or jump right into a project.</Subheading>
  <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
    {#each EXAMPLES as example (example.name)}
      <Card
        redirect
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
