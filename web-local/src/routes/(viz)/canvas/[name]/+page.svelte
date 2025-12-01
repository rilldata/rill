<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import ExploreChat from "@rilldata/web-common/features/chat/ExploreChat.svelte";
  import { ToolName } from "@rilldata/web-common/features/chat/core/types";
  import type { PageData } from "./$types";
  import {
    DashboardBannerID,
    DashboardBannerPriority,
  } from "@rilldata/web-common/components/banner/constants";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let data: PageData;

  $: hasBanner = !!data.dashboard.canvas?.state?.validSpec?.banner;

  $: if (hasBanner) {
    eventBus.emit("add-banner", {
      id: DashboardBannerID,
      priority: DashboardBannerPriority,
      message: {
        type: "default",
        message: data.dashboard.canvas?.state?.validSpec?.banner ?? "",
        iconType: "alert",
      },
    });
  }

  onNavigate(() => {
    if (hasBanner) {
      eventBus.emit("remove-banner", DashboardBannerID);
    }
  });
</script>

<div class="flex h-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <CanvasDashboardEmbed
      resource={data.dashboard}
      canvasName={data.dashboardName}
    />
  </div>
  <ExploreChat />
</div>
