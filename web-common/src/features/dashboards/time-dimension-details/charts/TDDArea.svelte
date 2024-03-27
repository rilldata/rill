<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/charts/render/VegaLiteRenderer.svelte";

  export let totalsData;
  export let dimensionData;

  /*
  Read Data
  Come up with a heuristic to check if data has time fields and nominal fields
  Suggest chart type options based on data
  Limit to Bar, stack bar, line, area and stacked area charts
  Write a builder function to create the spec based on the data
  Add extents for TDD charts
  For rest add Template UI 
  */

  $: console.log(totalsData);

  const spec = {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    description: "Google's stock price over time.",
    data: { name: "table" },
    width: "container",
    height: "250",
    mark: { type: "bar", width: { band: 0.5 } },
    encoding: {
      x: {
        field: "ts",
        type: "temporal",
        timeUnit: "monthdate",
      },
      y: {
        axis: {
          orient: "right",
        },
        field: "total_records",
        type: "quantitative",
      },
      // color: { field: "device_type", type: "nominal", legend: null },
      opacity: {
        condition: { param: "highlight", empty: false, value: 1 },
        value: 0.8,
      },
    },
    params: [
      {
        name: "highlight",
        select: {
          type: "point",
          on: "pointerover",
          nearest: true,
          encodings: ["x"],
        },
      },
    ],
  };
</script>

<VegaLiteRenderer data={{ table: totalsData }} {spec} />
