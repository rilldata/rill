<script lang="ts">
  import GlobalDimensionSearch from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearch.svelte";
  import { useExplore } from "@rilldata/web-common/features/explores/selectors";
  import { useRuntimeClient } from "../../runtime-client/v2";
  import ChatToggle from "../chat/layouts/sidebar/ChatToggle.svelte";
  import {
    dashboardChatActions,
    dashboardChatOpen,
  } from "../chat/layouts/sidebar/sidebar-store";
  import ViewAsButton from "../dashboards/granular-access-policies/ViewAsButton.svelte";
  import {
    useDashboardPolicyCheck,
    useRillYamlPolicyCheck,
  } from "../dashboards/granular-access-policies/useSecurityPolicyCheck";
  import StateManagersProvider from "../dashboards/state-managers/StateManagersProvider.svelte";
  import { featureFlags } from "../feature-flags";
  import ExploreEditDropdown from "./ExploreEditDropdown.svelte";

  export let exploreName: string;

  const runtimeClient = useRuntimeClient();

  $: exploreQuery = useExplore(runtimeClient, exploreName);
  $: exploreFilePath = $exploreQuery.data?.explore?.meta?.filePaths?.[0] ?? "";
  $: metricsViewFilePath =
    $exploreQuery.data?.metricsView?.meta?.filePaths?.[0] ?? "";
  $: metricsViewName = $exploreQuery.data?.metricsView?.meta?.name?.name ?? "";

  $: explorePolicyCheck = useDashboardPolicyCheck(
    runtimeClient,
    exploreFilePath,
  );
  $: metricsPolicyCheck = useDashboardPolicyCheck(
    runtimeClient,
    metricsViewFilePath,
  );
  $: rillYamlPolicyCheck = useRillYamlPolicyCheck(runtimeClient);

  const { readOnly, dashboardChat } = featureFlags;
</script>

<div class="flex gap-2 flex-shrink-0 ml-auto">
  {#if $explorePolicyCheck.data || $metricsPolicyCheck.data || $rillYamlPolicyCheck.data}
    <ViewAsButton />
  {/if}
  <StateManagersProvider {metricsViewName} {exploreName} let:ready>
    {#if $dashboardChat}
      <ChatToggle open={dashboardChatOpen} actions={dashboardChatActions} />
    {/if}
    {#if ready}
      <GlobalDimensionSearch />
    {/if}
  </StateManagersProvider>
  {#if !$readOnly}
    <ExploreEditDropdown {exploreName} />
  {/if}
</div>
