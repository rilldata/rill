<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import { useShareableURLMetricsView } from "@rilldata/web-admin/features/shareable-urls/selectors";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import LoadingPage from "@rilldata/web-common/components/LoadingPage.svelte";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import DashboardURLStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardURLStateProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/stores/DashboardStateProvider.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: ({ organization, project } = $page.params);

  $: ({ instanceId } = $runtime);
  $: metricsViewQuery = useShareableURLMetricsView(instanceId, true);
  $: ({
    data: resource,
    error: resourceError,
    isLoading: resourceIsLoading,
  } = $metricsViewQuery);
  $: metricsViewName = resource?.meta?.name?.name;

  // Query the `GetProject` API with cookie-based auth to determine if the user has access to the original dashboard
  $: cookieProjectQuery = createAdminServiceGetProject(organization, project);
  $: ({ data: cookieProject } = $cookieProjectQuery);
  $: if (cookieProject) {
    eventBus.emit("banner", {
      message: `Limited view. For full access and features, visit the <a href='/${organization}/${project}/${metricsViewName}'>original dashboard</a>.`,
      includesHtml: true,
    });
  }

  // When navigating away from this page, clear the banner (if any)
  onNavigate(() => {
    eventBus.emit("banner", null);
  });
</script>

{#if resourceIsLoading}
  <LoadingPage />
{:else if resourceError}
  <ErrorPage
    header="Unable to open shareable link"
    body={resourceError?.response?.data?.message}
  />
{:else if resource}
  {#key metricsViewName}
    <StateManagersProvider {metricsViewName}>
      <DashboardStateProvider metricViewName={metricsViewName}>
        <DashboardURLStateProvider metricViewName={metricsViewName}>
          <DashboardThemeProvider>
            <Dashboard metricViewName={metricsViewName} />
          </DashboardThemeProvider>
        </DashboardURLStateProvider>
      </DashboardStateProvider>
    </StateManagersProvider>
  {/key}
{/if}
