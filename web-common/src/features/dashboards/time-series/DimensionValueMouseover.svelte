<script lang="ts">
  import WithGraphicContexts from "@rilldata/web-common/components/data-graphic/functional-components/WithGraphicContexts.svelte";
  import MultiMetricMouseoverLabel from "@rilldata/web-common/components/data-graphic/marks/MultiMetricMouseoverLabel.svelte";
  import type {
    Point,
    YValue,
  } from "@rilldata/web-common/components/data-graphic/marks/types";
  import { bisectData } from "@rilldata/web-common/components/data-graphic/utils";
  import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
  import type { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import type { TimeSeriesDatum } from "./timeseries-data-store";

  export let point: TimeSeriesDatum;
  export let xAccessor: string;
  export let yAccessor: string;
  export let mouseoverFormat: ReturnType<typeof createMeasureValueFormatter>;
  export let dimensionData: DimensionDataItem[];
  export let dimensionValue: string | null | undefined;
  export let validPercTotal: number | null;
  export let hovered = false;
  export let hasTimeComparison = false;

  $: x = point?.[xAccessor];

  function truncate(str: string) {
    if (!str?.length) return str;

    const truncateLength = 34;

    if (str.length > truncateLength) {
      // Check if last character is space
      if (str[truncateLength - 1] === " ") {
        return str.slice(0, truncateLength - 1) + "...";
      }
      return str.slice(0, truncateLength) + "...";
    }
    return str;
  }

  let pointsData = dimensionData;
  $: if (dimensionValue !== undefined) {
    const higlighted = dimensionData.filter((d) => d.value === dimensionValue);
    pointsData = higlighted.length ? higlighted : dimensionData;
  } else {
    pointsData = dimensionData;
  }

  let yValues: YValue[] = [];
  $: {
    yValues = [];
    pointsData.forEach((dimension) => {
      if (!x) {
        yValues.push({ y: null, color: undefined, name: "" });
        return;
      }

      const { entry: bisected } = bisectData(
        new Date(x),
        "center",
        xAccessor,
        dimension?.data,
      );

      if (bisected !== undefined) {
        const y = bisected[yAccessor];
        yValues.push({
          y,
          color: dimension?.color,
          name: dimension?.value,
        });

        if (hasTimeComparison && pointsData.length === 1) {
          const { entry: bisectedComparison } = bisectData(
            new Date(x),
            "center",
            xAccessor,
            dimension?.data,
          );

          if (bisectedComparison !== undefined) {
            const comparisonY = bisectedComparison[`comparison.${yAccessor}`];
            yValues.push({
              y: comparisonY,
              color: dimension?.color,
              name: `${dimension?.value} (Comparison)`,
              isTimeComparison: true,
            });
          }
        }
      } else {
        yValues.push({ y: null, color: undefined, name: "" });
      }
    });
  }

  $: points = yValues
    .map((dimension) => {
      const currentPointIsNull = dimension.y === null;
      const y = Number(dimension.y);
      let value = mouseoverFormat(
        dimension.y === null || typeof dimension.y === "string"
          ? dimension.y
          : y,
      );

      if (validPercTotal) {
        const percOfTotal = y / validPercTotal;
        value =
          mouseoverFormat(y) + ",  " + (percOfTotal * 100).toFixed(2) + "%";
      }
      const point: Point = {
        x: Number(x),
        y,
        value: value ?? undefined,
        yOverride: currentPointIsNull,
        yOverrideLabel: "no current data",
        yOverrideStyleClass: `fill-gray-600 italic`,
        key: dimension.name === null ? "null" : String(dimension.name),
        label: hovered ? truncate(dimension.name ?? "null") : "",
        pointColor: dimension.color,
        pointOpacity: dimension.isTimeComparison ? 0.6 : 1,
        valueStyleClass: dimension.isTimeComparison
          ? "font-normal"
          : "font-semibold",
        valueColorClass: "fill-gray-600",
        labelColorClass: "fill-gray-600",
        labelStyleClass: "font-semibold",
      };

      return point;
    })
    .filter((d) => !d.yOverride);

  /** get the final point set*/
  $: pointSet = points;
</script>

{#if pointSet.length}
  <WithGraphicContexts>
    <MultiMetricMouseoverLabel
      isDimension={true}
      attachPointToLabel
      direction="left"
      flipAtEdge="body"
      formatValue={mouseoverFormat}
      point={pointSet}
    />
  </WithGraphicContexts>
{/if}
