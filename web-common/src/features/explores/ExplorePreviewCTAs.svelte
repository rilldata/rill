<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import MetricsViewIcon from "@rilldata/web-common/components/icons/MetricsViewIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
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
  import { BookmarkIcon, BellPlusIcon } from "lucide-svelte";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
  import { timeAgo } from "@rilldata/web-common/lib/time/relative-time";
  import Rocket from "svelte-radix/Rocket.svelte";

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
  <StateManagersProvider {metricsViewName} {exploreName} let:ready>
    {#if $dashboardChat}
      <ChatToggle />
    {/if}
    {#if ready}
      <GlobalDimensionSearch />
    {/if}
  </StateManagersProvider>
  {#if inPreviewMode}
    <!-- Home View -->
    <Popover.Root>
      <Popover.Trigger asChild let:builder>
        <Button type="secondary" compact gray builders={[builder]} label="Home">
          <HomeBookmark size="16px" />
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" class="w-64">
        <div class="flex flex-col gap-y-3">
          <div>
            <p class="text-sm font-semibold" style="color: var(--fg-primary)">Default Home View</p>
            <p class="text-xs mt-1" style="color: var(--fg-muted)">Set a default dashboard view for all users. They'll see the exact filters and settings you've configured.</p>
          </div>
          <Button type="primary" href="/deploy" compact>
            <Rocket size="14px" />
            Deploy to unlock
          </Button>
        </div>
      </Popover.Content>
    </Popover.Root>

    <!-- Bookmarks -->
    <Popover.Root>
      <Popover.Trigger asChild let:builder>
        <Button type="secondary" compact gray builders={[builder]} label="Bookmarks">
          <BookmarkIcon class="inline-flex" size="16px" />
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" class="w-64">
        <div class="flex flex-col gap-y-3">
          <div>
            <p class="text-sm font-semibold" style="color: var(--fg-primary)">Bookmarks</p>
            <p class="text-xs mt-1" style="color: var(--fg-muted)">Save and share specific dashboard views with your team. Keep track of important metric combinations and filter states.</p>
          </div>
          <Button type="primary" href="/deploy" compact>
            <Rocket size="14px" />
            Deploy to unlock
          </Button>
        </div>
      </Popover.Content>
    </Popover.Root>

    <!-- Alerts -->
    {#if $alertsFlag}
      <Popover.Root>
        <Popover.Trigger asChild let:builder>
          <Button type="secondary" compact gray builders={[builder]} label="Create alert">
            <BellPlusIcon class="inline-flex" size="16px" />
          </Button>
        </Popover.Trigger>
        <Popover.Content align="end" class="w-64">
          <div class="flex flex-col gap-y-3">
            <div>
              <p class="text-sm font-semibold" style="color: var(--fg-primary)">Alerts</p>
              <p class="text-xs mt-1" style="color: var(--fg-muted)">Get notified when metrics change beyond thresholds you define. Receive alerts via email, Slack, or webhooks.</p>
            </div>
            <Button type="primary" href="/deploy" compact>
              <Rocket size="14px" />
              Deploy to unlock
            </Button>
          </div>
        </Popover.Content>
      </Popover.Root>
    {/if}

    <!-- Share -->
    <Popover.Root>
      <Popover.Trigger asChild let:builder>
        <Button type="secondary" gray builders={[builder]}>Share</Button>
      </Popover.Trigger>
      <Popover.Content align="end" class="w-64">
        <div class="flex flex-col gap-y-3">
          <div>
            <p class="text-sm font-semibold" style="color: var(--fg-primary)">Share Dashboards</p>
            <p class="text-xs mt-1" style="color: var(--fg-muted)">Share dashboards with your team using public URLs, embed them in other tools, or set up role-based access controls.</p>
          </div>
          <Button type="primary" href="/deploy" compact>
            <Rocket size="14px" />
            Deploy to unlock
          </Button>
        </div>
      </Popover.Content>
    </Popover.Root>
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
