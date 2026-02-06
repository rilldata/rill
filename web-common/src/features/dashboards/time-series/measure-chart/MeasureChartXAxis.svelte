<script lang="ts">
  import { scaleLinear } from "d3-scale";
  import { type DateTime, type Interval } from "luxon";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import {
    V1TimeGrainToOrder,
    V1TimeGrainToDateTimeUnit,
  } from "@rilldata/web-common/lib/time/new-grains";
  import {
    LINE_MODE_MIN_POINTS,
    X_PAD,
    MARGIN_RIGHT,
    computeXTickIndices,
  } from "./scales";
  const DAY_GRAIN_ORDER = V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_DAY];

  export let interval: Interval<true> | undefined = undefined;
  export let timeGranularity: V1TimeGrain | undefined = undefined;

  let clientWidth = 0;

  // Compute time bins from interval + grain
  $: bins = computeBins(interval, timeGranularity);

  $: isSubDay = timeGranularity
    ? V1TimeGrainToOrder[timeGranularity] < DAY_GRAIN_ORDER
    : false;
  $: spansYears =
    bins.length > 1 && bins[0].year !== bins[bins.length - 1].year;
  $: needsStacked = isSubDay || spansYears;
  $: axisHeight = needsStacked ? 26 : 16;

  // X scale — mirrors MeasureChart logic
  $: lastIndex = Math.max(0, bins.length - 1);
  $: plotWidth = Math.max(0, clientWidth - MARGIN_RIGHT);
  $: mode =
    bins.length >= LINE_MODE_MIN_POINTS ? ("line" as const) : ("bar" as const);
  $: barSlotWidth = plotWidth / Math.max(1, bins.length);
  $: xRangeStart = mode === "line" ? X_PAD : barSlotWidth / 2;
  $: xRangeEnd =
    mode === "line" ? plotWidth - X_PAD : plotWidth - barSlotWidth / 2;
  $: xScale = scaleLinear()
    .domain([0, lastIndex])
    .range([xRangeStart, xRangeEnd]);

  $: tickIndices = computeXTickIndices(mode, bins.length);

  $: ticks = buildTicks(tickIndices, bins, xScale);

  function computeBins(
    iv: Interval<true> | undefined,
    grain: V1TimeGrain | undefined,
  ): DateTime[] {
    if (!iv || !grain) return [];
    const unit = V1TimeGrainToDateTimeUnit[grain];
    if (!unit) return [];

    // Truncate start to grain boundary (e.g. Wednesday → Monday for week grain)
    const aligned = iv.start.startOf(unit).until(iv.end);
    if (!aligned.isValid) return [];

    return aligned
      .splitBy({ [unit + "s"]: 1 })
      .map((i) => i.start!)
      .filter((dt): dt is DateTime => dt !== null);
  }

  interface Tick {
    x: number;
    anchor: string;
    timeLine: string;
    dateLine: string;
  }

  function buildTicks(
    indices: number[],
    b: DateTime[],
    scale: (i: number) => number,
  ): Tick[] {
    if (b.length === 0) return [];
    return indices.map((idx, i) => {
      const dt = b[idx];
      if (!dt)
        return { x: scale(idx), anchor: "middle", timeLine: "", dateLine: "" };

      const anchor =
        mode === "bar"
          ? "middle"
          : i === 0
            ? "start"
            : i === indices.length - 1
              ? "end"
              : "middle";

      const prevDt = i > 0 ? b[indices[i - 1]] : undefined;
      return { x: scale(idx), anchor, ...formatTick(dt, prevDt) };
    });
  }

  function formatTick(
    dt: DateTime,
    prevDt: DateTime | undefined,
  ): { timeLine: string; dateLine: string } {
    if (!timeGranularity) return { timeLine: "", dateLine: "" };

    if (isSubDay) {
      const grainOrder = V1TimeGrainToOrder[timeGranularity];
      const fmt: Intl.DateTimeFormatOptions = {
        hour: "numeric",
        hour12: true,
      };
      if (grainOrder < 1) fmt.minute = "2-digit";
      const timeLine = dt.toLocaleString(fmt);

      const dateChanged =
        !prevDt ||
        dt.day !== prevDt.day ||
        dt.month !== prevDt.month ||
        dt.year !== prevDt.year;

      if (!dateChanged) return { timeLine, dateLine: "" };

      const dateFmt: Intl.DateTimeFormatOptions = {
        month: "short",
        day: "numeric",
      };
      if (spansYears && (!prevDt || dt.year !== prevDt.year)) {
        dateFmt.year = "numeric";
      }
      return { timeLine, dateLine: dt.toLocaleString(dateFmt) };
    }

    // Year grain — just show the year, no stacking needed
    if (timeGranularity === V1TimeGrain.TIME_GRAIN_YEAR) {
      return { timeLine: dt.toFormat("yyyy"), dateLine: "" };
    }

    // Day, week, month, quarter grains
    let timeLine: string;
    if (timeGranularity === V1TimeGrain.TIME_GRAIN_WEEK) {
      timeLine = `W${dt.weekNumber}`;
    } else if (timeGranularity === V1TimeGrain.TIME_GRAIN_QUARTER) {
      timeLine = `Q${dt.quarter}`;
    } else if (timeGranularity === V1TimeGrain.TIME_GRAIN_MONTH) {
      timeLine = dt.toLocaleString({ month: "short" });
    } else {
      timeLine = dt.toLocaleString({ month: "short", day: "numeric" });
    }

    if (spansYears) {
      const yearChanged = !prevDt || dt.year !== prevDt.year;
      return {
        timeLine,
        dateLine: yearChanged ? dt.toFormat("yyyy") : "",
      };
    }

    return { timeLine, dateLine: "" };
  }
</script>

<div bind:clientWidth class="w-full pb-1">
  {#if bins.length > 0 && clientWidth > 0}
    <svg class="w-full overflow-visible" height={axisHeight}>
      {#each ticks as tick, tickIdx (tickIdx)}
        <text
          class="fill-fg-secondary text-[11px]"
          text-anchor={tick.anchor}
          x={tick.x}
          y={needsStacked ? 11 : axisHeight - 3}
        >
          {tick.timeLine}
        </text>
        {#if tick.dateLine}
          <text
            class="fill-fg-muted text-[11px]"
            text-anchor={tick.anchor}
            x={tick.x}
            y={23}
          >
            {tick.dateLine}
          </text>
        {/if}
      {/each}
    </svg>
  {/if}
</div>
