<script lang="ts">
  import { onDestroy, onMount } from "svelte";

  /** Session timeout duration in milliseconds (1 hour) */
  export let timeoutMs = 60 * 60 * 1000;
  /** When to start showing the warning (10 minutes before timeout) */
  export let warningMs = 10 * 60 * 1000;
  /** Session start time */
  export let sessionStartedAt: string | undefined;

  let showWarning = false;
  let minutesRemaining = 0;
  let interval: ReturnType<typeof setInterval>;

  onMount(() => {
    interval = setInterval(checkTimeout, 30_000);
    checkTimeout();
  });

  onDestroy(() => {
    if (interval) clearInterval(interval);
  });

  function checkTimeout() {
    if (!sessionStartedAt) return;

    const startTime = new Date(sessionStartedAt).getTime();
    const elapsed = Date.now() - startTime;
    const remaining = timeoutMs - elapsed;

    minutesRemaining = Math.max(0, Math.ceil(remaining / 60_000));
    showWarning = remaining > 0 && remaining <= warningMs;
  }
</script>

{#if showWarning}
  <div class="banner" role="alert">
    <span>
      Your edit session will end in {minutesRemaining} minute{minutesRemaining ===
      1
        ? ""
        : "s"} due to inactivity. Push your changes to keep them.
    </span>
  </div>
{/if}

<style lang="postcss">
  .banner {
    @apply px-4 py-2;
    @apply bg-yellow-50 border-b border-yellow-200;
    @apply text-sm text-yellow-800;
    @apply flex items-center justify-center;
  }
</style>
