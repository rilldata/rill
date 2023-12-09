<script lang="ts">
  import WithGraphicContexts from "@rilldata/web-common/components/data-graphic/functional-components/WithGraphicContexts.svelte";
  import MultiMetricMouseoverLabel from "@rilldata/web-common/components/data-graphic/marks/MultiMetricMouseoverLabel.svelte";
  import { bisectData } from "@rilldata/web-common/components/data-graphic/utils";
  import { mean } from "d3-array";
  export let point;
  export let xAccessor;
  export let yAccessor;
  export let mouseoverFormat;
  export let dimensionData;
  export let dimensionValue;
  export let validPercTotal;

  $: x = point[xAccessor];

  function truncate(str) {
    const truncateLength = 34;

    if (str.length > truncateLength) {
      // Check if last character is space
      if (str[truncateLength - 1] === " ") {
        return str.slice(0, truncateLength - 1) + "...";
      }
      return str.slice(0, truncateLength) + "...";
    }
    return str
};

  let pointsData = dimensionData;
  $: if (dimensionValue !== undefined) {
    const higlighted = dimensionData.filter((d) => d.value === dimensionValue);

    if (higlighted.length) {
      pointsData = higlighted;
    }
  }
  $: yValues = pointsData.map((dimension) => {
    const y = bisectData(x, "center", xAccessor, dimension?.data)[yAccessor];
    return {
      y,
      fillClass: dimension?.fillClass,
      name: dimension?.value,
    };
  });

  let lastAvailableCurrentY = 0;
  $: if (yValues.length) {
    lastAvailableCurrentY = mean(yValues, (d) => d.y);
  }

  $: points = yValues
    .map((dimension) => {
      const y = dimension.y;
      const currentPointIsNull = y === null;
      let value = mouseoverFormat(y);
      if (validPercTotal) {
        const percOfTotal = y / validPercTotal;
         value = mouseoverFormat(y) + ",  " + (percOfTotal * 100).toFixed(2) + "%";
      }
      return {
        x,
        y,
        value,
        yOverride: currentPointIsNull,
        yOverrideLabel: "no current data",
        yOverrideStyleClass: `fill-gray-600 italic`,
        key: dimension.name,
        label: truncate(dimension.name),
        pointColorClass: dimension.fillClass,
        valueStyleClass: "font-bold",
        valueColorClass: "fill-gray-600",
        labelColorClass: "fill-gray-600",
        labelStyleClass: "font-semibold",
      };
    })
    .filter((d) => !d.yOverride);

  /** get the final point set*/
  $: pointSet = points;
</script>

{#if pointSet.length}
  <WithGraphicContexts let:xScale let:yScale>
    <MultiMetricMouseoverLabel
      isDimension={true}
      attachPointToLabel
      direction="left"
      flipAtEdge="body"
      formatValue={mouseoverFormat}
      point={pointSet || []}
    />
  </WithGraphicContexts>
{/if}
