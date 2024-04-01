<script lang="ts">
  import Chart from "@rilldata/web-common/components/icons/Chart.svelte";
  import Pivot from "@rilldata/web-common/components/icons/Pivot.svelte";
  import Tab from "./Tab.svelte";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";

  $: viewing = $page.params.view;
  $: metricsViewName = $page.params.name;

  const tabs = [
    {
      label: "Explore",
      route: undefined,
      Icon: Chart,
    },
    {
      label: "Pivot",
      route: "pivot",
      Icon: Pivot,
      beta: true,
    },
  ];

  async function handleTabChange(route: string | undefined) {
    if (viewing === route) return;

    await goto(
      route
        ? `/dashboard/${metricsViewName}/${route}`
        : `/dashboard/${metricsViewName}/`,
    );

    await behaviourEvent.fireNavigationEvent(
      metricsViewName,
      BehaviourEventMedium.Tab,
      MetricsEventSpace.Workspace,
      MetricsEventScreenName.Dashboard,
      route === "pivot"
        ? MetricsEventScreenName.Pivot
        : MetricsEventScreenName.Explore,
    );
  }
</script>

<div class="mr-4">
  <div class="flex gap-x-2">
    {#each tabs as { label, Icon, beta, route } (label)}
      {@const selected = viewing === route}
      <Tab {selected} on:click={() => handleTabChange(route)}>
        <Icon />
        <div class="flex gap-x-1 items-center group">
          {label}
          {#if beta}
            <Tag height={18} color={selected ? "blue" : "gray"}>BETA</Tag>
          {/if}
        </div>
      </Tab>
    {/each}
  </div>
</div>
