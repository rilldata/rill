<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
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
  import { isCustomRegionBoundaries } from "./regions/boundaries";
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

  const CUSTOM_BOUNDARIES = "custom";

  const boundariesOptions = [
    ...Object.entries(BOUNDARY_SETS).map(([value, set]) => ({
      value,
      label: set.label,
    })),
    { value: CUSTOM_BOUNDARIES, label: "Custom file" },
  ];

  const customBoundaries = $derived(
    isCustomRegionBoundaries(regions?.boundaries)
      ? regions.boundaries
      : undefined,
  );

  // Custom boundaries are only written to the spec once both `file` and `id` are set,
  // so the inputs keep their own state while they are being filled in.
  let customSelected = $state(false);
  let customFile = $state("");
  let customId = $state("");
  let customName = $state("");
  $effect(() => {
    if (!customBoundaries) return;
    customSelected = true;
    customFile = customBoundaries.file;
    customId = customBoundaries.id;
    customName = customBoundaries.name ?? "";
  });
  // Switching to a built-in set, e.g. by editing the YAML, leaves custom mode.
  const builtinBoundaries = $derived(
    isRegionBoundaries(regions?.boundaries) ? regions.boundaries : undefined,
  );
  $effect(() => {
    if (builtinBoundaries) customSelected = false;
  });
  const isCustom = $derived(!!customBoundaries || customSelected);

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
    if (boundaries === CUSTOM_BOUNDARIES) {
      customSelected = true;
      commitCustomBoundaries();
      return;
    }
    customSelected = false;
    // Keys differ between boundary sets, so go back to automatic matching.
    onChange({
      dimension: regions.dimension,
      boundaries: boundaries as RegionBoundaries,
      color: regions.color,
    });
  }

  function commitCustomBoundaries() {
    if (!regions) return;
    const file = customFile.trim();
    const id = customId.trim();
    const name = customName.trim();
    if (!file || !id) return;
    if (
      customBoundaries?.file === file &&
      customBoundaries.id === id &&
      (customBoundaries.name ?? "") === name
    )
      return;
    // `key` does not apply to custom boundaries, which always match on `id`.
    onChange({
      dimension: regions.dimension,
      boundaries: name ? { file, id, name } : { file, id },
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
      value={isCustom ? CUSTOM_BOUNDARIES : (regions.boundaries as string)}
      full
      size="sm"
      sameWidth
      fontSize={12}
      onChange={handleBoundariesChange}
    />

    {#if isCustom}
      <Input
        id="map-regions-file"
        label="File"
        hint="GeoJSON or TopoJSON file under a public path, such as public/regions.topojson"
        placeholder="public/regions.topojson"
        capitalizeLabel={false}
        textClass="text-sm"
        size="sm"
        labelGap={2}
        bind:value={customFile}
        onBlur={commitCustomBoundaries}
        onEnter={commitCustomBoundaries}
      />
      <Input
        id="map-regions-id"
        label="Match values by property"
        hint="Feature property matched against the dimension's values"
        placeholder="id"
        capitalizeLabel={false}
        textClass="text-sm"
        size="sm"
        labelGap={2}
        bind:value={customId}
        onBlur={commitCustomBoundaries}
        onEnter={commitCustomBoundaries}
      />
      <Input
        id="map-regions-name"
        label="Name property"
        hint="Feature property shown as the region name; defaults to the matched property"
        placeholder="name"
        optional
        capitalizeLabel={false}
        textClass="text-sm"
        size="sm"
        labelGap={2}
        bind:value={customName}
        onBlur={commitCustomBoundaries}
        onEnter={commitCustomBoundaries}
      />
    {:else}
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
    {/if}

    <MapColorSelector
      {canvasName}
      id="map-regions-color-measure"
      {metricsView}
      colorConfig={regions.color}
      onChange={(color) => onChange({ ...regions, color })}
    />
  {/if}
</div>
