<script lang="ts">
  import { Interval, DateTime } from "luxon";

  export let interval: Interval<true>;

  $: date = interval.toLocaleString(DateTime.DATE_MED);

  $: time = interval.toFormat("h a", { separator: "-" });
</script>

<div class="flex gap-x-1" title="{date} {time}">
  <span class="line-clamp-1 text-left">{date} ({time})</span>
</div>
