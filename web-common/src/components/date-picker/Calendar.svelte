<script lang="ts">
  import Month from "./Month.svelte";
  import { DateTime, Settings } from "luxon";
  import { Interval } from "luxon";

  type MaybeDate = DateTime | undefined;

  export let selection: Interval | DateTime;
  export let maxDate: MaybeDate = DateTime.now().startOf("day");
  export let minDate: MaybeDate = undefined;
  export let visibleMonths = 1;
  export let anchorDay: DateTime<true> | undefined = undefined;
  export let firstVisibleMonth: MaybeDate = isValidDateTime(selection)
    ? selection
    : isValidInterval(selection)
      ? (selection.start ?? DateTime.now())
      : DateTime.now();
  export let singleDaySelection = false;
  export let onSelectDay: (date: DateTime<true>) => void;
  export let onHoverDay: (date: DateTime<true>) => void = () => {};

  // This is set globally in DashboardStateDataLoader or canvas/stores/time-control.ts
  $: firstDayOfWeek = Settings.defaultWeekSettings?.firstDay ?? 1;

  $: firstMonth = isValidDateTime(firstVisibleMonth)
    ? firstVisibleMonth
    : DateTime.now();

  $: interval = isValidInterval(selection)
    ? selection
    : isValidDateTime(selection)
      ? (Interval.fromDateTimes(
          selection.startOf("day"),
          selection.startOf("day").plus({ day: 1 }),
        ) as Interval<true>)
      : (Interval.fromDateTimes(
          DateTime.now().startOf("day"),
          DateTime.now().startOf("day").plus({ day: 1 }),
        ) as Interval<true>);

  function onPan(direction: -1 | 1) {
    firstMonth = firstMonth.plus({ month: direction });
  }

  function isValidDateTime(
    value: MaybeDate | Interval<true>,
  ): value is DateTime<true> {
    return Boolean(value && value instanceof DateTime && value?.isValid);
  }

  function isValidInterval(
    value: MaybeDate | Interval,
  ): value is Interval<true> {
    return Boolean(value && value instanceof Interval && value?.isValid);
  }
</script>

<div class="flex gap-x-3 w-full">
  {#each { length: visibleMonths } as month, i (month)}
    <Month
      {firstDayOfWeek}
      {maxDate}
      {minDate}
      {anchorDay}
      {singleDaySelection}
      {interval}
      startDay={firstMonth.plus({ month: i }).set({ day: 1 }).startOf("day")}
      {visibleMonths}
      visibleIndex={i}
      {onSelectDay}
      {onPan}
      {onHoverDay}
    />
  {/each}
</div>
