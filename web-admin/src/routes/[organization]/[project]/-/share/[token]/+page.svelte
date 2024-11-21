<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import DashboardURLStateSync from "@rilldata/web-common/features/dashboards/url-state/DashboardURLStateSync.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { createRuntimeServiceGetExplore } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({
    basePreset,
    partialExploreState,
    token: { resourceName },
  } = data);
  $: ({ organization, project } = $page.params);

  // Query the `GetProject` API with cookie-based auth to determine if the user has access to the original dashboard
  $: cookieProjectQuery = createAdminServiceGetProject(organization, project);
  $: ({ data: cookieProject } = $cookieProjectQuery);
  $: if (cookieProject) {
    eventBus.emit("banner", {
      type: "default",
      message: `Limited view. For full access and features, visit the <a href='/${organization}/${project}/explore/${resourceName}'>original dashboard</a>.`,
      includesHtml: true,
      iconType: "alert",
    });
  }

  // Call `GetExplore` to get the Explore's metrics view
  $: exploreQuery = createRuntimeServiceGetExplore($runtime.instanceId, {
    name: resourceName,
  });
  $: ({ data: explore } = $exploreQuery);

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

{#key resourceName}
  {#if explore?.metricsView}
    <StateManagersProvider
      metricsViewName={explore.metricsView.meta.name.name}
      exploreName={resourceName}
    >
      <DashboardURLStateSync {basePreset} {partialExploreState}>
        <DashboardThemeProvider>
          <Dashboard
            exploreName={resourceName}
            metricsViewName={explore.metricsView.meta.name.name}
          />
        </DashboardThemeProvider>
      </DashboardURLStateSync>
    </StateManagersProvider>
  {/if}
{/key}
