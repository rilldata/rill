<script lang="ts">
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { resolveThemeColors } from "@rilldata/web-common/features/themes/theme-utils";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import mapboxgl from "mapbox-gl";
  import "mapbox-gl/dist/mapbox-gl.css";
  import { onDestroy, onMount, untrack } from "svelte";
  import type { MapComponent, MapSpec } from ".";
  import MapPointsLayer from "./layers/MapPointsLayer.svelte";
  import MapRegionLayer from "./layers/MapRegionLayer.svelte";
  import type { MapLayerBoundsHandler, MapLayerId } from "./layers/types";
  import { removeTooltip } from "./map-utils";
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
  let map = $state.raw<mapboxgl.Map | null>(null);
  // Bumped each time a style finishes loading, so layers re-add their sources
  // and layers that `setStyle` dropped; null while a style is loading.
  let styleKey = $state<number | null>(null);
  let styleLoads = 0;
  let mapLoaded = $state(false);

  let currentMapStyle = "";
  // Latest bounds reported by each layer, and the data keys of the layers the
  // camera was last auto-fit to, so a filter change does not re-fit the camera.
  const layerBounds = new Map<
    MapLayerId,
    { dataKey: string; bounds: mapboxgl.LngLatBoundsLike | null }
  >();
  let fittedDataKey: string | null = null;

  const runtimeClient = useRuntimeClient();
  const { instanceId } = runtimeClient;

  // The component instance owns the stores below for the lifetime of this
  // render, so they are read once rather than tracked.
  const {
    specStore,
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
  const filters = $derived({
    time_filters: mapSpec.time_filters,
    dimension_filters: mapSpec.dimension_filters,
  });
  const hasPoints = $derived(!!mapSpec.geo_dimension?.field);
  const hasRegions = $derived(!!mapSpec.regions?.dimension);

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

  const metricsViewQuery = $derived(getMetricsViewFromName(metricsView));
  const metricsViewSpec = $derived($metricsViewQuery?.metricsView);

  // Auto-fit the camera once per dataset, after every active layer has
  // reported its bounds. Re-fitting on every filter change moves the map under
  // the user and discards any panning or zooming they did.
  // An `initial_view` in the spec opts out of auto-fitting entirely.
  const handleLayerBounds: MapLayerBoundsHandler = (layer, dataKey, bounds) => {
    layerBounds.set(layer, { dataKey, bounds });
    if (!map || hasInitialView) return;

    const activeLayers: MapLayerId[] = [];
    if (hasRegions) activeLayers.push("regions");
    if (hasPoints) activeLayers.push("points");
    const reported = activeLayers.map((l) => layerBounds.get(l));
    if (reported.some((r) => !r)) return;

    const combinedKey = reported.map((r) => r?.dataKey).join("|");
    if (fittedDataKey === combinedKey) return;

    const combined = new mapboxgl.LngLatBounds();
    for (const r of reported) {
      if (r?.bounds) combined.extend(r.bounds);
    }
    if (combined.isEmpty()) return;

    fittedDataKey = combinedKey;
    map.fitBounds(combined, { padding: 50, maxZoom: 10 });
  };

  onMount(() => {
    if (!mapContainer) return;

    mapboxgl.accessToken = MAPBOX_ACCESS_TOKEN;
    currentMapStyle = isThemeModeDark ? MAPBOX_STYLE_DARK : MAPBOX_STYLE_LIGHT;

    const newMap = new mapboxgl.Map({
      container: mapContainer,
      style: currentMapStyle,
      ...(initialLongitude !== undefined && initialLatitude !== undefined
        ? {
            center: [initialLongitude, initialLatitude] as [number, number],
            ...(initialZoom !== undefined && { zoom: initialZoom }),
          }
        : {}),
    });

    newMap.addControl(new mapboxgl.NavigationControl(), "top-right");
    newMap.addControl(
      new mapboxgl.ScaleControl({ maxWidth: 100, unit: "metric" }),
      "bottom-left",
    );

    // Fires for the initial style and again after every `setStyle`.
    newMap.on("style.load", () => {
      styleKey = ++styleLoads;
    });
    newMap.on("load", () => {
      mapLoaded = true;
    });

    map = newMap;
  });

  // Keeps the camera in sync while the initial view is being edited.
  $effect(() => {
    if (!map || !mapLoaded) return;
    if (initialLongitude === undefined || initialLatitude === undefined) return;
    map.jumpTo({
      center: [initialLongitude, initialLatitude],
      ...(initialZoom !== undefined && { zoom: initialZoom }),
    });
  });

  // setStyle() strips all sources and layers. Layers wait for the new style
  // to finish loading before re-adding them, since anything added while it
  // loads is wiped.
  $effect(() => {
    const targetStyle = isThemeModeDark
      ? MAPBOX_STYLE_DARK
      : MAPBOX_STYLE_LIGHT;
    if (!map || targetStyle === currentMapStyle) return;

    currentMapStyle = targetStyle;
    styleKey = null;
    map.setStyle(targetStyle);
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
    {#if map}
      <!-- Layers draw on the map; the markup they render is their legend. -->
      <div class="absolute top-2 left-2 flex flex-col gap-1">
        {#if hasRegions}
          <MapRegionLayer
            {map}
            {styleKey}
            {component}
            {metricsViewSpec}
            theme={resolvedTheme}
            {isThemeModeDark}
            onBounds={handleLayerBounds}
          />
        {/if}
        {#if hasPoints}
          <MapPointsLayer
            {map}
            {styleKey}
            {component}
            {metricsViewSpec}
            theme={resolvedTheme}
            {isThemeModeDark}
            onBounds={handleLayerBounds}
          />
        {/if}
      </div>
    {/if}
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
