<script lang="ts">
  import { Interval, DateTime } from "luxon";

  export let interval: Interval<true>;

  $: inclusiveInterval = interval.set({
    end: interval.end.minus({ millisecond: interval.end.hour === 0 ? 1 : 0 }),
  });

  $: date = inclusiveInterval.toLocaleString(DateTime.DATE_MED);

  $: time = inclusiveInterval.toFormat("h a", { separator: "-" });

  // Only show time if either time is not 12AM
  $: showTime = interval.end.hour !== 0 || interval.end.minute !== 0;
</script>

<div class="flex gap-x-1" title="{date} {time}">
  <span class="line-clamp-1 text-left"
    >{date}
    {#if showTime}
      ({time}){/if}</span
  >
</div>
