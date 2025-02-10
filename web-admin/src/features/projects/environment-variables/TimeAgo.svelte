<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { timeAgo } from "../../dashboards/listing/utils";
  import { Duration } from "luxon";

  export let datetime: Date | string;

  let formattedTime: string;
  let interval: ReturnType<typeof setInterval> | null = null;
  let lastInterval: number | null = null;

  const INTERVALS = {
    SECOND: Duration.fromObject({ seconds: 1 }).toMillis(),
    MINUTE: Duration.fromObject({ minutes: 1 }).toMillis(),
    FIVE_MINUTES: Duration.fromObject({ minutes: 5 }).toMillis(),
    HOUR: Duration.fromObject({ hours: 1 }).toMillis(),
  };

  function getInterval(date: Date): number {
    const diff = Duration.fromMillis(Date.now() - date.getTime());
    if (diff.as("minutes") < 1) return INTERVALS.SECOND;
    if (diff.as("hours") < 1) return INTERVALS.MINUTE;
    if (diff.as("days") < 1) return INTERVALS.FIVE_MINUTES;
    return INTERVALS.HOUR;
  }

  function updateTimeAgo(date: Date) {
    formattedTime = timeAgo(date);

    const newInterval = getInterval(date);
    if (newInterval !== lastInterval) {
      lastInterval = newInterval;
      resetInterval(newInterval);
    }
  }

  function resetInterval(newInterval: number) {
    if (interval) clearInterval(interval);
    interval = setInterval(
      () => updateTimeAgo(new Date(datetime)),
      newInterval,
    );
  }

  $: {
    const date = new Date(datetime);
    updateTimeAgo(date);
  }

  onDestroy(() => {
    if (interval) clearInterval(interval);
  });
</script>

<time datetime={datetime.toString()}>{formattedTime}</time>
