<script lang="ts">
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import type {
    V1MetricsViewAggregationMeasure,
    V1MetricsViewSpec,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type * as GeoJSON from "geojson";
  import type mapboxgl from "mapbox-gl";
  import type { ExpressionSpecification } from "mapbox-gl";
  import { onDestroy, onMount, untrack } from "svelte";
  import { DEFAULT_MAP_COLOR_RANGE, type MapComponent, type MapSpec } from "..";
  import {
    buildColorExpression,
    buildSizeExpression,
    computeMinMax,
    resolveColorRange,
    type MapTheme,
  } from "../color-utils";
  import MapLegend from "../MapLegend.svelte";
  import {
    buildTooltipHTML,
    calculateBounds,
    detectPolygonMode,
    removeTooltip,
    showTooltip,
    transformToGeoJSON,
    type TooltipRow,
  } from "../map-utils";
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

  const SOURCE_ID = "map-data";
  const POINTS_LAYER = "points";
  const FILL_LAYER = "polygons-fill";
  const OUTLINE_LAYER = "polygons-outline";

  const runtimeClient = useRuntimeClient();
  const { specStore } = untrack(() => component);

  const mapSpec = $derived($specStore ?? ({} as Partial<MapSpec>));
  const metricsView = $derived(mapSpec.metrics_view ?? "");
  const geoDimension = $derived(mapSpec.geo_dimension?.field ?? "");
  const tooltipDimension = $derived(mapSpec.tooltip_dimension?.field);
  const colorMeasure = $derived(mapSpec.color?.measure ?? "");
  const colorRange = $derived(
    mapSpec.color?.colorRange ?? DEFAULT_MAP_COLOR_RANGE,
  );
  const sizeMeasure = $derived(mapSpec.size_measure?.field);

  const dataQuery = createMapLayerQuery(
    runtimeClient,
    untrack(() => component),
    (spec) => {
      const gd = spec.geo_dimension?.field;
      const cm = spec.color?.measure;
      if (!gd || !cm) return null;

      const measures: V1MetricsViewAggregationMeasure[] = [{ name: cm }];
      if (spec.size_measure?.field) {
        measures.push({ name: spec.size_measure.field });
      }
      return {
        dimensions: [
          { name: gd },
          ...(spec.tooltip_dimension?.field
            ? [{ name: spec.tooltip_dimension.field }]
            : []),
        ],
        measures,
      };
    },
  );

  const rows = $derived($dataQuery.data?.data ?? []);
  // `isPending` stays true while the query is disabled (incomplete spec), so
  // nothing renders until the first result arrives. The color measure is
  // checked separately because a spec edit can leave placeholder data from the
  // previous query in place.
  const canRender = $derived(
    styleKey !== null && !$dataQuery.isPending && !!colorMeasure,
  );

  const geoJson = $derived(
    transformToGeoJSON(rows, {
      geoDimension,
      colorMeasure,
      sizeMeasure,
      tooltipDimension,
    }),
  );
  const colors = $derived(resolveColorRange(colorRange, theme));
  const colorDomain = $derived(computeMinMax(geoJson.features, colorMeasure));

  function getMeasure(name: string | undefined) {
    return metricsViewSpec?.measures?.find((m) => m.name === name);
  }

  function formatMeasure(name: string, value: number): string {
    const measure = getMeasure(name);
    return measure
      ? createMeasureValueFormatter(measure)(value)
      : String(value);
  }

  function getMeasureDisplayName(name: string): string {
    return getMeasure(name)?.displayName || name;
  }

  function handleMouseMove(
    e: mapboxgl.MapMouseEvent & { features?: GeoJSON.Feature[] },
  ) {
    map.getCanvas().style.cursor = "pointer";
    const properties = e.features?.[0]?.properties;
    if (!properties) return removeTooltip();

    const title =
      tooltipDimension && properties[tooltipDimension] != null
        ? String(properties[tooltipDimension])
        : undefined;
    const tooltipRows: TooltipRow[] = [];
    for (const measure of new Set([colorMeasure, sizeMeasure])) {
      if (!measure || typeof properties[measure] !== "number") continue;
      tooltipRows.push({
        label: getMeasureDisplayName(measure),
        value: formatMeasure(measure, properties[measure]),
      });
    }
    showTooltip(e, buildTooltipHTML(title, tooltipRows));
  }

  function handleMouseLeave() {
    map.getCanvas().style.cursor = "";
    removeTooltip();
  }

  function addLayers() {
    map.addSource(SOURCE_ID, { type: "geojson", data: geoJson });
    map.addLayer({
      id: FILL_LAYER,
      type: "fill",
      source: SOURCE_ID,
      filter: ["==", ["geometry-type"], "Polygon"],
    });
    map.addLayer({
      id: OUTLINE_LAYER,
      type: "line",
      source: SOURCE_ID,
      filter: ["==", ["geometry-type"], "Polygon"],
    });
    map.addLayer({
      id: POINTS_LAYER,
      type: "circle",
      source: SOURCE_ID,
      filter: ["==", ["geometry-type"], "Point"],
    });
  }

  function removeLayers() {
    for (const id of [POINTS_LAYER, OUTLINE_LAYER, FILL_LAYER]) {
      if (map.getLayer(id)) map.removeLayer(id);
    }
    if (map.getSource(SOURCE_ID)) map.removeSource(SOURCE_ID);
  }

  // Layer event handlers bind by layer id and survive style reloads,
  // so they are registered once for the lifetime of this layer.
  onMount(() => {
    for (const id of [POINTS_LAYER, FILL_LAYER]) {
      map.on("mousemove", id, handleMouseMove);
      map.on("mouseleave", id, handleMouseLeave);
    }
  });

  onDestroy(() => {
    for (const id of [POINTS_LAYER, FILL_LAYER]) {
      map.off("mousemove", id, handleMouseMove);
      map.off("mouseleave", id, handleMouseLeave);
    }
    // The map itself may already be torn down when the whole component unmounts.
    if (styleKey !== null && map.getStyle()) removeLayers();
  });

  // Reruns on any spec, theme or data change, and after a style reload,
  // which drops all sources and layers.
  $effect(() => {
    if (!canRender) return;

    const source = map.getSource(SOURCE_ID) as
      | mapboxgl.GeoJSONSource
      | undefined;
    if (source) {
      source.setData(geoJson);
    } else {
      addLayers();
    }

    const colorPaint = buildColorExpression(
      ["coalesce", ["get", colorMeasure], colorDomain[0]],
      colorDomain[0],
      colorDomain[1],
      colors,
    );
    let radiusPaint: number | ExpressionSpecification = 6;
    if (sizeMeasure) {
      const [min, max] = computeMinMax(geoJson.features, sizeMeasure);
      radiusPaint = buildSizeExpression(sizeMeasure, min, max);
    }

    // Contrast ring around points: white on the light base map, dark on the
    // dark base map, so the stroke stays visible in either theme mode.
    const strokeColor = isThemeModeDark ? "#1a1a1a" : "#fff";

    map.setPaintProperty(POINTS_LAYER, "circle-color", colorPaint);
    map.setPaintProperty(POINTS_LAYER, "circle-radius", radiusPaint);
    map.setPaintProperty(POINTS_LAYER, "circle-opacity", 0.8);
    map.setPaintProperty(POINTS_LAYER, "circle-stroke-width", 1);
    map.setPaintProperty(POINTS_LAYER, "circle-stroke-color", strokeColor);

    map.setPaintProperty(FILL_LAYER, "fill-color", colorPaint);
    map.setPaintProperty(FILL_LAYER, "fill-opacity", 0.4);

    map.setPaintProperty(OUTLINE_LAYER, "line-color", theme.primary.hex());
    map.setPaintProperty(OUTLINE_LAYER, "line-width", 2);

    onBounds(
      "points",
      `${metricsView}::${geoDimension}`,
      calculateBounds(geoJson.features),
    );
  });

  $effect(() => {
    const isPolygonMode = detectPolygonMode(rows, geoDimension);
    if (component._isPolygonMode === isPolygonMode) return;
    component._isPolygonMode = isPolygonMode;
    component.specStore.update((s) => ({ ...s }));
  });
</script>

{#if canRender && geoJson.features.length > 0}
  <MapLegend
    title={getMeasureDisplayName(colorMeasure)}
    {colors}
    min={formatMeasure(colorMeasure, colorDomain[0])}
    max={formatMeasure(colorMeasure, colorDomain[1])}
  />
{/if}
