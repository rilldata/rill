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

  $: x = point[xAccessor];

  $: yValues = dimensionData.map((dimension) => {
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
      return {
        x,
        y: currentPointIsNull ? lastAvailableCurrentY : y,
        yOverride: currentPointIsNull,
        yOverrideLabel: "no current data",
        yOverrideStyleClass: `fill-gray-600 italic`,
        key: dimension.name,
        label: "",
        pointColorClass: dimension.fillClass,
        valueStyleClass: "font-semibold",
        valueColorClass: "fill-gray-600",
        labelColorClass: dimension.fillClass,
      };
    })
    .filter((d) => !d.yOverride);

  /** get the final point set*/
  $: pointSet = points;
</script>

{#if pointSet.length}
  <WithGraphicContexts let:xScale let:yScale>
    <MultiMetricMouseoverLabel
      attachPointToLabel
      direction="right"
      flipAtEdge="body"
      formatValue={mouseoverFormat}
      point={pointSet || []}
    />
  </WithGraphicContexts>
{/if}
