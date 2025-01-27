<script lang="ts">
  import { Interval, DateTime } from "luxon";

  export let interval: Interval<true>;
  export let grain: string;
  export let abbreviation: string | undefined = undefined;

  $: showTime = grain === "TIME_GRAIN_HOUR" || grain === "TIME_GRAIN_MINUTE";

  $: intervalStartsAndEndsOnHour =
    interval.start.minute === 0 && interval.end.minute === 0;

  $: timeFormat = intervalStartsAndEndsOnHour ? "h a" : "h:mm a";

  $: inclusiveInterval = interval.set({
    end: interval.end.minus({ millisecond: 1 }),
  });

  // temp while developinng
  $: displayedInterval = showTime ? interval : inclusiveInterval;

  // temp display while developing
  $: date = displayedInterval.toLocaleString(DateTime.DATETIME_FULL);

  $: time = displayedInterval.toFormat(timeFormat, { separator: "-" });
</script>

<div class="flex gap-x-1 whitespace-nowrap" title="{date} {time}">
  <span class="line-clamp-1 text-left">
    {date}
    {#if showTime}
      ({time})
    {/if}
    {#if abbreviation}
      {abbreviation}
    {/if}
  </span>
</div>
