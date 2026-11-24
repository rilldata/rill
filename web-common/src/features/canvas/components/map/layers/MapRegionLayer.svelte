<script lang="ts">
  import { getDimensionDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
  import { getThemeBoundaryClass } from "@rilldata/web-common/features/themes/theme-boundary";
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
    type MapLegendNote,
    removeTooltip,
    showTooltip,
  } from "../map-utils";
  import {
    createRegionBoundariesQuery,
    isValidRegionBoundaries,
    regionBoundariesKey,
    regionIndexKeys,
  } from "../regions/boundaries";
  import { buildRegionIndex, matchRegions } from "../regions/match";
  import { createMapLayerQuery } from "./query";
  import type { MapLayerBoundsHandler, MapLayerExportHandler } from "./types";

  type Props = {
    map: mapboxgl.Map;
    /** Changes every time a map style finishes loading; null while a style is loading. */
    styleKey: number | null;
    component: MapComponent;
    metricsViewSpec: V1MetricsViewSpec | undefined;
    theme: MapTheme;
    isThemeModeDark: boolean;
    onBounds: MapLayerBoundsHandler;
    onExport: MapLayerExportHandler;
    /** Called whenever the layer hands the map new data or paint to draw. */
    onDraw: () => void;
    /** Names the layer in its legend, to tell it apart from the other layer's. */
    showLayerInLegend: boolean;
  };

  let {
    map,
    styleKey,
    component,
    metricsViewSpec,
    theme,
    isThemeModeDark,
    onBounds,
    onExport,
    onDraw,
    showLayerInLegend,
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
  // Set on regions that several dimension values match, such as "US" and "USA".
  // Their values can't be combined into one color, so they get a color outside the scale.
  const IS_AMBIGUOUS: ExpressionSpecification = [
    "boolean",
    ["feature-state", "ambiguous"],
    false,
  ];
  const IS_HOVERED: ExpressionSpecification = [
    "boolean",
    ["feature-state", "hover"],
    false,
  ];
  const MAX_LISTED_VALUES = 10;

  const runtimeClient = useRuntimeClient();
  const themeBoundaryClass = getThemeBoundaryClass();
  const { viewSpecStore: specStore } = untrack(() => component);

  const regions = $derived(
    ($specStore as Partial<MapSpec> | undefined)?.regions,
  );
  const metricsView = $derived($specStore?.metrics_view ?? "");
  const dimension = $derived(regions?.dimension ?? "");
  const country = $derived(regions?.country);
  const boundariesConfig = $derived(
    isValidRegionBoundaries(regions?.boundaries, country)
      ? regions.boundaries
      : undefined,
  );
  // Primitives, so that unrelated spec updates don't rebuild the region index.
  const boundariesKey = $derived(
    boundariesConfig
      ? regionBoundariesKey(boundariesConfig, country)
      : undefined,
  );
  const indexKeys = $derived(
    boundariesConfig
      ? regionIndexKeys(boundariesConfig, regions?.key).join(",")
      : undefined,
  );
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
        !isValidRegionBoundaries(r.boundaries, r.country)
      )
        return null;
      return {
        dimensions: [{ name: r.dimension }],
        measures: [{ name: r.color.measure }],
      };
    },
  );

  const boundariesQuery = createRegionBoundariesQuery(
    runtimeClient,
    untrack(() => component),
  );
  const boundaries = $derived(
    boundariesConfig ? ($boundariesQuery.data ?? null) : null,
  );
  const loadError = $derived(
    boundariesConfig ? ($boundariesQuery.error?.message ?? null) : null,
  );

  const regionIndex = $derived.by(() => {
    if (!boundaries || !indexKeys) return null;
    return buildRegionIndex(boundaries.features, indexKeys.split(","));
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
    computeValueRange(
      [...(match?.values.values() ?? [])]
        .filter((entries) => entries.length === 1)
        .map(([entry]) => entry.value),
    ),
  );
  const ambiguousColor = $derived(isThemeModeDark ? "#78716c" : "#a8a29e");

  const measureSpec = $derived(
    metricsViewSpec?.measures?.find((m) => m.name === colorMeasure),
  );
  const measureLabel = $derived(measureSpec?.displayName || colorMeasure);

  const formatValue = $derived(
    measureSpec ? createMeasureValueFormatter(measureSpec) : String,
  );
  const dimensionLabel = $derived(
    getDimensionDisplayName(
      metricsViewSpec?.dimensions?.find((d) => d.name === dimension),
    ) || dimension,
  );

  const unmatchedNote = $derived.by((): MapLegendNote | undefined => {
    const unmatched = match?.unmatched ?? [];
    if (unmatched.length === 0) return undefined;
    return {
      text: `${unmatched.length} ${unmatched.length === 1 ? "value" : "values"} did not match a region`,
      tooltip: listValues(unmatched),
    };
  });

  const ambiguousNote = $derived.by((): MapLegendNote | undefined => {
    if (!match || !boundaries) return undefined;
    const ambiguous = boundaries.features.flatMap((f) => {
      const entries = match.values.get(f.properties?.id as string);
      if (!entries || entries.length < 2) return [];
      const name = (f.properties?.name as string | undefined) ?? "";
      return [`${name}: ${entries.map((e) => e.label).join(", ")}`];
    });
    if (ambiguous.length === 0) return undefined;
    return {
      text: `${ambiguous.length} ${ambiguous.length === 1 ? "region matches" : "regions match"} several values`,
      tooltip: `${listValues(ambiguous, "; ")}\nThese regions aren't colored by value. Normalize the dimension so that each region has a single value.`,
      color: ambiguousColor,
    };
  });

  function listValues(values: string[], separator = ", ") {
    return (
      values.slice(0, MAX_LISTED_VALUES).join(separator) +
      (values.length > MAX_LISTED_VALUES ? `${separator}…` : "")
    );
  }

  $effect(() => {
    if ($dataQuery.isPending || !colorMeasure) {
      onExport("regions", undefined);
      return;
    }
    onExport("regions", {
      name: "regions",
      rows,
      dimensions: [
        metricsViewSpec?.dimensions?.find((d) => d.name === dimension) ?? {
          name: dimension,
        },
      ],
      measures: [measureSpec ?? { name: colorMeasure }],
    });
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

    const entries = match?.values.get(feature.id as string) ?? [];
    map.getCanvas().style.cursor = entries.length > 0 ? "pointer" : "";
    showTooltip(
      e,
      buildTooltipHTML(
        feature.properties?.name as string | undefined,
        entries.length === 0
          ? [{ label: measureLabel, value: "No data" }]
          : entries.map((entry) => ({
              // Names the dimension value each value comes from when the region has several.
              label:
                entries.length === 1
                  ? measureLabel
                  : `${measureLabel} (${entry.label})`,
              value: formatValue(entry.value),
            })),
      ),
      themeBoundaryClass,
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
  // Without boundaries, such as while they load or after they fail to load,
  // the previous regions are removed so they don't pass for the current config.
  $effect(() => {
    if (styleKey === null) return;
    if (!boundaries) {
      // The layer's mouseleave won't fire once it is removed.
      if (hoveredId !== undefined) handleMouseLeave();
      removeLayers();
      return;
    }
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

    onDraw();
    map.removeFeatureState({ source: SOURCE_ID });
    for (const [id, entries] of match.values) {
      map.setFeatureState(
        { source: SOURCE_ID, id },
        entries.length === 1
          ? { value: entries[0].value }
          : { ambiguous: true },
      );
    }
    hoveredId = undefined;

    const noDataColor = isThemeModeDark ? "#3f3f46" : "#e4e4e7";
    const outlineColor = isThemeModeDark ? "#18181b" : "#ffffff";

    map.setPaintProperty(FILL_LAYER, "fill-color", [
      "case",
      HAS_VALUE,
      buildColorExpression(VALUE_STATE, colorDomain[0], colorDomain[1], colors),
      IS_AMBIGUOUS,
      ambiguousColor,
      noDataColor,
    ]);
    map.setPaintProperty(FILL_LAYER, "fill-opacity", [
      "case",
      ["any", HAS_VALUE, IS_AMBIGUOUS],
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
      `${metricsView}::${dimension}::${boundariesKey}`,
      calculateWrappedBounds(matched),
    );
  });
</script>

{#if loadError}
  <div class="load-error" title={loadError}>{loadError}</div>
{:else if canRender && match && match.values.size > 0}
  <MapLegend
    title={measureLabel}
    layer={showLayerInLegend ? dimensionLabel : undefined}
    {colors}
    min={formatValue(colorDomain[0])}
    max={formatValue(colorDomain[1])}
    notes={[unmatchedNote, ambiguousNote].filter(
      (note): note is MapLegendNote => !!note,
    )}
  />
{:else if canRender && unmatchedNote}
  <div class="no-match" title={unmatchedNote.tooltip}>
    {unmatchedNote.text}
  </div>
{/if}

<style lang="postcss">
  .no-match {
    @apply px-2 py-1.5 rounded shadow-sm text-[11px] text-fg-secondary;
    @apply bg-surface-background/90;
  }

  .load-error {
    @apply max-w-80 px-2 py-1.5 rounded shadow-sm text-xs text-fg-primary;
    @apply bg-destructive/30 break-words;
  }
</style>
