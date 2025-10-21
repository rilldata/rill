<script lang="ts">
  import type { MapCanvasComponent } from "@rilldata/web-common/features/canvas/components/map";
  import ComponentHeader from "../../ComponentHeader.svelte";
  import { validateMapSchema } from "./selector";
  import { createQueryServiceMetricsViewAggregation } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount, onDestroy } from "svelte";
  import mapboxgl from "mapbox-gl";
  import { cellToBoundary, getResolution, cellToParent } from "h3-js";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import "mapbox-gl/dist/mapbox-gl.css";

  export let component: MapCanvasComponent;

  $: ({
    parent: {
      metricsView: { getMetricsViewFromName, getMeasureForMetricView },
    },
    specStore,
    timeAndFilterStore,
  } = component);

  $: mapSpec = $specStore;

  $: ({ title, description, dimension_filters, time_filters } = mapSpec);

  $: filters = {
    time_filters,
    dimension_filters,
  };

  $: _metricViewSpec = getMetricsViewFromName(mapSpec.metrics_view);
  $: metricsViewSpec = $_metricViewSpec.metricsView;

  $: schema = validateMapSchema(metricsViewSpec, mapSpec);

  // Get measure metadata for formatting
  $: measureStore = getMeasureForMetricView(
    mapSpec.measure,
    mapSpec.metrics_view,
  );
  $: measure = $measureStore;
  $: measureFormatter = measure
    ? createMeasureValueFormatter(measure)
    : (v: number) => String(v);

  // Get size measure metadata if specified
  $: sizeMeasureStore = mapSpec.size_measure
    ? getMeasureForMetricView(mapSpec.size_measure, mapSpec.metrics_view)
    : null;
  $: sizeMeasure = sizeMeasureStore ? $sizeMeasureStore : null;
  $: sizeMeasureFormatter = sizeMeasure
    ? createMeasureValueFormatter(sizeMeasure)
    : (v: number) => String(v);

  $: ({ instanceId } = $runtime);

  $: ({
    timeRange: { timeZone, start, end },
    where,
  } = $timeAndFilterStore);

  // Query data - include size_measure and label_dimension if specified
  $: measures = mapSpec.size_measure
    ? [{ name: mapSpec.measure }, { name: mapSpec.size_measure }]
    : [{ name: mapSpec.measure }];

  $: dimensions = mapSpec.label_dimension
    ? [{ name: mapSpec.dimension }, { name: mapSpec.label_dimension }]
    : [{ name: mapSpec.dimension }];

  $: dataQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    mapSpec.metrics_view,
    {
      dimensions,
      measures,
      timeRange: {
        start,
        end,
        timeZone,
      },
      where,
      priority: 50,
    },
    {
      query: {
        enabled: schema.isValid && !!start && !!end,
      },
    },
  );

  $: queryData = $dataQuery;
  $: data = queryData?.data?.data || [];
  $: isError = queryData?.isError || false;
  $: isFetching = queryData?.isFetching || false;

  // Auto-detect and set geometry_type if not explicitly set
  let lastAutoDetectedType: "h3" | "geojson" | null = null;
  $: if (data.length > 0 && !mapSpec.geometry_type) {
    const firstValue = data[0]?.[mapSpec.dimension];
    const detectedType = detectGeometryType(firstValue);

    // Only update if detection result changed (avoid infinite loops)
    if (detectedType !== lastAutoDetectedType) {
      lastAutoDetectedType = detectedType;
      specStore.update((spec) => ({
        ...spec,
        geometry_type: detectedType,
      }));
    }
  } else if (mapSpec.geometry_type) {
    // User has explicitly set the type, clear our tracking
    lastAutoDetectedType = null;
  }
  $: errorMessage = queryData?.error?.message || "An error occurred";

  // Map state
  let mapContainer: HTMLDivElement;
  let map: mapboxgl.Map | null = null;
  let mounted = false;
  let mapLoaded = false;
  let initialDataLoaded = false;

  // Calculate bounding box for all H3 cells
  function calculateBounds(
    h3Cells: string[],
  ): mapboxgl.LngLatBoundsLike | null {
    if (h3Cells.length === 0) return null;

    let minLng = Infinity;
    let maxLng = -Infinity;
    let minLat = Infinity;
    let maxLat = -Infinity;

    for (const h3Index of h3Cells) {
      try {
        const boundary = cellToBoundary(h3Index, true);
        for (const [lng, lat] of boundary) {
          minLng = Math.min(minLng, lng);
          maxLng = Math.max(maxLng, lng);
          minLat = Math.min(minLat, lat);
          maxLat = Math.max(maxLat, lat);
        }
      } catch (e) {
        console.warn(`Invalid H3 index: ${h3Index}`, e);
      }
    }

    if (minLng === Infinity) return null;

    return [
      [minLng, minLat],
      [maxLng, maxLat],
    ];
  }

  // Initialize map
  onMount(() => {
    mounted = true;
    initializeMap();

    return () => {
      if (map) {
        map.remove();
        map = null;
      }
    };
  });

  function initializeMap(
    initialCenter?: { lng: number; lat: number },
    initialZoom?: number,
  ) {
    if (!mapContainer) return;

    const accessToken =
      "pk.eyJ1IjoicmlsbGRhdGEiLCJhIjoiY21nemp4Mnl3MDViaGQzc2J0MzB1NjdvMiJ9.4Q8jXek0-EF4RLA_TF4-oA";
    mapboxgl.accessToken = accessToken;

    map = new mapboxgl.Map({
      container: mapContainer,
      style: "mapbox://styles/mapbox/light-v11",
      center: initialCenter || [0, 0],
      zoom: initialZoom || 2,
      projection: "mercator" as any,
    });

    map.addControl(new mapboxgl.NavigationControl(), "top-right");

    map.on("load", () => {
      if (!map) return;

      const canvas = map.getCanvas();
      if (canvas) {
        canvas.setAttribute("data-no-drag", "true");
      }

      map.addSource("h3-data", {
        type: "geojson",
        data: {
          type: "FeatureCollection",
          features: [],
        },
      });

      mapLoaded = true;

      map.addLayer({
        id: "h3-fill",
        type: "fill",
        source: "h3-data",
        paint: {
          "fill-color": ["get", "color"],
          "fill-opacity": 0.7,
        },
      });

      map.addLayer({
        id: "h3-outline",
        type: "line",
        source: "h3-data",
        paint: {
          "line-color": "#ffffff",
          "line-width": 1,
        },
      });

      map.addLayer({
        id: "h3-points",
        type: "circle",
        source: "h3-data",
        filter: ["==", ["geometry-type"], "Point"],
        paint: {
          "circle-color": ["get", "color"],
          "circle-radius": ["get", "pointSize"],
          "circle-stroke-width": 1,
          "circle-stroke-color": "#ffffff",
          "circle-opacity": 0.8,
        },
      });

      setupHoverHandlers();
    });
  }

  function setupHoverHandlers() {
    if (!map) return;

    let popup: mapboxgl.Popup | null = null;

    // Handle hover for fill layer (polygons)
    map.on("mouseenter", "h3-fill", () => {
      if (map) map.getCanvas().style.cursor = "pointer";
    });

    map.on("mousemove", "h3-fill", (e) => {
      if (!map || !e.features || e.features.length === 0) return;

      const feature = e.features[0];
      const value = feature.properties?.value;
      const label = feature.properties?.label;

      const formattedValue = measureFormatter(value);
      const measureLabel = measure?.displayName || mapSpec.measure;

      if (!popup) {
        popup = new mapboxgl.Popup({
          closeButton: false,
          closeOnClick: false,
        });
        popup.addTo(map);
      }

      let content = `<div style="font-size: 12px; padding: 4px;">`;

      // Add label if available
      if (label) {
        content += `<strong>${label}</strong><br>`;
      }

      content += `<strong>${measureLabel}:</strong> ${formattedValue}</div>`;

      popup.setLngLat(e.lngLat).setHTML(content);
    });

    map.on("mouseleave", "h3-fill", () => {
      if (map) map.getCanvas().style.cursor = "";
      if (popup) {
        popup.remove();
        popup = null;
      }
    });

    // Handle hover for circle layer (points)
    map.on("mouseenter", "h3-points", () => {
      if (map) map.getCanvas().style.cursor = "pointer";
    });

    map.on("mousemove", "h3-points", (e) => {
      if (!map || !e.features || e.features.length === 0) return;

      const feature = e.features[0];
      const value = feature.properties?.value;
      const sizeValue = feature.properties?.sizeValue;
      const label = feature.properties?.label;

      const formattedValue = measureFormatter(value);
      const measureLabel = measure?.displayName || mapSpec.measure;

      if (!popup) {
        popup = new mapboxgl.Popup({
          closeButton: false,
          closeOnClick: false,
        });
        popup.addTo(map);
      }

      let content = `<div style="font-size: 12px; padding: 4px;">`;

      // Add label if available
      if (label) {
        content += `<strong>${label}</strong><br>`;
      }

      content += `<strong>${measureLabel}:</strong> ${formattedValue}`;

      // Add size measure if available
      if (mapSpec.size_measure && sizeValue && sizeMeasure) {
        const formattedSizeValue = sizeMeasureFormatter(sizeValue);
        const sizeLabel = sizeMeasure.displayName || mapSpec.size_measure;
        content += `<br><strong>${sizeLabel}:</strong> ${formattedSizeValue}`;
      }

      content += `</div>`;

      popup.setLngLat(e.lngLat).setHTML(content);
    });

    map.on("mouseleave", "h3-points", () => {
      if (map) map.getCanvas().style.cursor = "";
      if (popup) {
        popup.remove();
        popup = null;
      }
    });
  }

  onDestroy(() => {
    if (map) {
      map.remove();
      map = null;
    }
  });

  // Reset auto-detect tracking when dimension or metrics view changes
  $: {
    void mapSpec.dimension;
    void mapSpec.metrics_view;
    lastAutoDetectedType = null;
  }

  // Update map data when data or any relevant property changes
  $: if (map && mapLoaded && mounted && data.length > 0 && schema.isValid) {
    // Include all relevant properties in reactive dependencies to trigger redraw
    void mapSpec.metrics_view;
    void mapSpec.dimension;
    void mapSpec.measure;
    void mapSpec.resolution;
    void mapSpec.geometry_type;
    void mapSpec.size_measure;
    void mapSpec.label_dimension;
    updateMapData();
  }

  // Auto-detect geometry type from data
  function detectGeometryType(sampleValue: unknown): "h3" | "geojson" {
    if (!sampleValue) return "h3";

    const strValue = String(sampleValue);

    // Try to detect H3 index (15-character hex string)
    if (/^[0-9a-f]{15}$/i.test(strValue)) {
      return "h3";
    }

    // Try to parse as JSON (GeoJSON)
    try {
      const parsed = JSON.parse(strValue);
      if (
        parsed &&
        typeof parsed === "object" &&
        (parsed.type === "Point" ||
          parsed.type === "LineString" ||
          parsed.type === "Polygon" ||
          parsed.type === "MultiPoint" ||
          parsed.type === "MultiLineString" ||
          parsed.type === "MultiPolygon" ||
          parsed.type === "GeometryCollection")
      ) {
        return "geojson";
      }
    } catch {
      // Not valid JSON, likely H3
    }

    // Default to H3
    return "h3";
  }

  function updateMapData() {
    if (!map || !data || data.length === 0) return;

    // Detect or use specified geometry type
    const firstValue = data[0]?.[mapSpec.dimension];
    const geometryType =
      mapSpec.geometry_type || detectGeometryType(firstValue);

    let features: Array<{
      type: "Feature";
      properties: { value: number; color: string; [key: string]: unknown };
      geometry: GeoJSON.Geometry;
    }> = [];

    if (geometryType === "h3") {
      features = processH3Data();
    } else {
      features = processGeoJSONData();
    }

    if (features.length === 0) return;

    // Update map source (replaces all existing features)
    const source = map.getSource("h3-data") as mapboxgl.GeoJSONSource;
    if (source) {
      source.setData({
        type: "FeatureCollection",
        features,
      });
    }

    // Fit map to show all features only on initial load
    if (!initialDataLoaded) {
      fitMapToFeatures(features);
      initialDataLoaded = true;
    }
  }

  function processH3Data() {
    // Extract H3 cells, values, and optional label
    let h3Data = data
      .map((row: Record<string, unknown>) => ({
        h3Index: String(row[mapSpec.dimension] || ""),
        value: Number(row[mapSpec.measure]) || 0,
        label: mapSpec.label_dimension
          ? String(row[mapSpec.label_dimension] || "")
          : undefined,
      }))
      .filter((item) => item.h3Index);

    if (h3Data.length === 0) return [];

    // Aggregate to target resolution if specified
    if (mapSpec.resolution !== undefined && mapSpec.resolution !== null) {
      const aggregated = new Map<
        string,
        { value: number; labels: Set<string> }
      >();

      for (const { h3Index, value, label } of h3Data) {
        try {
          const currentRes = getResolution(h3Index);
          let targetIndex = h3Index;

          // Convert to target resolution if different
          if (currentRes > mapSpec.resolution) {
            // Aggregate up to parent cell
            targetIndex = cellToParent(h3Index, mapSpec.resolution);
          } else if (currentRes < mapSpec.resolution) {
            // Can't subdivide to higher resolution from data, use current
            targetIndex = h3Index;
          }

          // Sum values for same target cell
          const existing = aggregated.get(targetIndex) || {
            value: 0,
            labels: new Set<string>(),
          };
          existing.value += value;
          if (label) existing.labels.add(label);
          aggregated.set(targetIndex, existing);
        } catch (e) {
          console.warn(`Could not process H3 index: ${h3Index}`, e);
        }
      }

      h3Data = Array.from(aggregated.entries()).map(
        ([h3Index, { value, labels }]) => ({
          h3Index,
          value,
          // If multiple labels after aggregation, show count; if one label, show it
          label:
            labels.size === 1
              ? Array.from(labels)[0]
              : labels.size > 1
                ? `${labels.size} locations`
                : undefined,
        }),
      );
    }

    // Calculate min/max for color scale
    const values = h3Data
      .map((d) => d.value)
      .filter((v) => typeof v === "number");
    const min = Math.min(...values);
    const max = Math.max(...values);

    // Create GeoJSON features from H3 cells
    return h3Data
      .map(({ h3Index, value, label }) => {
        try {
          const boundary = cellToBoundary(h3Index, true);
          const coordinates = [...boundary, boundary[0]];

          // Calculate color based on value
          const ratio = max === min ? 0.5 : (value - min) / (max - min);
          const color = `rgba(8, 81, 156, ${0.2 + ratio * 0.8})`;

          return {
            type: "Feature" as const,
            properties: {
              h3Index,
              value,
              label,
              color,
            },
            geometry: {
              type: "Polygon" as const,
              coordinates: [coordinates],
            },
          };
        } catch (e) {
          console.warn(`Invalid H3 index: ${h3Index}`, e);
          return null;
        }
      })
      .filter((f): f is NonNullable<typeof f> => f !== null);
  }

  // Validate coordinates are within valid bounds
  function isValidCoordinate(coord: number[]): boolean {
    const [lng, lat] = coord;
    return (
      lng >= -180 &&
      lng <= 180 &&
      lat >= -90 &&
      lat <= 90 &&
      !isNaN(lng) &&
      !isNaN(lat)
    );
  }

  // Validate geometry has valid coordinates
  function isValidGeometry(geometry: GeoJSON.Geometry): boolean {
    try {
      if (geometry.type === "Point") {
        return isValidCoordinate(geometry.coordinates);
      } else if (geometry.type === "LineString") {
        return geometry.coordinates.every(isValidCoordinate);
      } else if (geometry.type === "Polygon") {
        return geometry.coordinates.every((ring) =>
          ring.every(isValidCoordinate),
        );
      } else if (geometry.type === "MultiPoint") {
        return geometry.coordinates.every(isValidCoordinate);
      } else if (geometry.type === "MultiLineString") {
        return geometry.coordinates.every((line) =>
          line.every(isValidCoordinate),
        );
      } else if (geometry.type === "MultiPolygon") {
        return geometry.coordinates.every((polygon) =>
          polygon.every((ring) => ring.every(isValidCoordinate)),
        );
      }
      return false;
    } catch (e) {
      return false;
    }
  }

  function processGeoJSONData() {
    // Extract GeoJSON geometries, values, and optional label
    const geoData = data
      .map((row: Record<string, unknown>) => {
        const geomValue = row[mapSpec.dimension];
        const value = Number(row[mapSpec.measure]) || 0;
        const sizeValue = mapSpec.size_measure
          ? Number(row[mapSpec.size_measure]) || 0
          : 0;
        const label = mapSpec.label_dimension
          ? String(row[mapSpec.label_dimension] || "")
          : undefined;

        if (!geomValue) return null;

        let geometry: GeoJSON.Geometry | null = null;

        // Try to parse as JSON if it's a string
        if (typeof geomValue === "string") {
          try {
            geometry = JSON.parse(geomValue);
          } catch {
            return null;
          }
        } else if (typeof geomValue === "object") {
          geometry = geomValue as GeoJSON.Geometry;
        }

        if (!geometry || !geometry.type) return null;

        // Validate geometry coordinates
        if (!isValidGeometry(geometry)) {
          console.warn("Invalid geometry coordinates, skipping:", geometry);
          return null;
        }

        return { geometry, value, sizeValue, label };
      })
      .filter((item) => item !== null) as Array<{
      geometry: GeoJSON.Geometry;
      value: number;
      sizeValue: number;
      label?: string;
    }>;

    if (geoData.length === 0) return [];

    // Calculate min/max for color scale
    const values = geoData
      .map((d) => d.value)
      .filter((v) => typeof v === "number");
    const min = Math.min(...values);
    const max = Math.max(...values);

    // Calculate min/max for size scale if size measure is provided
    let sizeMin = 0;
    let sizeMax = 0;
    if (mapSpec.size_measure) {
      const sizeValues = geoData
        .map((d) => d.sizeValue)
        .filter((v) => typeof v === "number" && v > 0);
      if (sizeValues.length > 0) {
        sizeMin = Math.min(...sizeValues);
        sizeMax = Math.max(...sizeValues);
      }
    }

    // Create GeoJSON features
    return geoData.map(({ geometry, value, sizeValue, label }) => {
      const ratio = max === min ? 0.5 : (value - min) / (max - min);
      const color = `rgba(8, 81, 156, ${0.2 + ratio * 0.8})`;

      // Calculate point size if size measure is provided
      let pointSize = 5; // Default size
      if (mapSpec.size_measure && sizeMax > sizeMin) {
        const sizeRatio = (sizeValue - sizeMin) / (sizeMax - sizeMin);
        pointSize = 3 + sizeRatio * 12; // Range from 3 to 15
      }

      return {
        type: "Feature" as const,
        properties: {
          value,
          sizeValue,
          pointSize,
          label,
          color,
        },
        geometry,
      };
    });
  }

  function fitMapToFeatures(
    features: Array<{ geometry: GeoJSON.Geometry; [key: string]: unknown }>,
  ) {
    if (!map || features.length === 0) return;

    // Calculate bounds from all geometries
    let minLng = Infinity;
    let minLat = Infinity;
    let maxLng = -Infinity;
    let maxLat = -Infinity;

    features.forEach((feature) => {
      const geom = feature.geometry;

      // Helper to update bounds from coordinates
      const updateBounds = (coords: number[]) => {
        const [lng, lat] = coords;
        minLng = Math.min(minLng, lng);
        minLat = Math.min(minLat, lat);
        maxLng = Math.max(maxLng, lng);
        maxLat = Math.max(maxLat, lat);
      };

      // Handle different geometry types
      if (geom.type === "Point") {
        updateBounds(geom.coordinates as number[]);
      } else if (geom.type === "LineString" || geom.type === "MultiPoint") {
        (geom.coordinates as number[][]).forEach(updateBounds);
      } else if (geom.type === "Polygon" || geom.type === "MultiLineString") {
        (geom.coordinates as number[][][]).forEach((ring) =>
          ring.forEach(updateBounds),
        );
      } else if (geom.type === "MultiPolygon") {
        (geom.coordinates as number[][][][]).forEach((polygon) =>
          polygon.forEach((ring) => ring.forEach(updateBounds)),
        );
      }
    });

    if (
      isFinite(minLng) &&
      isFinite(minLat) &&
      isFinite(maxLng) &&
      isFinite(maxLat)
    ) {
      map.fitBounds(
        [
          [minLng, minLat],
          [maxLng, maxLat],
        ],
        {
          padding: 50,
          duration: 1000,
          maxZoom: 15,
        },
      );
    }
  }
</script>

<div class="h-full w-full flex flex-col">
  <ComponentHeader {component} {title} {description} {filters} />

  {#if !schema.isValid}
    <div class="flex items-center justify-center h-full p-4 text-gray-500">
      <div class="text-center">
        <div class="text-lg font-medium mb-2">Map Configuration Required</div>
        <div class="text-sm">
          {schema.error || "Please select a dimension and measure"}
        </div>
      </div>
    </div>
  {:else if isError}
    <div class="flex items-center justify-center h-full p-4 text-red-500">
      <div class="text-center">
        <div class="text-lg font-medium mb-2">Error Loading Data</div>
        <div class="text-sm">{errorMessage}</div>
      </div>
    </div>
  {:else}
    <div
      bind:this={mapContainer}
      class="w-full flex-1 map-container"
      class:opacity-50={isFetching}
    />
  {/if}
</div>

<style>
  :global(.mapboxgl-canvas) {
    outline: none;
  }

  .map-container :global(.mapboxgl-canvas-container) {
    cursor: grab;
  }

  .map-container :global(.mapboxgl-canvas-container:active) {
    cursor: grabbing;
  }
</style>
