<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { timeAgo } from "../../dashboards/listing/utils";

  export let datetime: Date | string;

  let formattedTime: string;
  let interval: NodeJS.Timeout | null = null;
  let lastInterval: number | null = null;

  const INTERVALS = {
    SECOND: 1000,
    MINUTE: 60 * 1000,
    FIVE_MINUTES: 5 * 60 * 1000,
    HOUR: 60 * 60 * 1000,
  };

  function getInterval(date: Date): number {
    const diff = Date.now() - date.getTime();
    if (diff < 60 * 1000) return INTERVALS.SECOND;
    if (diff < 60 * 60 * 1000) return INTERVALS.MINUTE;
    if (diff < 24 * 60 * 60 * 1000) return INTERVALS.FIVE_MINUTES;
    return INTERVALS.HOUR;
  }

  function updateTimeAgo() {
    const date = new Date(datetime);
    formattedTime = timeAgo(date);

    const newInterval = getInterval(date);
    if (newInterval !== lastInterval) {
      lastInterval = newInterval;
      resetInterval(newInterval);
    }
  }

  function resetInterval(newInterval: number) {
    if (interval) clearInterval(interval);
    interval = setInterval(updateTimeAgo, newInterval);
  }

  $: updateTimeAgo(); // Reactively update whenever `datetime` changes

  onMount(updateTimeAgo);

  onDestroy(() => {
    if (interval) clearInterval(interval);
  });
</script>

<time datetime={datetime.toString()}>{formattedTime}</time>
