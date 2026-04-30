<script lang="ts">
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { onDestroy, onMount } from "svelte";

  const BANNER_ID = "edit-session-timeout";
  const BANNER_PRIORITY = 0; // Highest priority; session loss is urgent

  /** Session timeout duration in milliseconds (1 hour) */
  export let timeoutMs = 60 * 60 * 1000;
  /** When to start showing the warning (10 minutes before timeout) */
  export let warningMs = 10 * 60 * 1000;
  /** Session start time */
  export let sessionStartedAt: string | undefined;

  let showing = false;
  let interval: ReturnType<typeof setInterval>;

  onMount(() => {
    interval = setInterval(checkTimeout, 30_000);
    checkTimeout();
  });

  onDestroy(() => {
    if (interval) clearInterval(interval);
    if (showing) eventBus.emit("remove-banner", BANNER_ID);
  });

  function checkTimeout() {
    if (!sessionStartedAt) return;

    const startTime = new Date(sessionStartedAt).getTime();
    const elapsed = Date.now() - startTime;
    const remaining = timeoutMs - elapsed;

    const minutesRemaining = Math.max(0, Math.ceil(remaining / 60_000));
    const shouldShow = remaining > 0 && remaining <= warningMs;

    if (shouldShow) {
      showing = true;
      eventBus.emit("add-banner", {
        id: BANNER_ID,
        priority: BANNER_PRIORITY,
        message: {
          type: "warning",
          iconType: "alert",
          message: `Your edit session will end in ${minutesRemaining} minute${minutesRemaining === 1 ? "" : "s"} due to inactivity. Push your changes to keep them.`,
        },
      });
    } else if (showing) {
      showing = false;
      eventBus.emit("remove-banner", BANNER_ID);
    }
  }
</script>
