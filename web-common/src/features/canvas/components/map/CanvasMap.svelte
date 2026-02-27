<script lang="ts">
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import { isMapColorConfig, type MapComponent } from ".";
  import { onMount, onDestroy } from "svelte";
  import mapboxgl from "mapbox-gl";
  import "mapbox-gl/dist/mapbox-gl.css";
  import {
    getQueryServiceMetricsViewAggregationQueryOptions,
    type V1MetricsViewAggregationResponse,
    type V1MetricsViewAggregationResponseDataItem,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createQuery } from "@tanstack/svelte-query";
  import { derived } from "svelte/store";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { resolveThemeColors } from "@rilldata/web-common/features/themes/theme-utils";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import {
    resolveStaticColor,
    resolveColorRange,
    computeMinMax,
    buildColorExpression,
    buildSizeExpression,
  } from "./color-utils";
  import {
    transformToGeoJSON,
    calculateBounds,
    detectPolygonMode,
    showTooltip,
    removeTooltip,
  } from "./map-utils";

  export let component: MapComponent;

  let mapContainer: HTMLDivElement;
  let map: mapboxgl.Map | null = null;

  const MAPBOX_TOKEN =
    "pk.eyJ1IjoicmlsbGRhdGEiLCJhIjoiY21nemp4Mnl3MDViaGQzc2J0MzB1NjdvMiJ9.4Q8jXek0-EF4RLA_TF4-oA";
  const MAPBOX_STYLE_LIGHT = "mapbox://styles/mapbox/light-v11";
  const MAPBOX_STYLE_DARK = "mapbox://styles/mapbox/dark-v11";

  let mapReady = false;
  let currentMapStyle: string;

  $: ({ instanceId } = $runtime);

  const {
    specStore,
    parent: {
      name: canvasName,
      metricsView: { getMetricsViewFromName },
    },
  } = component;
  $: mapSpec = $specStore ?? ({} as Partial<import(".").MapSpec>);
  $: title = mapSpec.title;
  $: description = mapSpec.description;
  $: show_description_as_tooltip = mapSpec.show_description_as_tooltip;
  $: metrics_view = mapSpec.metrics_view ?? "";
  $: geo_dimension = mapSpec.geo_dimension ?? "";
  $: color = mapSpec.color;
  $: size_measure = mapSpec.size_measure;
  $: time_filters = mapSpec.time_filters;
  $: dimension_filters = mapSpec.dimension_filters;

  $: filters = {
    time_filters,
    dimension_filters,
  };

  $: isThemeModeDark = $themeControl === "dark";
  $: ({
    canvasEntity: { theme: canvasTheme },
  } = getCanvasStore(canvasName, instanceId));
  $: resolvedTheme = resolveThemeColors($canvasTheme?.spec, isThemeModeDark);

  $: colorMeasure = isMapColorConfig(color) ? color.measure : null;

  $: metricsViewQuery = getMetricsViewFromName(metrics_view);
  $: metricsViewSpec = $metricsViewQuery?.metricsView;

  function getMeasureDisplayName(measureName: string): string {
    const measure = metricsViewSpec?.measures?.find(
      (m) => m.name === measureName,
    );
    return measure?.displayName || measureName;
  }

  $: tooltipCtx = {
    tooltipDimension: mapSpec?.tooltip_dimension,
    colorMeasure,
    sizeMeasure: size_measure,
    getDisplayName: getMeasureDisplayName,
  };

  const queryOptionsStore = derived(
    [runtime, specStore],
    ([runtimeVal, specVal]) => {
      const spec = specVal ?? ({} as Partial<import(".").MapSpec>);
      const mv = spec.metrics_view ?? "";
      const gd = spec.geo_dimension ?? "";

      const dims: { name: string }[] = [{ name: gd }];
      if (spec.tooltip_dimension) dims.push({ name: spec.tooltip_dimension });

      const meas: { name: string }[] = [];
      const cm = isMapColorConfig(spec.color) ? spec.color.measure : null;
      if (cm) meas.push({ name: cm });
      if (spec.size_measure) meas.push({ name: spec.size_measure });

      return getQueryServiceMetricsViewAggregationQueryOptions(
        runtimeVal.instanceId,
        mv || "_",
        {
          dimensions: dims,
          ...(meas.length > 0 ? { measures: meas } : {}),
          priority: 50,
        },
        {
          query: {
            enabled: !!(runtimeVal.instanceId && mv && gd),
          },
        },
      );
    },
  );

  const mapDataQuery = createQuery(queryOptionsStore);

  $: queryResult = $mapDataQuery;
  $: rows =
    ((queryResult?.data as V1MetricsViewAggregationResponse | undefined)
      ?.data as V1MetricsViewAggregationResponseDataItem[]) ?? [];

  $: geoJsonOpts = {
    geoDimension: geo_dimension,
    colorMeasure,
    sizeMeasure: size_measure,
    tooltipDimension: mapSpec?.tooltip_dimension,
  };

  function getColorPaint(
    geoJson: GeoJSON.FeatureCollection,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ): any {
    if (!isMapColorConfig(color)) {
      return resolveStaticColor(color || "primary", resolvedTheme);
    }

    const measure = color.measure;
    if (!measure) return resolveStaticColor("primary", resolvedTheme);

    const colorRange = color.colorRange ?? {
      mode: "scheme" as const,
      scheme: "tealblues" as const,
    };
    const colors = resolveColorRange(colorRange, resolvedTheme);
    const [min, max] = computeMinMax(geoJson.features, measure);
    return buildColorExpression(measure, min, max, colors);
  }

  function getRadiusPaint(
    geoJson: GeoJSON.FeatureCollection,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ): any {
    if (!size_measure) return 6;

    const [min, max] = computeMinMax(geoJson.features, size_measure);
    return buildSizeExpression(size_measure, min, max);
  }

  function updateMap(geoJson: GeoJSON.FeatureCollection) {
    if (!map) return;

    if (!map.isStyleLoaded()) {
      map.once("load", () => {
        updateMap(geoJson);
      });
      return;
    }

    const source = map.getSource("map-data") as mapboxgl.GeoJSONSource;
    if (source) {
      source.setData(geoJson);
    } else {
      map.addSource("map-data", {
        type: "geojson",
        data: geoJson,
      });

      map.addLayer({
        id: "points",
        type: "circle",
        source: "map-data",
        filter: ["==", ["geometry-type"], "Point"],
        paint: {},
      });

      map.addLayer({
        id: "polygons-fill",
        type: "fill",
        source: "map-data",
        filter: ["==", ["geometry-type"], "Polygon"],
        paint: {},
      });

      map.addLayer({
        id: "polygons-outline",
        type: "line",
        source: "map-data",
        filter: ["==", ["geometry-type"], "Polygon"],
        paint: {},
      });

      for (const layerId of ["points", "polygons-fill"]) {
        map.on("mousemove", layerId, (e) => {
          if (map) map.getCanvas().style.cursor = "pointer";
          showTooltip(e, tooltipCtx);
        });
        map.on("mouseleave", layerId, () => {
          if (map) map.getCanvas().style.cursor = "";
          removeTooltip();
        });
      }
    }

    const colorPaint = getColorPaint(geoJson);
    const radiusPaint = getRadiusPaint(geoJson);

    map.setPaintProperty("points", "circle-color", colorPaint);
    map.setPaintProperty("points", "circle-radius", radiusPaint);
    map.setPaintProperty("points", "circle-opacity", 0.8);
    map.setPaintProperty("points", "circle-stroke-width", 1);
    map.setPaintProperty("points", "circle-stroke-color", "#fff");

    map.setPaintProperty("polygons-fill", "fill-color", colorPaint);
    map.setPaintProperty("polygons-fill", "fill-opacity", 0.4);

    map.setPaintProperty("polygons-outline", "line-color", "#2563eb");
    map.setPaintProperty("polygons-outline", "line-width", 2);

    if (geoJson.features.length > 0) {
      const bounds = calculateBounds(geoJson.features);
      if (bounds) {
        map.fitBounds(bounds, {
          padding: 50,
          maxZoom: 10,
        });
      }
    }
  }

  onMount(() => {
    if (!mapContainer) return;

    mapboxgl.accessToken = MAPBOX_TOKEN;
    currentMapStyle = isThemeModeDark ? MAPBOX_STYLE_DARK : MAPBOX_STYLE_LIGHT;

    map = new mapboxgl.Map({
      container: mapContainer,
      style: currentMapStyle,
    });

    map.addControl(new mapboxgl.NavigationControl(), "top-right");
    map.addControl(
      new mapboxgl.ScaleControl({ maxWidth: 100, unit: "metric" }),
      "bottom-left",
    );

    map.on("load", () => {
      mapReady = true;
    });
  });

  // setStyle() strips all sources/layers, so re-add data after the new style loads
  $: {
    const targetStyle = isThemeModeDark
      ? MAPBOX_STYLE_DARK
      : MAPBOX_STYLE_LIGHT;
    if (map && mapReady && targetStyle !== currentMapStyle) {
      currentMapStyle = targetStyle;
      map.setStyle(targetStyle);
      if (rows.length > 0) {
        updateMap(transformToGeoJSON(rows, geoJsonOpts));
      }
    }
  }

  $: mapRenderDeps = {
    color,
    colorMeasure,
    size_measure,
    tooltipDim: mapSpec?.tooltip_dimension,
    geoDim: geo_dimension,
    resolvedTheme,
  };

  $: isPolygonMode = detectPolygonMode(rows, geo_dimension);
  $: if (component._isPolygonMode !== isPolygonMode) {
    component._isPolygonMode = isPolygonMode;
    component.specStore.update((s) => ({ ...s }));
  }

  $: if (mapReady && rows.length > 0 && mapRenderDeps) {
    updateMap(transformToGeoJSON(rows, geoJsonOpts));
  }

  onDestroy(() => {
    removeTooltip();
    if (map) {
      map.remove();
      map = null;
    }
  });
</script>

<div class="size-full flex flex-col overflow-hidden">
  <ComponentHeader
    faint={!title}
    {title}
    {description}
    showDescriptionAsTooltip={show_description_as_tooltip}
    {filters}
    {component}
  />
  <div class="relative flex-1 min-h-[300px]">
    <div bind:this={mapContainer} class="size-full" />
  </div>
</div>

<style>
  :global(#rill-map-tooltip) {
    position: absolute;
    padding: 8px 12px;
    border-radius: 5px;
    pointer-events: none;
    z-index: 1000;
    background: var(--tooltip);
    color: var(--fg-inverse);
    font-family: "Inter", sans-serif;
  }

  :global(#rill-map-tooltip h2) {
    font-size: 0.875rem;
    font-weight: 600;
    margin: 0 0 6px;
    color: color-mix(in oklab, var(--fg-inverse) 90%, transparent 30%);
  }

  :global(#rill-map-tooltip table) {
    border-collapse: separate;
    border-spacing: 0;
  }

  :global(#rill-map-tooltip table tr td) {
    padding: 2px 0;
    white-space: nowrap;
  }

  :global(#rill-map-tooltip table tr td.key) {
    text-align: left;
    font-weight: 400;
    font-size: 0.75rem;
    padding-right: 12px;
    color: color-mix(in oklab, var(--fg-inverse) 70%, transparent 30%);
  }

  :global(#rill-map-tooltip table tr td.value) {
    text-align: right;
    font-weight: 600;
    font-size: 0.75rem;
  }
</style>
