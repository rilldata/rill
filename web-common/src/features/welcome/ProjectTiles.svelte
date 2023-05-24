<script lang="ts">
  import Subheading from "@rilldata/web-common/components/typography/Subheading.svelte";
  import Card from "../../components/card/Card.svelte";
  import CardDescription from "../../components/card/CardDescription.svelte";
  import CardHeader from "../../components/card/CardHeader.svelte";
  import { createRuntimeServiceUnpackExample } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import EmptyProject from "./EmptyProject.svelte";

  const EXAMPLES = [
    {
      name: "rill-cost-monitoring",
      title: "Cost Monitoring",
      description: "Monitoring cloud infrastructure",
      image:
        "bg-[url('/img/welcome-bg-cost-monitoring.png')] bg-no-repeat bg-cover",
    },
    {
      name: "rill-openrtb-prog-ads",
      title: "OpenRTB Programmatic Ads",
      description: "Real-time Bidding (RTB) advertising",
      image: "bg-[url('/img/welcome-bg-openrtb.png')] bg-no-repeat bg-cover",
    },
    {
      name: "rill-311-ops",
      title: "311 Call Center Operations",
      description: "Citizen reports across US cities",
      image: "bg-[url('/img/welcome-bg-311.png')] bg-no-repeat bg-cover",
    },
  ];

  const unpackExampleProject = createRuntimeServiceUnpackExample();
</script>

<section class="flex flex-col items-center gap-y-5 mt-4 mb-6">
  <Subheading>Or jump right into a project.</Subheading>
  <div class="flex flex-row justify-center gap-4 flex-wrap">
    <EmptyProject />
    {#each EXAMPLES as example}
      <Card
        bgClasses={example.image}
        on:click={() => {
          $unpackExampleProject.mutate({
            instanceId: $runtime.instanceId,
            data: {
              name: example.name,
            },
          });
        }}
      >
        <CardHeader>{example.title}</CardHeader>
        <CardDescription>{example.description}</CardDescription>
      </Card>
    {/each}
  </div>
</section>
