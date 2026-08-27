<script lang="ts">
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { canQueryWithTimeRange } from "@rilldata/web-common/features/components/charts/query-util";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { resolveThemeColors } from "@rilldata/web-common/features/themes/theme-utils";
  import {
    getQueryServiceMetricsViewAggregationQueryOptions,
    type V1MetricsViewAggregationDimension,
    type V1MetricsViewAggregationMeasure,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { createQuery, keepPreviousData } from "@tanstack/svelte-query";
  import type * as GeoJSON from "geojson";
  import mapboxgl from "mapbox-gl";
  import "mapbox-gl/dist/mapbox-gl.css";
  import { onDestroy, onMount, untrack } from "svelte";
  import { derived } from "svelte/store";
  import { DEFAULT_MAP_COLOR_RANGE, type MapComponent, type MapSpec } from ".";
  import {
    buildColorExpression,
    buildSizeExpression,
    computeMinMax,
    resolveColorRange,
  } from "./color-utils";
  import {
    calculateBounds,
    detectPolygonMode,
    removeTooltip,
    showTooltip,
    transformToGeoJSON,
  } from "./map-utils";
  import {
    MAPBOX_ACCESS_TOKEN,
    MAPBOX_STYLE_DARK,
    MAPBOX_STYLE_LIGHT,
  } from "./mapbox";

  type Props = {
    component: MapComponent;
  };

  let { component }: Props = $props();

  let mapContainer = $state<HTMLDivElement>();
  let mapReady = $state(false);

  let map: mapboxgl.Map | null = null;
  let currentMapStyle = "";
  let layerHandlersRegistered = false;
  // Identifies the dataset the camera was last auto-fit to, so a filter change
  // does not re-fit the camera.
  let fittedDataKey: string | null = null;

  const runtimeClient = useRuntimeClient();
  const { instanceId } = runtimeClient;

  // The component instance owns the stores below for the lifetime of this
  // render, so they are read once rather than tracked.
  const {
    specStore,
    timeAndFilterStore,
    dataEnabled: visible,
    parent: {
      name: canvasName,
      metricsView: { getMetricsViewFromName },
    },
  } = untrack(() => component);

  const mapSpec = $derived($specStore ?? ({} as Partial<MapSpec>));
  const title = $derived(mapSpec.title);
  const description = $derived(mapSpec.description);
  const showDescriptionAsTooltip = $derived(
    mapSpec.show_description_as_tooltip,
  );
  const metricsView = $derived(mapSpec.metrics_view ?? "");
  const geoDimension = $derived(mapSpec.geo_dimension?.field ?? "");
  const tooltipDimension = $derived(mapSpec.tooltip_dimension?.field);
  const color = $derived(mapSpec.color);
  const sizeMeasure = $derived(mapSpec.size_measure?.field);
  const filters = $derived({
    time_filters: mapSpec.time_filters,
    dimension_filters: mapSpec.dimension_filters,
  });

  // Read as primitives so that editing the view in YAML moves the camera,
  // without re-running on every unrelated spec update.
  const initialLongitude = $derived(mapSpec.initial_view?.longitude);
  const initialLatitude = $derived(mapSpec.initial_view?.latitude);
  const initialZoom = $derived(mapSpec.initial_view?.zoom);
  const hasInitialView = $derived(
    initialLongitude !== undefined && initialLatitude !== undefined,
  );

  const isThemeModeDark = $derived($themeControl === "dark");
  const canvasStore = $derived(getCanvasStore(canvasName, instanceId));
  const canvasTheme = $derived(canvasStore.canvasEntity.theme);
  const resolvedTheme = $derived(
    resolveThemeColors($canvasTheme?.spec, isThemeModeDark),
  );

  const colorMeasure = $derived(color?.measure ?? "");

  const metricsViewQuery = $derived(getMetricsViewFromName(metricsView));
  const metricsViewSpec = $derived($metricsViewQuery?.metricsView);

  function getMeasureDisplayName(measureName: string): string {
    const measure = metricsViewSpec?.measures?.find(
      (m) => m.name === measureName,
    );
    return measure?.displayName || measureName;
  }

  const tooltipCtx = $derived({
    tooltipDimension,
    colorMeasure,
    sizeMeasure,
    getDisplayName: getMeasureDisplayName,
  });

  const queryOptionsStore = derived(
    [specStore, timeAndFilterStore, visible],
    ([specVal, $timeAndFilterStore, $visible]) => {
      const spec = specVal ?? ({} as Partial<MapSpec>);
      const mv = spec.metrics_view ?? "";
      const gd = spec.geo_dimension?.field ?? "";

      const { timeRange, where, hasTimeSeries } = $timeAndFilterStore;

      const dimensions: V1MetricsViewAggregationDimension[] = [{ name: gd }];
      if (spec.tooltip_dimension?.field) {
        dimensions.push({ name: spec.tooltip_dimension.field });
      }

      // Color is always measure-driven, so there is nothing to draw without it.
      const cm = spec.color?.measure ?? "";
      const measures: V1MetricsViewAggregationMeasure[] = [];
      if (cm) measures.push({ name: cm });
      if (spec.size_measure?.field)
        measures.push({ name: spec.size_measure.field });

      const enabled =
        $visible &&
        !!mv &&
        !!gd &&
        !!cm &&
        canQueryWithTimeRange(hasTimeSeries, timeRange);

      return getQueryServiceMetricsViewAggregationQueryOptions(
        runtimeClient,
        {
          metricsView: mv,
          dimensions,
          measures,
          limit: "5000",
          where,
          timeRange: hasTimeSeries ? timeRange : undefined,
        },
        {
          query: {
            enabled,
            placeholderData: keepPreviousData,
          },
        },
      );
    },
  );

  const mapDataQuery = createQuery(queryOptionsStore);

  const rows = $derived($mapDataQuery.data?.data ?? []);
  // `isPending` stays true while the query is disabled (incomplete spec), so
  // nothing renders until the first result arrives. The color measure is
  // checked separately because a spec edit can leave placeholder data from the
  // previous query in place.
  const isPending = $derived($mapDataQuery.isPending);
  const canRender = $derived(mapReady && !isPending && !!colorMeasure);

  const geoJsonOpts = $derived({
    geoDimension,
    colorMeasure,
    sizeMeasure,
    tooltipDimension,
  });

  // Only reached once `canRender` holds, so `colorMeasure` is set.
  function getColorPaint(
    geoJson: GeoJSON.FeatureCollection,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ): any {
    const colors = resolveColorRange(
      color?.colorRange ?? DEFAULT_MAP_COLOR_RANGE,
      resolvedTheme,
    );
    const [min, max] = computeMinMax(geoJson.features, colorMeasure);
    return buildColorExpression(colorMeasure, min, max, colors);
  }

  function getRadiusPaint(
    geoJson: GeoJSON.FeatureCollection,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ): any {
    if (!sizeMeasure) return 6;

    const [min, max] = computeMinMax(geoJson.features, sizeMeasure);
    return buildSizeExpression(sizeMeasure, min, max);
  }

  // Auto-fit the camera once per dataset. Re-fitting on every filter change
  // moves the map under the user and discards any panning or zooming they did.
  // An `initial_view` in the spec opts out of auto-fitting entirely.
  function fitToData(geoJson: GeoJSON.FeatureCollection) {
    if (!map || hasInitialView) return;

    const dataKey = `${metricsView}::${geoDimension}`;
    if (fittedDataKey === dataKey) return;

    const bounds = calculateBounds(geoJson.features);
    if (!bounds) return;

    fittedDataKey = dataKey;
    map.fitBounds(bounds, { padding: 50, maxZoom: 10 });
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

      // Layer event handlers persist across style reloads (they bind by layer
      // id), so register them only once to avoid stacking duplicates.
      if (!layerHandlersRegistered) {
        layerHandlersRegistered = true;
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
    }

    const colorPaint = getColorPaint(geoJson);
    const radiusPaint = getRadiusPaint(geoJson);

    // Contrast ring around points: white on the light base map, dark on the
    // dark base map, so the stroke stays visible in either theme mode.
    const strokeColor = isThemeModeDark ? "#1a1a1a" : "#fff";
    const outlineColor = resolvedTheme.primary.hex();

    map.setPaintProperty("points", "circle-color", colorPaint);
    map.setPaintProperty("points", "circle-radius", radiusPaint);
    map.setPaintProperty("points", "circle-opacity", 0.8);
    map.setPaintProperty("points", "circle-stroke-width", 1);
    map.setPaintProperty("points", "circle-stroke-color", strokeColor);

    map.setPaintProperty("polygons-fill", "fill-color", colorPaint);
    map.setPaintProperty("polygons-fill", "fill-opacity", 0.4);

    map.setPaintProperty("polygons-outline", "line-color", outlineColor);
    map.setPaintProperty("polygons-outline", "line-width", 2);

    fitToData(geoJson);
  }

  onMount(() => {
    if (!mapContainer) return;

    mapboxgl.accessToken = MAPBOX_ACCESS_TOKEN;
    currentMapStyle = isThemeModeDark ? MAPBOX_STYLE_DARK : MAPBOX_STYLE_LIGHT;

    map = new mapboxgl.Map({
      container: mapContainer,
      style: currentMapStyle,
      ...(initialLongitude !== undefined && initialLatitude !== undefined
        ? {
            center: [initialLongitude, initialLatitude] as [number, number],
            ...(initialZoom !== undefined && { zoom: initialZoom }),
          }
        : {}),
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

  // Keeps the camera in sync while the initial view is being edited.
  $effect(() => {
    if (!map || !mapReady) return;
    if (initialLongitude === undefined || initialLatitude === undefined) return;
    map.jumpTo({
      center: [initialLongitude, initialLatitude],
      ...(initialZoom !== undefined && { zoom: initialZoom }),
    });
  });

  // setStyle() strips all sources/layers, so re-add data only after the new
  // style has finished loading. Re-adding synchronously would race the async
  // style load and get wiped, causing the points to disappear.
  $effect(() => {
    const targetStyle = isThemeModeDark
      ? MAPBOX_STYLE_DARK
      : MAPBOX_STYLE_LIGHT;
    if (!map || !mapReady || targetStyle === currentMapStyle) return;

    currentMapStyle = targetStyle;
    map.setStyle(targetStyle);
    map.once("style.load", () => {
      if (canRender) updateMap(transformToGeoJSON(rows, geoJsonOpts));
    });
  });

  $effect(() => {
    const isPolygonMode = detectPolygonMode(rows, geoDimension);
    if (component._isPolygonMode === isPolygonMode) return;
    component._isPolygonMode = isPolygonMode;
    component.specStore.update((s) => ({ ...s }));
  });

  // Reruns on any spec, theme or data change read by `updateMap`.
  $effect(() => {
    if (!canRender) return;
    updateMap(transformToGeoJSON(rows, geoJsonOpts));
  });

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
    {showDescriptionAsTooltip}
    {filters}
    {component}
  />
  <div class="relative flex-1 min-h-[300px]">
    <div bind:this={mapContainer} class="size-full"></div>
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
