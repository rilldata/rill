<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/charts/render/VegaLiteRenderer.svelte";
  import { buildVegaLiteSpec } from "@rilldata/web-common/features/charts/templates/build-template";
  import { reduceDimensionData, sanitizeSpecForTDD } from "./utils";

  export let totalsData;
  export let dimensionData;
  export let expandedMeasureName: string;
  export let chartType: string;

  let vegaSpec;
  let data = totalsData;

  $: if (chartType === "bar") {
    vegaSpec = buildVegaLiteSpec("bar", ["ts"], [expandedMeasureName]);
  } else if (chartType === "stacked bar") {
    data = dimensionData.length
      ? reduceDimensionData(dimensionData)
      : totalsData;
    vegaSpec = buildVegaLiteSpec(
      "bar",
      ["ts"],
      [expandedMeasureName],
      ["dimension"],
    );
  } else if (chartType === "stacked area") {
    data = dimensionData.length
      ? reduceDimensionData(dimensionData)
      : totalsData;
    vegaSpec = buildVegaLiteSpec(
      "stacked area",
      ["ts"],
      [expandedMeasureName],
      ["dimension"],
    );
  }

  $: sanitizedVegaSpec = sanitizeSpecForTDD(vegaSpec);
</script>

{#if sanitizedVegaSpec}
  <VegaLiteRenderer data={{ table: data }} spec={sanitizedVegaSpec} />
{/if}
