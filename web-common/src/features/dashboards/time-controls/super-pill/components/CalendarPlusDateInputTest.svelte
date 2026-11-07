<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import CalendarPlusDateInput from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/CalendarPlusDateInput.svelte";
  import type { DateTime, Duration, Interval, DateTimeUnit } from "luxon";

  /**
   * Test-only wrapper that supplies the Tooltip.Provider that the app layout
   * normally provides. Date inputs render a tooltip when a date is out of range.
   */

  export let interval: Interval<true> | undefined;
  export let minDate: DateTime | undefined = undefined;
  export let maxDate: DateTime | undefined = undefined;
  export let minTimeGrain: DateTimeUnit;
  export let zone: string;
  export let maxQueryTimeRange: Duration | undefined = undefined;
  export let updateRange: (range: string) => void = () => {};
  export let onApply: (interval: Interval<true>) => void;
  export let closeMenu: () => void;
</script>

<Tooltip.Provider>
  <CalendarPlusDateInput
    {interval}
    {minDate}
    {maxDate}
    {minTimeGrain}
    {zone}
    {maxQueryTimeRange}
    {updateRange}
    {onApply}
    {closeMenu}
  />
</Tooltip.Provider>
