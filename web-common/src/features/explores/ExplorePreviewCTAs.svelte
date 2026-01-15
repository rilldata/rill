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
  import { resourceColorMapping } from "../entity-management/resource-icon-mapping";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import { featureFlags } from "../feature-flags";
  import { BookmarkIcon, BellPlusIcon } from "lucide-svelte";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
  import { DateTime, Duration } from "luxon";

  const disabledTooltip = "Deploy your project to access this feature";

  function timeAgo(date: Date): string {
    const now = DateTime.now();
    const then = DateTime.fromJSDate(date);
    const diff = Duration.fromMillis(now.diff(then).milliseconds);

    if (diff.as("minutes") < 1) return "Just now";

    const minutes = Math.round(diff.as("minutes"));
    if (diff.as("hours") < 1) return `${minutes} ${minutes === 1 ? "minute" : "minutes"} ago`;

    const hours = Math.round(diff.as("hours"));
    if (diff.as("days") < 1) return `${hours} ${hours === 1 ? "hour" : "hours"} ago`;

    const days = Math.round(diff.as("days"));
    return `${days} ${days === 1 ? "day" : "days"} ago`;
  }

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
  $: lastRefreshedDate = $exploreQuery.data?.metricsView?.metricsView?.state?.dataRefreshedOn
    ? new Date($exploreQuery.data.metricsView.metricsView.state.dataRefreshedOn)
    : null;

  const { readOnly, dashboardChat, alerts: alertsFlag } = featureFlags;
</script>

<div class="flex gap-2 flex-shrink-0 ml-auto">
  {#if lastRefreshedDate}
    <Tooltip distance={8}>
      <div class="text-[11px] text-gray-600 flex items-center">
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
  <StateManagersProvider {metricsViewName} {exploreName}>
    {#if $dashboardChat}
      <ChatToggle />
    {/if}
    <GlobalDimensionSearch />
  </StateManagersProvider>
  {#if inPreviewMode}
    <Tooltip distance={8}>
      <Button type="secondary" compact gray href="https://docs.rilldata.com/explore/bookmarks" target="_blank" label="Home">
        <HomeBookmark size="16px" />
      </Button>
      <TooltipContent slot="tooltip-content">
        <div>Set a default view for all users.</div>
        <div class=" text-[10px] mt-1">Click to learn more</div>
      </TooltipContent>
    </Tooltip>
    <Tooltip distance={8}>
      <Button type="secondary" compact gray href="https://docs.rilldata.com/explore/bookmarks" target="_blank" label="Bookmarks">
        <BookmarkIcon class="inline-flex" size="16px" />
      </Button>
      <TooltipContent slot="tooltip-content">
        <div>Save and share dashboard views.</div>
        <div class=" text-[10px] mt-1">Click to learn more</div>
      </TooltipContent>
    </Tooltip>
    {#if $alertsFlag}
      <Tooltip distance={8}>
        <Button type="secondary" compact gray href="https://docs.rilldata.com/explore/alerts" target="_blank" label="Create alert">
          <BellPlusIcon class="inline-flex" size="16px" />
        </Button>
        <TooltipContent slot="tooltip-content">
          <div>Get notified when metrics change.</div>
          <div class=" text-[10px] mt-1">Click to learn more</div>
        </TooltipContent>
      </Tooltip>
    {/if}
    <Tooltip distance={8}>
      <Button type="secondary" gray href="https://docs.rilldata.com/explore/public-url" target="_blank">Share</Button>
      <TooltipContent slot="tooltip-content">
        <div>Share dashboards with your team.</div>
        <div class=" text-[10px] mt-1">Click to learn more</div>
      </TooltipContent>
    </Tooltip>
  {/if}
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
          <ExploreIcon
            color={resourceColorMapping[ResourceKind.Explore]}
            size="16px"
          />
          Explore dashboard
        </DropdownMenu.Item>
        <DropdownMenu.Item href={`/files${metricsViewFilePath}`}>
          <MetricsViewIcon
            color={resourceColorMapping[ResourceKind.MetricsView]}
            size="16px"
          />
          Metrics View
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {/if}
</div>
