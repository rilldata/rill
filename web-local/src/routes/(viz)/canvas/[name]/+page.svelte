<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import type { PageData } from "./$types";
  import {
    DashboardBannerID,
    DashboardBannerPriority,
  } from "@rilldata/web-common/components/banner/constants";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let data: PageData;

  $: ({ instanceId } = $runtime);

  $: ({ canvasName } = data);

  $: ({
    canvasEntity: { _banner },
  } = getCanvasStore(canvasName, instanceId));

  $: banner = $_banner;

  $: hasBanner = !!banner;

  $: if (hasBanner) {
    eventBus.emit("add-banner", {
      id: DashboardBannerID,
      priority: DashboardBannerPriority,
      message: {
        type: "default",
        message: banner ?? "",
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

<CanvasDashboardEmbed {canvasName} />
