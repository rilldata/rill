<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import SingleFieldInput from "@rilldata/web-common/features/canvas/inspector/fields/SingleFieldInput.svelte";
  import { get } from "svelte/store";
  import {
    DEFAULT_MAP_COLOR_RANGE,
    type MapComponent,
    type MapRegionsConfig,
  } from ".";
  import MapColorSelector from "./MapColorSelector.svelte";
  import {
    AUTO_REGION_KEY,
    BOUNDARY_SETS,
    guessRegionBoundaries,
    isRegionBoundaries,
    type RegionBoundaries,
  } from "./regions/registry";

  type Props = {
    component: MapComponent;
    canvasName: string;
    metricsView: string;
    regions: MapRegionsConfig | undefined;
    onChange: (updatedConfig: MapRegionsConfig | undefined) => void;
  };

  let { component, canvasName, metricsView, regions, onChange }: Props =
    $props();

  const metricsViewQuery = $derived(
    component.parent.metricsView.getMetricsViewFromName(metricsView),
  );

  const boundariesOptions = Object.entries(BOUNDARY_SETS).map(
    ([value, set]) => ({ value, label: set.label }),
  );

  const keyOptions = $derived([
    { value: AUTO_REGION_KEY, label: "Automatic" },
    ...(isRegionBoundaries(regions?.boundaries)
      ? BOUNDARY_SETS[regions.boundaries].keys.map((k) => ({
          value: k.name,
          label: k.label,
        }))
      : []),
  ]);

  function handleDimensionSelect(dimension: string) {
    if (regions) {
      onChange({ ...regions, dimension });
      return;
    }
    // Start the layer complete, so the spec stays valid: default the color to
    // the points' measure, or the first measure of the metrics view.
    const measure =
      get(component.specStore)?.color?.measure ||
      $metricsViewQuery?.metricsView?.measures?.[0]?.name ||
      "";
    onChange({
      dimension,
      boundaries: guessRegionBoundaries(dimension) ?? "countries",
      color: { measure, colorRange: DEFAULT_MAP_COLOR_RANGE },
    });
  }

  function handleBoundariesChange(boundaries: string) {
    if (!regions) return;
    // Keys differ between boundary sets, so go back to automatic matching.
    onChange({
      dimension: regions.dimension,
      boundaries: boundaries as RegionBoundaries,
      color: regions.color,
    });
  }

  function handleKeyChange(key: string) {
    if (!regions) return;
    const { dimension, boundaries, color } = regions;
    onChange(
      key === AUTO_REGION_KEY
        ? { dimension, boundaries, color }
        : { dimension, boundaries, key, color },
    );
  }
</script>

<div class="space-y-2">
  <InputLabel small label="Regions" id="map-regions-dimension" />

  <SingleFieldInput
    {canvasName}
    metricName={metricsView}
    id="map-regions-dimension"
    type="dimension"
    selectedItem={regions?.dimension}
    isRemovable
    onSelect={handleDimensionSelect}
    onRemove={() => onChange(undefined)}
  />

  {#if regions}
    <Select
      id="map-regions-boundaries"
      label="Boundaries"
      options={boundariesOptions}
      value={regions.boundaries}
      full
      size="sm"
      sameWidth
      fontSize={12}
      onChange={handleBoundariesChange}
    />

    <Select
      id="map-regions-key"
      label="Match values by"
      options={keyOptions}
      value={regions.key ?? AUTO_REGION_KEY}
      full
      size="sm"
      sameWidth
      fontSize={12}
      onChange={handleKeyChange}
    />

    <MapColorSelector
      {canvasName}
      id="map-regions-color-measure"
      {metricsView}
      colorConfig={regions.color}
      onChange={(color) => onChange({ ...regions, color })}
    />
  {/if}
</div>
