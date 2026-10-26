<script lang="ts">
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type * as GeoJSON from "geojson";
  import type mapboxgl from "mapbox-gl";
  import type { ExpressionSpecification } from "mapbox-gl";
  import { onDestroy, onMount, untrack } from "svelte";
  import { DEFAULT_MAP_COLOR_RANGE, type MapComponent, type MapSpec } from "..";
  import {
    buildColorExpression,
    computeValueRange,
    resolveColorRange,
    type MapTheme,
  } from "../color-utils";
  import MapLegend from "../MapLegend.svelte";
  import {
    buildTooltipHTML,
    calculateWrappedBounds,
    removeTooltip,
    showTooltip,
  } from "../map-utils";
  import { buildRegionIndex, matchRegions } from "../regions/match";
  import {
    AUTO_REGION_KEY,
    BOUNDARY_SETS,
    isRegionBoundaries,
  } from "../regions/registry";
  import { createMapLayerQuery } from "./query";
  import type { MapLayerBoundsHandler } from "./types";

  type Props = {
    map: mapboxgl.Map;
    /** Changes every time a map style finishes loading; null while a style is loading. */
    styleKey: number | null;
    component: MapComponent;
    metricsViewSpec: V1MetricsViewSpec | undefined;
    theme: MapTheme;
    isThemeModeDark: boolean;
    onBounds: MapLayerBoundsHandler;
  };

  let {
    map,
    styleKey,
    component,
    metricsViewSpec,
    theme,
    isThemeModeDark,
    onBounds,
  }: Props = $props();

  const SOURCE_ID = "regions";
  const FILL_LAYER = "regions-fill";
  const OUTLINE_LAYER = "regions-outline";
  const VALUE_STATE: ExpressionSpecification = ["feature-state", "value"];
  const HAS_VALUE: ExpressionSpecification = [
    "==",
    ["typeof", VALUE_STATE],
    "number",
  ];
  const IS_HOVERED: ExpressionSpecification = [
    "boolean",
    ["feature-state", "hover"],
    false,
  ];
  const MAX_LISTED_UNMATCHED = 10;

  const runtimeClient = useRuntimeClient();
  const { specStore } = untrack(() => component);

  const regions = $derived(
    ($specStore as Partial<MapSpec> | undefined)?.regions,
  );
  const metricsView = $derived($specStore?.metrics_view ?? "");
  const dimension = $derived(regions?.dimension ?? "");
  const boundariesName = $derived(
    isRegionBoundaries(regions?.boundaries) ? regions.boundaries : undefined,
  );
  const boundarySet = $derived(
    boundariesName ? BOUNDARY_SETS[boundariesName] : undefined,
  );
  const regionKey = $derived(regions?.key ?? AUTO_REGION_KEY);
  const colorMeasure = $derived(regions?.color?.measure ?? "");
  const colorRange = $derived(
    regions?.color?.colorRange ?? DEFAULT_MAP_COLOR_RANGE,
  );

  const dataQuery = createMapLayerQuery(
    runtimeClient,
    untrack(() => component),
    (spec) => {
      const r = spec.regions;
      if (
        !r?.dimension ||
        !r.color?.measure ||
        !isRegionBoundaries(r.boundaries)
      )
        return null;
      return {
        dimensions: [{ name: r.dimension }],
        measures: [{ name: r.color.measure }],
      };
    },
  );

  let boundaries = $state.raw<GeoJSON.FeatureCollection | null>(null);
  let loadError = $state<string | null>(null);

  $effect(() => {
    const set = boundarySet;
    boundaries = null;
    loadError = null;
    if (!set) return;

    let cancelled = false;
    set.load().then(
      (fc) => {
        if (!cancelled) boundaries = fc;
      },
      (err: unknown) => {
        if (!cancelled) loadError = `Failed to load ${set.label}: ${err}`;
      },
    );
    return () => {
      cancelled = true;
    };
  });

  const regionIndex = $derived.by(() => {
    if (!boundaries || !boundarySet) return null;
    const keys =
      regionKey === AUTO_REGION_KEY
        ? boundarySet.keys.map((k) => k.name)
        : [regionKey];
    return buildRegionIndex(boundaries.features, keys);
  });

  const rows = $derived($dataQuery.data?.data ?? []);
  const match = $derived(
    regionIndex
      ? matchRegions(rows, dimension, colorMeasure, regionIndex)
      : null,
  );

  const canRender = $derived(
    styleKey !== null &&
      !!boundaries &&
      !!match &&
      !$dataQuery.isPending &&
      !!colorMeasure,
  );

  const colors = $derived(resolveColorRange(colorRange, theme));
  const colorDomain = $derived(
    computeValueRange([...(match?.values.values() ?? [])].map((v) => v.value)),
  );

  const measureSpec = $derived(
    metricsViewSpec?.measures?.find((m) => m.name === colorMeasure),
  );
  const measureLabel = $derived(measureSpec?.displayName || colorMeasure);

  function formatValue(value: number): string {
    return measureSpec
      ? createMeasureValueFormatter(measureSpec)(value)
      : String(value);
  }

  const unmatchedNote = $derived.by(() => {
    const unmatched = match?.unmatched ?? [];
    if (unmatched.length === 0) return undefined;
    return {
      note: `${unmatched.length} ${unmatched.length === 1 ? "value" : "values"} did not match a region`,
      tooltip:
        unmatched.slice(0, MAX_LISTED_UNMATCHED).join(", ") +
        (unmatched.length > MAX_LISTED_UNMATCHED ? ", …" : ""),
    };
  });

  let hoveredId: string | number | undefined;

  function setHovered(id: string | number | undefined) {
    if (hoveredId === id) return;
    if (hoveredId !== undefined && map.getSource(SOURCE_ID)) {
      map.setFeatureState(
        { source: SOURCE_ID, id: hoveredId },
        { hover: false },
      );
    }
    hoveredId = id;
    if (id !== undefined) {
      map.setFeatureState({ source: SOURCE_ID, id }, { hover: true });
    }
  }

  function handleMouseMove(
    e: mapboxgl.MapMouseEvent & { features?: mapboxgl.GeoJSONFeature[] },
  ) {
    const feature = e.features?.[0];
    setHovered(feature?.id);
    if (!feature) return removeTooltip();

    const value = feature.state?.value as number | undefined;
    map.getCanvas().style.cursor = typeof value === "number" ? "pointer" : "";
    showTooltip(
      e,
      buildTooltipHTML(feature.properties?.name as string | undefined, [
        {
          label: measureLabel,
          value: typeof value === "number" ? formatValue(value) : "No data",
        },
      ]),
    );
  }

  function handleMouseLeave() {
    setHovered(undefined);
    map.getCanvas().style.cursor = "";
    removeTooltip();
  }

  function addLayers(data: GeoJSON.FeatureCollection) {
    map.addSource(SOURCE_ID, { type: "geojson", data, promoteId: "id" });

    // Regions go below the base map's labels, and so below the points layer too,
    // so place names and points stay readable on top of the fill.
    const beforeId = map
      .getStyle()
      ?.layers?.find((layer) => layer.type === "symbol")?.id;
    map.addLayer({ id: FILL_LAYER, type: "fill", source: SOURCE_ID }, beforeId);
    map.addLayer(
      { id: OUTLINE_LAYER, type: "line", source: SOURCE_ID },
      beforeId,
    );
  }

  function removeLayers() {
    for (const id of [OUTLINE_LAYER, FILL_LAYER]) {
      if (map.getLayer(id)) map.removeLayer(id);
    }
    if (map.getSource(SOURCE_ID)) map.removeSource(SOURCE_ID);
  }

  // Layer event handlers bind by layer id and survive style reloads,
  // so they are registered once for the lifetime of this layer.
  onMount(() => {
    map.on("mousemove", FILL_LAYER, handleMouseMove);
    map.on("mouseleave", FILL_LAYER, handleMouseLeave);
  });

  onDestroy(() => {
    map.off("mousemove", FILL_LAYER, handleMouseMove);
    map.off("mouseleave", FILL_LAYER, handleMouseLeave);
    // The map itself may already be torn down when the whole component unmounts.
    if (styleKey !== null && map.getStyle()) removeLayers();
  });

  // Switching boundary sets replaces the geometry; feature state from the
  // previous set would otherwise leak onto regions that share an id.
  $effect(() => {
    if (styleKey === null || !boundaries) return;
    const source = map.getSource(SOURCE_ID) as
      | mapboxgl.GeoJSONSource
      | undefined;
    if (source) {
      map.removeFeatureState({ source: SOURCE_ID });
      source.setData(boundaries);
    } else {
      addLayers(boundaries);
    }
  });

  // Reruns on any spec, theme or data change, and after a style reload.
  $effect(() => {
    if (!canRender || !match || !boundaries) return;
    if (!map.getSource(SOURCE_ID)) return;

    map.removeFeatureState({ source: SOURCE_ID });
    for (const [id, { value }] of match.values) {
      map.setFeatureState({ source: SOURCE_ID, id }, { value });
    }
    hoveredId = undefined;

    const noDataColor = isThemeModeDark ? "#3f3f46" : "#e4e4e7";
    const outlineColor = isThemeModeDark ? "#18181b" : "#ffffff";

    map.setPaintProperty(FILL_LAYER, "fill-color", [
      "case",
      HAS_VALUE,
      buildColorExpression(VALUE_STATE, colorDomain[0], colorDomain[1], colors),
      noDataColor,
    ]);
    map.setPaintProperty(FILL_LAYER, "fill-opacity", [
      "case",
      HAS_VALUE,
      ["case", IS_HOVERED, 0.95, 0.75],
      0.5,
    ]);
    map.setPaintProperty(OUTLINE_LAYER, "line-color", [
      "case",
      IS_HOVERED,
      theme.primary.hex(),
      outlineColor,
    ]);
    map.setPaintProperty(OUTLINE_LAYER, "line-width", [
      "case",
      IS_HOVERED,
      1.5,
      0.5,
    ]);

    const matched = boundaries.features.filter((f) =>
      match.values.has(f.properties?.id as string),
    );
    onBounds(
      "regions",
      `${metricsView}::${dimension}::${boundariesName}`,
      calculateWrappedBounds(matched),
    );
  });
</script>

{#if loadError}
  <div class="no-match">{loadError}</div>
{:else if canRender && match && match.values.size > 0}
  <MapLegend
    title={measureLabel}
    {colors}
    min={formatValue(colorDomain[0])}
    max={formatValue(colorDomain[1])}
    note={unmatchedNote?.note}
    noteTooltip={unmatchedNote?.tooltip}
  />
{:else if canRender && unmatchedNote}
  <div class="no-match" title={unmatchedNote.tooltip}>
    {unmatchedNote.note}
  </div>
{/if}

<style lang="postcss">
  .no-match {
    @apply px-2 py-1.5 rounded shadow-sm text-[11px] text-fg-secondary;
    @apply bg-surface-background/90;
  }
</style>
