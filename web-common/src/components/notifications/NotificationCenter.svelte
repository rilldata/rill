<script lang="ts">
  import Notification from "./Notification.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import type { NotificationMessage } from "@rilldata/web-common/lib/event-bus/events";
  import { onMount } from "svelte";

  let notification: NotificationMessage | null = null;
  let currentTimeoutId: number | null = null;

  onMount(() => {
    const unsubscribe = eventBus.on("notification", (newNotification) => {
      // Clear any existing timeout
      if (currentTimeoutId) {
        clearTimeout(currentTimeoutId);
        currentTimeoutId = null;
      }

      notification = newNotification;

      // Set up auto-dismiss for non-persisted notifications
      if (
        !newNotification.options?.persisted &&
        newNotification.type !== "loading"
      ) {
        const timeout = newNotification.options?.timeout ?? 3500;
        currentTimeoutId = window.setTimeout(clear, timeout);
      }
    });

    return () => {
      unsubscribe();
      if (currentTimeoutId) {
        clearTimeout(currentTimeoutId);
      }
    };
  });

  function clear() {
    notification = null;
    if (currentTimeoutId) {
      clearTimeout(currentTimeoutId);
      currentTimeoutId = null;
    }
  }
</script>

{#if notification}
  <Notification {notification} onClose={clear} />
{/if}
