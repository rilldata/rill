<script lang="ts">
  import Chart from "@rilldata/web-common/components/icons/Chart.svelte";
  import Pivot from "@rilldata/web-common/components/icons/Pivot.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import Tab from "./Tab.svelte";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  // import { featureFlags } from "../../feature-flags";

  // const { pivot: pivotAllowed } = featureFlags;

  const StateManagers = getStateManagers();

  const {
    exploreName,
    selectors: {
      pivot: { showPivot },
    },
  } = StateManagers;

  const tabs = [
    {
      label: "Explore",
      Icon: Chart,
      beta: false,
    },
    {
      label: "Pivot",
      Icon: Pivot,
      beta: false,
    },
  ];

  $: currentTabIndex = $showPivot ? 1 : 0;

  function handleTabChange(index: number) {
    if (currentTabIndex === index) return;
    const selectedTab = tabs[index];

    metricsExplorerStore.setPivotMode(
      $exploreName,
      selectedTab.label === "Pivot",
    );

    // We do not have behaviour events in cloud
    behaviourEvent?.fireNavigationEvent(
      $exploreName,
      BehaviourEventMedium.Tab,
      MetricsEventSpace.Workspace,
      MetricsEventScreenName.Dashboard,
      selectedTab.label === "Pivot"
        ? MetricsEventScreenName.Pivot
        : MetricsEventScreenName.Explore,
    );
  }
</script>

<div class="mr-4">
  <div class="flex gap-x-2">
    {#each tabs as { label, Icon, beta }, i (label)}
      {@const selected = currentTabIndex === i}
      <Tab {selected} on:click={() => handleTabChange(i)}>
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
