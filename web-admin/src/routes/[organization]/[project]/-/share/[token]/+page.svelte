<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import LoadingPage from "@rilldata/web-common/components/LoadingPage.svelte";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import DashboardURLStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardURLStateProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/stores/DashboardStateProvider.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: ({ organization, project } = $page.params);

  $: ({ instanceId } = $runtime);
  // Use the ListResources API to get the target dashboard
  // The provided JWT will only have access to one dashboard, so we can assume the first one is the correct one
  $: resourceQuery = createRuntimeServiceListResources(
    instanceId,
    {
      kind: ResourceKind.MetricsView,
    },
    {
      query: {
        select: (data) => ({
          resource: data.resources[0],
        }),
      },
    },
  );
  $: ({
    data: resource,
    error: resourceError,
    isLoading: resourceIsLoading,
  } = $resourceQuery);

  $: dashboard = resource?.resource?.meta?.name?.name;

  // Query the `GetProject` API with cookie-based auth to determine if the user has access to the original dashboard
  $: cookieProjectQuery = createAdminServiceGetProject(organization, project);
  $: ({ data: cookieProject } = $cookieProjectQuery);
  $: if (cookieProject) {
    eventBus.emit("banner", {
      message: `Limited view. For full access and features, visit the <a href='/${organization}/${project}/${dashboard}'>original dashboard</a>.`,
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
  {#key dashboard}
    <StateManagersProvider metricsViewName={dashboard}>
      <DashboardStateProvider metricViewName={dashboard}>
        <DashboardURLStateProvider metricViewName={dashboard}>
          <DashboardThemeProvider>
            <Dashboard metricViewName={dashboard} />
          </DashboardThemeProvider>
        </DashboardURLStateProvider>
      </DashboardStateProvider>
    </StateManagersProvider>
  {/key}
{/if}
