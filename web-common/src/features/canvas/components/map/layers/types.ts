import type mapboxgl from "mapbox-gl";

export type MapLayerId = "points" | "regions";

/**
 * Reports the extent of a layer's data so the map can fit its camera.
 * The data key identifies the dataset the bounds belong to: the camera is fit once per combination of data keys,
 * so a filter change, which only changes the data, does not move the camera.
 */
export type MapLayerBoundsHandler = (
  layer: MapLayerId,
  dataKey: string,
  bounds: mapboxgl.LngLatBoundsLike | null,
) => void;
