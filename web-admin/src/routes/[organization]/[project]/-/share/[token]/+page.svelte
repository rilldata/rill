<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import DashboardURLStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardURLStateProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/stores/DashboardStateProvider.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ metricsView } = data.token);

  $: ({ organization, project } = $page.params);

  // Query the `GetProject` API with cookie-based auth to determine if the user has access to the original dashboard
  $: cookieProjectQuery = createAdminServiceGetProject(organization, project);
  $: ({ data: cookieProject } = $cookieProjectQuery);
  $: if (cookieProject) {
    eventBus.emit("banner", {
      message: `Limited view. For full access and features, visit the <a href='/${organization}/${project}/${metricsView}'>original dashboard</a>.`,
      includesHtml: true,
    });
  }

  // Clear the banner when navigating away from the Public URL page
  // (We make sure to not clear it when the user interacts with the dashboard)
  onNavigate(({ from, to }) => {
    const currentPath = from?.url.pathname;
    const newPath = to?.url.pathname;
    if (newPath !== currentPath) {
      eventBus.emit("banner", null);
    }
  });
</script>

{#key metricsView}
  <StateManagersProvider metricsViewName={metricsView}>
    <DashboardStateProvider metricViewName={metricsView}>
      <DashboardURLStateProvider metricViewName={metricsView}>
        <DashboardThemeProvider>
          <Dashboard metricViewName={metricsView} />
        </DashboardThemeProvider>
      </DashboardURLStateProvider>
    </DashboardStateProvider>
  </StateManagersProvider>
{/key}
