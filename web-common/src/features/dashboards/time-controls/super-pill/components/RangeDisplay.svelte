<script lang="ts">
  import { DateTime, Interval } from "luxon";

  const fullTimeFormat = "h:mm:ss:SSS";

  export let interval: Interval;
  export let abbreviation: string | undefined = undefined;

  const shouldShowMilliseconds = false;

  $: shouldShowSeconds =
    interval.start?.second !== 0 || interval.end?.second !== 0;

  $: showShouldMinutes =
    interval.start?.minute !== 0 || interval.end?.minute !== 0;

  $: shouldShowHours = interval.start?.hour !== 0 || interval.end?.hour !== 0;

  $: intervalStartsAndEndsOnHour =
    interval.start?.minute === 0 && interval.end?.minute === 0;

  $: timeFormat =
    getTimeFormat(
      intervalStartsAndEndsOnHour,
      showShouldMinutes,
      shouldShowSeconds,
      shouldShowMilliseconds,
    ) + " a";

  $: showTime =
    shouldShowMilliseconds ||
    shouldShowSeconds ||
    showShouldMinutes ||
    shouldShowHours;

  $: inclusiveInterval = interval.set({
    end: interval.end?.minus({ millisecond: 1 }),
  });

  $: displayedInterval = showTime ? interval : inclusiveInterval;

  $: date = displayedInterval.toLocaleString(DateTime.DATE_MED);

  $: time = displayedInterval.toFormat(timeFormat, { separator: "-" });

  function getTimeFormat(
    hours: boolean,
    minutes: boolean,
    seconds: boolean,
    milliseconds: boolean,
  ) {
    if (milliseconds) {
      return fullTimeFormat;
    } else if (seconds) {
      return fullTimeFormat.replace(/:SSS/, "");
    } else if (minutes) {
      return fullTimeFormat.replace(/:SSS/, "").replace(/:ss/, "");
    } else if (hours) {
      return "h";
    }
  }
</script>

<div class="flex gap-x-1 whitespace-nowrap truncate">
  <span class="line-clamp-1 text-left">
    {#if interval.isValid}
      {date}
      {#if showTime}
        ({time})
      {/if}
      {#if abbreviation}
        {abbreviation}
      {/if}
    {:else}
      Invalid Interval
    {/if}
  </span>
</div>
