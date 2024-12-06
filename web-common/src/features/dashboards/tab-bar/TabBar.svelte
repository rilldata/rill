<script lang="ts">
  import { page } from "$app/stores";
  import Chart from "@rilldata/web-common/components/icons/Chart.svelte";
  import Pivot from "@rilldata/web-common/components/icons/Pivot.svelte";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { getUrlForWebView } from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-store";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { V1ExploreWebView } from "@rilldata/web-common/runtime-client";
  import Tab from "./Tab.svelte";

  export let hidePivot: boolean = false;

  const StateManagers = getStateManagers();

  const {
    exploreName,
    selectors: {
      pivot: { showPivot },
    },
    defaultExploreState,
  } = StateManagers;

  $: tabs = [
    {
      label: "Explore",
      Icon: Chart,
      beta: false,
      href: getUrlForWebView(
        $page.url,
        V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE,
        $defaultExploreState,
      ),
    },
    ...(hidePivot
      ? []
      : [
          {
            label: "Pivot",
            Icon: Pivot,
            beta: false,
            href: getUrlForWebView(
              $page.url,
              V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
              $defaultExploreState,
            ),
          },
        ]),
  ];

  $: currentTabIndex = $showPivot ? 1 : 0;

  function handleTabChange(index: number) {
    if (currentTabIndex === index) return;
    const selectedTab = tabs[index];

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
    {#each tabs as { label, Icon, beta, href }, i (label)}
      {@const selected = currentTabIndex === i}
      <Tab {selected} {href} on:click={() => handleTabChange(i)}>
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
