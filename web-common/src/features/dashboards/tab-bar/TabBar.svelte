<script lang="ts">
  import Chart from "@rilldata/web-common/components/icons/Chart.svelte";
  import Pivot from "@rilldata/web-common/components/icons/Pivot.svelte";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import Tab from "./Tab.svelte";
  import type { ComponentType } from "svelte";

  type TabName = MetricsEventScreenName.Pivot | MetricsEventScreenName.Explore;
  type TabData = { label: string; Icon: ComponentType; beta?: true };

  const tabs = new Map<TabName, TabData>([
    [
      MetricsEventScreenName.Explore,
      {
        label: "Explore",
        Icon: Chart,
      },
    ],
    [
      MetricsEventScreenName.Pivot,
      {
        label: "Pivot",
        Icon: Pivot,
      },
    ],
  ]);

  export let hidePivot: boolean = false;
  export let exploreName: string;
  export let onPivot = false;

  $: currentTab = onPivot
    ? MetricsEventScreenName.Pivot
    : MetricsEventScreenName.Explore;

  async function handleTabChange(tab: MetricsEventScreenName) {
    // We do not have behaviour events in cloud
    await behaviourEvent?.fireNavigationEvent(
      exploreName,
      BehaviourEventMedium.Tab,
      MetricsEventSpace.Workspace,
      MetricsEventScreenName.Dashboard,
      tab,
    );
  }
</script>

<div class="mr-4">
  <div class="flex gap-x-2">
    {#each tabs as [tab, { label, Icon, beta }] (tab)}
      {#if !hidePivot || tab === MetricsEventScreenName.Explore}
        {@const selected = tab === currentTab}
        <Tab
          {selected}
          href="?view={tab}"
          on:click={() => handleTabChange(tab)}
        >
          <Icon />
          <div class="flex gap-x-1 items-center group">
            {label}
            {#if beta}
              <Tag height={18} color={selected ? "blue" : "gray"}>BETA</Tag>
            {/if}
          </div>
        </Tab>
      {/if}
    {/each}
  </div>
</div>
