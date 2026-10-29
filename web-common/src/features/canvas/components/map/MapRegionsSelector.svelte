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
  import {
    isCustomRegionBoundaries,
    loadStatesCountries,
  } from "./regions/boundaries";
  import {
    AUTO_REGION_KEY,
    BOUNDARY_SETS,
    DEFAULT_STATES_COUNTRY,
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

  // `us_states` is only written by hand; the editor shows it as `states` in the US.
  const boundariesOptions = [
    ...(["countries", "states"] as const).map((value) => ({
      value,
      label: BOUNDARY_SETS[value].label,
    })),
    { value: CUSTOM_BOUNDARIES, label: "Custom file" },
  ];

  const countryNames = new Intl.DisplayNames(undefined, { type: "region" });
  let countryOptions = $state<{ value: string; label: string }[]>([]);
  let countryOptionsLoading = $state(true);
  $effect(() => {
    loadStatesCountries()
      .then((countries) => {
        countryOptions = countries
          .map((value) => ({ value, label: countryNames.of(value) ?? value }))
          .sort((a, b) => a.label.localeCompare(b.label));
      })
      .catch(() => {
        // The region layer shows the error when it loads the boundaries.
      })
      .finally(() => (countryOptionsLoading = false));
  });

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

  const isStates = $derived(
    regions?.boundaries === "states" || regions?.boundaries === "us_states",
  );
  const country = $derived(
    regions?.boundaries === "us_states"
      ? DEFAULT_STATES_COUNTRY
      : regions?.country,
  );

  const keyOptions = $derived([
    { value: AUTO_REGION_KEY, label: "Automatic" },
    ...(isRegionBoundaries(regions?.boundaries)
      ? BOUNDARY_SETS[regions.boundaries].keys
          // FIPS codes only exist for US states.
          .filter((k) => k.name !== "fips" || country === "US")
          .map((k) => ({ value: k.name, label: k.label }))
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
      ...(guessRegionBoundaries(dimension) ?? { boundaries: "countries" }),
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
      ...(boundaries === "states"
        ? { country: country ?? DEFAULT_STATES_COUNTRY }
        : {}),
      color: regions.color,
    });
  }

  function handleCountryChange(newCountry: string) {
    if (!regions) return;
    // Every country has the same keys, so the key is kept; FIPS codes only match US states.
    const { key, ...rest } = regions;
    onChange({
      ...rest,
      boundaries: "states",
      country: newCountry,
      ...(key && (key !== "fips" || newCountry === "US") ? { key } : {}),
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
    const updated: MapRegionsConfig = { ...regions, key };
    if (key === AUTO_REGION_KEY) delete updated.key;
    onChange(updated);
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
      value={isCustom
        ? CUSTOM_BOUNDARIES
        : isStates
          ? "states"
          : (regions.boundaries as string)}
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
      {#if isStates}
        <Select
          id="map-regions-country"
          label="Country"
          options={countryOptions}
          optionsLoading={countryOptionsLoading}
          value={country ?? ""}
          placeholder="Select a country"
          enableSearch
          full
          size="sm"
          sameWidth
          fontSize={12}
          onChange={handleCountryChange}
        />
      {/if}
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
