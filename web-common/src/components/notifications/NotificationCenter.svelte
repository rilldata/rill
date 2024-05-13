<script lang="ts">
  import Notification from "./Notification.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { NotificationMessage } from "@rilldata/web-common/lib/event-bus/events";
  import { onMount } from "svelte";

  let notification: NotificationMessage | null = null;

  onMount(() => {
    const unsubscribe = eventBus.on("notification", (newNotification) => {
      notification = newNotification;
    });

    return unsubscribe;
  });

  function clear() {
    notification = null;
  }
</script>

{#if notification}
  <Notification {notification} onClose={clear} />
{/if}
