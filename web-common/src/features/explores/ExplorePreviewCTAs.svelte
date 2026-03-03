<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import MetricsViewIcon from "@rilldata/web-common/components/icons/MetricsViewIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import GlobalDimensionSearch from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearch.svelte";
  import { useExplore } from "@rilldata/web-common/features/explores/selectors";
  import { Button } from "../../components/button";
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
  export let inPreviewMode = false;

  $: ({ instanceId } = $runtime);

  $: exploreQuery = useExplore(instanceId, exploreName);
  $: exploreFilePath = $exploreQuery.data?.explore?.meta?.filePaths?.[0] ?? "";
  $: metricsViewFilePath =
    $exploreQuery.data?.metricsView?.meta?.filePaths?.[0] ?? "";
  $: metricsViewName = $exploreQuery.data?.metricsView?.meta?.name?.name ?? "";

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

  const { readOnly, dashboardChat } = featureFlags;
</script>

<div class="flex gap-2 flex-shrink-0 ml-auto">
  {#if lastRefreshedDate}
    <Tooltip distance={8}>
      <div class="text-[11px] flex items-center" style="color: var(--fg-muted)">
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
  {#if !$readOnly && !inPreviewMode}
    <DropdownMenu.Root>
      <DropdownMenu.Trigger asChild let:builder>
        <Button type="secondary" builders={[builder]}>
          Edit
          <CaretDownIcon />
        </Button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        <DropdownMenu.Item href={`/files${exploreFilePath}`}>
          <ExploreIcon size="16px" />
          Explore dashboard
        </DropdownMenu.Item>
        <DropdownMenu.Item href={`/files${metricsViewFilePath}`}>
          <MetricsViewIcon size="16px" />
          Metrics View
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {/if}
</div>
