<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { createDashboardStateSync } from "@rilldata/web-common/features/dashboards/stores/syncDashboardState";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import Spinner from "../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../entity-management/types";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { selectedMockUserStore } from "../granular-access-policies/stores";
  import { useDashboard } from "../selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let metricViewName: string;

  const dashboardStoreReady = createDashboardStateSync(getStateManagers());

  $: initLocalUserPreferenceStore(metricViewName);

  $: dashboard = useDashboard($runtime.instanceId, metricViewName);
  $: mockUserHasNoAccess =
    $selectedMockUserStore && $dashboard.error?.response?.status === 404;
</script>

{#if $dashboardStoreReady.isFetching}
  <div class="grid place-items-center size-full">
    <Spinner status={EntityStatus.Running} size="40px" />
  </div>
{:else if $dashboardStoreReady.error && mockUserHasNoAccess}
  <ErrorPage
    statusCode={$dashboard.error?.response?.status}
    header="This user can't access this dashboard"
    body="The security policy for this dashboard may make contents invisible to you. If you deploy this dashboard, {$selectedMockUserStore?.email} will see a 404."
  />
{:else}
  <slot />
{/if}
