<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import {
    TokenBannerID,
    TokenBannerPriority,
  } from "@rilldata/web-common/components/banner/constants";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: ({ instanceId } = $runtime);

  $: ({ organization, project, dashboard: canvasName } = $page.params);

  // Query the `GetProject` API with cookie-based auth to determine if the user has access to the original dashboard
  $: cookieProjectQuery = createAdminServiceGetProject(organization, project);
  $: ({ data: cookieProject } = $cookieProjectQuery);
  $: if (cookieProject) {
    eventBus.emit("add-banner", {
      id: TokenBannerID,
      priority: TokenBannerPriority,
      message: {
        type: "default",
        message: `Limited view. For full access and features, visit the <a href='/${organization}/${project}/canvas/${canvasName}'>original dashboard</a>.`,
        includesHtml: true,
        iconType: "alert",
      },
    });
  }

  // Clear the banner when navigating away from the Public URL page
  // (We make sure to not clear it when the user interacts with the dashboard)
  onNavigate(({ from, to }) => {
    const currentPath = from?.url.pathname;
    const newPath = to?.url.pathname;
    if (newPath !== currentPath) {
      eventBus.emit("remove-banner", TokenBannerID);
    }
  });
</script>

{#key canvasName}
  <CanvasProvider {canvasName} {instanceId} projectId={project} showBanner>
    <CanvasDashboardEmbed {canvasName} />
  </CanvasProvider>
{/key}
