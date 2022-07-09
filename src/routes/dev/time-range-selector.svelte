<script lang="ts">
  import { tick } from "svelte";
  import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";
  import ExploreMenu from "$lib/components/menu/ExploreMenu.svelte";
  import ExploreMenuItem from "$lib/components/menu/ExploreMenuItem.svelte";
  import Button from "$lib/components/Button.svelte";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import { onClickOutside } from "$lib/util/on-click-outside";

  enum TimeRangeName {
    LastHour = "Last hour",
    Last6Hours = "Last 6 hours",
    LastDay = "Last day",
    Last2Days = "Last 2 days",
    Last5Days = "Last 5 days",
    LastWeek = "Last week",
    Last2Weeks = "Last 2 weeks",
    Last30Days = "Last 30 days",
    Last60Days = "Last 60 days",
    Today = "Today",
    MonthToDate = "Month to date",
    // LastMonth = "Last month",
    // CustomRange = "Custom range",
  }

  interface TimeRange {
    name: TimeRangeName;
    start: Date;
    end: Date;
  }

  const makeTimeRange = (name: TimeRangeName): TimeRange => {
    switch (name) {
      case TimeRangeName.LastHour:
        return {
          name,
          start: new Date(Date.now() - 60 * 60 * 1000),
          end: new Date(),
        };
      case TimeRangeName.Last6Hours:
        return {
          name,
          start: new Date(Date.now() - 6 * 60 * 60 * 1000),
          end: new Date(),
        };
      case TimeRangeName.LastDay:
        return {
          name,
          start: new Date(Date.now() - 24 * 60 * 60 * 1000),
          end: new Date(),
        };
      case TimeRangeName.Last2Days:
        return {
          name,
          start: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000),
          end: new Date(),
        };
      case TimeRangeName.Last5Days:
        return {
          name,
          start: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000),
          end: new Date(),
        };
      case TimeRangeName.LastWeek:
        return {
          name,
          start: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000),
          end: new Date(),
        };
      case TimeRangeName.Last2Weeks:
        return {
          name,
          start: new Date(Date.now() - 14 * 24 * 60 * 60 * 1000),
          end: new Date(),
        };
      case TimeRangeName.Last30Days:
        return {
          name,
          start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
          end: new Date(),
        };
      case TimeRangeName.Last60Days:
        return {
          name,
          start: new Date(Date.now() - 60 * 24 * 60 * 60 * 1000),
          end: new Date(),
        };
      case TimeRangeName.Today:
        return {
          name,
          start: new Date(new Date().setHours(0, 0, 0, 0)),
          end: new Date(),
        };
      case TimeRangeName.MonthToDate:
        return {
          name,
          start: new Date(new Date(new Date().setDate(1)).setHours(0, 0, 0, 0)),
          end: new Date(),
        };
      // case TimeRangeName.LastMonth:
      //   return {
      //     name,
      //     start: new Date(new Date().setMonth(new Date().getMonth() - 1)),
      //     end: new Date(),
      //   };
      //   // const lastMonth = new Date(new Date().setMonth(new Date().getMonth() - 1));
      //   return {
      //     name,
      //     start: new Date(lastMonth.setDate(1)),
      //     end: new Date(lastMonth.setMonth(lastMonth.getMonth() + 1)),
      //   };
      // case TimeRangeName.CustomRange:
      //   return {
      //     name,
      //     start: new Date(),
      //     end: new Date(),
      //   };
      default:
        throw new Error(`Unknown time range name: ${name}`);
    }
  };

  const timeRanges: TimeRange[] = Object.keys(TimeRangeName).map((name) =>
    makeTimeRange(TimeRangeName[name])
  );

  const prettyFormatTimeRange = (timeRange: TimeRange): string => {
    // day is the same
    if (
      timeRange.start.getDate() === timeRange.end.getDate() &&
      timeRange.start.getMonth() === timeRange.end.getMonth() &&
      timeRange.start.getFullYear() === timeRange.end.getFullYear()
    ) {
      return `${timeRange.start.toLocaleDateString(undefined, {
        month: "long",
      })} ${timeRange.start.getDate()}, ${timeRange.start.getFullYear()} (${timeRange.start.toLocaleString(
        undefined,
        { hour12: true, hour: "numeric", minute: "numeric" }
      )} - ${timeRange.end.toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
      })})`;
    }
    // month is the same
    if (
      timeRange.start.getMonth() === timeRange.end.getMonth() &&
      timeRange.start.getFullYear() === timeRange.end.getFullYear()
    ) {
      return `${timeRange.start.toLocaleDateString(undefined, {
        month: "long",
      })} ${timeRange.start.getDate()}-${timeRange.end.getDate()}, ${timeRange.start.getFullYear()}`;
    }
    // year is the same
    if (timeRange.start.getFullYear() === timeRange.end.getFullYear()) {
      return `${timeRange.start.toLocaleDateString(undefined, {
        month: "long",
        day: "numeric",
      })} - ${timeRange.end.toLocaleDateString(undefined, {
        month: "long",
        day: "numeric",
      })}, ${timeRange.start.getFullYear()}`;
    }
    // year is different
    const dateFormatOptions: Intl.DateTimeFormatOptions = {
      year: "numeric",
      month: "long",
      day: "numeric",
    };
    return `${timeRange.start.toLocaleDateString(
      undefined,
      dateFormatOptions
    )} - ${timeRange.end.toLocaleDateString(undefined, dateFormatOptions)}`;
  };

  let timeSelectorMenu;
  let timeSelectorMenuOpen = false;
  let menuX;
  let menuY;
  let clickOutsideListener;
  $: if (!timeSelectorMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
  }
  let target: HTMLElement;
</script>

<!-- Timerange menu -->
<Button
  bind:element={target}
  on:click={async (event) => {
    timeSelectorMenuOpen = !timeSelectorMenuOpen;
    menuX = event.clientX;
    menuY = event.clientY;
    if (!clickOutsideListener) {
      await tick();
      clickOutsideListener = onClickOutside(() => {
        timeSelectorMenuOpen = false;
      }, timeSelectorMenu);
    }
  }}
>
  open menu...
  <CaretDownIcon size="16px" />
</Button>

{#if timeSelectorMenuOpen}
  <!-- {#if true} -->
  <div bind:this={timeSelectorMenu}>
    <FloatingElement
      relationship="direct"
      location="bottom"
      alignment="start"
      {target}
      distance={8}
    >
      <ExploreMenu>
        {#each timeRanges as timeRange}
          <ExploreMenuItem on:click={() => console.log(timeRange.name)}>
            <div>
              <span class="font-bold">
                {timeRange.name}
              </span>
              <span />
            </div>
            <div slot="right" let:hovered>
              <span class:opacity-0={!hovered}>
                {prettyFormatTimeRange(timeRange)}
              </span>
            </div>
            <!-- <div class="text-base flex gap-x-4">
              <span class="font-bold">
                {timeRange.name}
              </span>
              <span>
                {prettyFormatTimeRange(timeRange)}
              </span>
            </div> -->
          </ExploreMenuItem>
        {/each}
      </ExploreMenu>
    </FloatingElement>
  </div>
{/if}
