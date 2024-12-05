<script lang="ts">
  import { page } from "$app/stores";
  import Chart from "@rilldata/web-common/components/icons/Chart.svelte";
  import Pivot from "@rilldata/web-common/components/icons/Pivot.svelte";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { getUrlForWebView } from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-store";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { V1ExploreWebView } from "@rilldata/web-common/runtime-client";
  import Tab from "./Tab.svelte";
  import type { ComponentType } from "svelte";

  const tabs = new Map<
    V1ExploreWebView,
    { label: string; Icon: ComponentType; beta?: true }
  >([
    [
      V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE,
      {
        label: "Explore",
        Icon: Chart,
      },
    ],
    [
      V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
      {
        label: "Pivot",
        Icon: Pivot,
      },
    ],
  ]);

  export let hidePivot: boolean = false;
  export let exploreName: string;
  export let view: V1ExploreWebView;

  $: ({ url } = $page);

  async function handleTabChange(tab: V1ExploreWebView) {
    // We do not have behaviour events in cloud
    await behaviourEvent?.fireNavigationEvent(
      exploreName,
      BehaviourEventMedium.Tab,
      MetricsEventSpace.Workspace,
      MetricsEventScreenName.Dashboard,
      tab === V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT
        ? MetricsEventScreenName.Pivot
        : MetricsEventScreenName.Explore,
    );
  }
</script>

<div class="mr-4">
  <div class="flex gap-x-2">
    {#each tabs as [tab, { label, Icon, beta }] (tab)}
      {#if !hidePivot || tab === V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE}
        {@const selected = view === tab}
        <Tab
          {selected}
          href={getUrlForWebView(url, tab)}
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
