<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import GlobalDimensionSearch from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearch.svelte";
  import { useExplore } from "@rilldata/web-common/features/explores/selectors";
  import { runtime } from "../../runtime-client/runtime-store";
  import ChatToggle from "../chat/layouts/sidebar/ChatToggle.svelte";
  import ViewAsButton from "../dashboards/granular-access-policies/ViewAsButton.svelte";
  import {
    useDashboardPolicyCheck,
    useRillYamlPolicyCheck,
  } from "../dashboards/granular-access-policies/useSecurityPolicyCheck";
  import StateManagersProvider from "../dashboards/state-managers/StateManagersProvider.svelte";
  import { featureFlags } from "../feature-flags";
  import { timeAgo } from "@rilldata/web-common/lib/time/relative-time";

  export let exploreName: string;

  $: ({ instanceId } = $runtime);

  $: exploreQuery = useExplore(instanceId, exploreName);
  $: metricsViewName = $exploreQuery.data?.metricsView?.meta?.name?.name ?? "";

  $: exploreFilePath = $exploreQuery.data?.explore?.meta?.filePaths?.[0] ?? "";
  $: metricsViewFilePath =
    $exploreQuery.data?.metricsView?.meta?.filePaths?.[0] ?? "";
  $: explorePolicyCheck = useDashboardPolicyCheck(instanceId, exploreFilePath);
  $: metricsPolicyCheck = useDashboardPolicyCheck(
    instanceId,
    metricsViewFilePath,
  );
  $: rillYamlPolicyCheck = useRillYamlPolicyCheck(instanceId);

  // Get last refreshed date for preview mode
  $: lastRefreshedDate = $exploreQuery.data?.metricsView?.metricsView?.state
    ?.dataRefreshedOn
    ? new Date($exploreQuery.data.metricsView.metricsView.state.dataRefreshedOn)
    : null;

  const { dashboardChat } = featureFlags;
</script>

<div class="flex gap-2 flex-shrink-0 ml-auto">
  {#if lastRefreshedDate}
    <Tooltip distance={8}>
      <div class="text-[11px] flex items-center text-fg-secondary">
        Last refreshed {timeAgo(lastRefreshedDate)}
      </div>
      <TooltipContent slot="tooltip-content">
        {lastRefreshedDate.toLocaleString()}
      </TooltipContent>
    </Tooltip>
  {/if}
  {#if $explorePolicyCheck.data || $metricsPolicyCheck.data || $rillYamlPolicyCheck.data}
    <ViewAsButton />
  {/if}
  <StateManagersProvider {metricsViewName} {exploreName} let:ready>
    {#if $dashboardChat}
      <ChatToggle />
    {/if}
    {#if ready}
      <GlobalDimensionSearch />
    {/if}
  </StateManagersProvider>
</div>
