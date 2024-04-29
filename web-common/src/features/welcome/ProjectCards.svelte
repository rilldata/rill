<script lang="ts">
  import { goto } from "$app/navigation";
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
  import EmptyProject from "./EmptyProject.svelte";

  const EXAMPLES = [
    {
      name: "rill-cost-monitoring",
      title: "Cost Monitoring",
      description: "Monitoring cloud infrastructure",
      image:
        "bg-[url('$img/welcome-bg-cost-monitoring.png')] bg-no-repeat bg-cover",
      firstPage: "/files/dashboards/customer_margin_dash.yaml",
    },
    {
      name: "rill-openrtb-prog-ads",
      title: "OpenRTB Programmatic Ads",
      description: "Real-time Bidding (RTB) advertising",
      image: "bg-[url('$img/welcome-bg-openrtb.png')] bg-no-repeat bg-cover",
      firstPage: "/files/dashboards/auction.yaml",
    },
    {
      name: "rill-github-analytics",
      title: "Github Analytics",
      description: "A Git project's commit activity",
      image:
        "bg-[url('$img/welcome-bg-github-analytics.png')] bg-no-repeat bg-cover",
      firstPage: "/files/dashboards/duckdb_commits.yaml",
    },
  ];

  const unpackExampleProject = createRuntimeServiceUnpackExample();

  async function startWithExampleProject(example: (typeof EXAMPLES)[number]) {
    await behaviourEvent?.fireSplashEvent(
      BehaviourEventAction.ExampleAdd,
      BehaviourEventMedium.Card,
      MetricsEventSpace.Workspace,
      example.name,
    );

    const firstPage = example.firstPage;
    await $unpackExampleProject.mutateAsync({
      instanceId: $runtime.instanceId,
      data: {
        name: example.name,
        force: true,
      },
    });
    await goto(firstPage);
  }
</script>

<section class="flex flex-col items-center gap-y-5">
  <Subheading>Or jump right into a project.</Subheading>
  <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
    <EmptyProject />
    {#each EXAMPLES as example}
      <Card
        bgClasses={example.image}
        on:click={async () => {
          await startWithExampleProject(example);
        }}
      >
        <CardTitle>{example.title}</CardTitle>
        <CardDescription>{example.description}</CardDescription>
      </Card>
    {/each}
  </div>
</section>
