<script lang="ts">
  import { goto } from "$app/navigation";
  import Subheading from "@rilldata/web-common/components/typography/Subheading.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Card from "../../components/card/Card.svelte";
  import CardDescription from "../../components/card/CardDescription.svelte";
  import CardTitle from "../../components/card/CardTitle.svelte";
  import { overlay } from "../../layout/overlay-store";
  import {
    createRuntimeServiceReconcile,
    createRuntimeServiceUnpackExample,
  } from "../../runtime-client";
  import { invalidateAfterReconcile } from "../../runtime-client/invalidation";
  import { runtime } from "../../runtime-client/runtime-store";
  import EmptyProject from "./EmptyProject.svelte";
  import { behaviourEvent } from "../../metrics/initMetrics";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../metrics/service/MetricsTypes";

  const queryClient = useQueryClient();

  const EXAMPLES = [
    {
      name: "rill-cost-monitoring",
      title: "Cost Monitoring",
      description: "Monitoring cloud infrastructure",
      image:
        "bg-[url('/img/welcome-bg-cost-monitoring.png')] bg-no-repeat bg-cover",
      firstPage: "/dashboard/customer_margin_dash",
    },
    {
      name: "rill-openrtb-prog-ads",
      title: "OpenRTB Programmatic Ads",
      description: "Real-time Bidding (RTB) advertising",
      image: "bg-[url('/img/welcome-bg-openrtb.png')] bg-no-repeat bg-cover",
      firstPage: "dashboard/auction",
    },
    {
      name: "rill-311-ops",
      title: "311 Call Center Operations",
      description: "Citizen reports across US cities",
      image: "bg-[url('/img/welcome-bg-311.png')] bg-no-repeat bg-cover",
      firstPage: "dashboard/call_center_metrics",
    },
  ];

  let firstPage: string;
  const unpackExampleProject = createRuntimeServiceUnpackExample({
    mutation: {
      onSuccess: () => {
        overlay.set({
          title: "Loading the example project",
          message: "Hang tight! This might take a minute or two.",
        });
        $reconcile.mutate({
          instanceId: $runtime.instanceId,
          data: undefined,
        });
      },
    },
  });

  const reconcile = createRuntimeServiceReconcile({
    mutation: {
      onSuccess: (response) => {
        invalidateAfterReconcile(queryClient, $runtime.instanceId, response);
        // behaviourEvent.fireSourceSuccessEvent(
        //   BehaviourEventMedium.Card,
        //   MetricsEventScreenName.Splash,
        //   MetricsEventSpace.Workspace,
        //   SourceConnectionType.Https,
        //   fileType,
        //   false
        // );
        goto(firstPage);
      },
      onError: (err) => {
        console.error(err);
      },
      onSettled: () => {
        overlay.set(null);
      },
    },
  });

  function startWithExampleProject(example: (typeof EXAMPLES)[number]) {
    behaviourEvent.fireSplashEvent(
      BehaviourEventAction.ExampleAdd,
      BehaviourEventMedium.Card,
      MetricsEventSpace.Workspace,
      example.name
    );

    firstPage = example.firstPage;
    $unpackExampleProject.mutate({
      instanceId: $runtime.instanceId,
      data: {
        name: example.name,
        force: true,
      },
    });
  }
</script>

<section class="flex flex-col items-center gap-y-5 mt-4 mb-6">
  <Subheading>Or jump right into a project.</Subheading>
  <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
    <EmptyProject />
    {#each EXAMPLES as example}
      <Card
        bgClasses={example.image}
        on:click={() => {
          startWithExampleProject(example);
        }}
      >
        <CardTitle>{example.title}</CardTitle>
        <CardDescription>{example.description}</CardDescription>
      </Card>
    {/each}
  </div>
</section>
