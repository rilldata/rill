import { cellToBoundary } from "h3-js";

/**
 * Convert H3 cell index to GeoJSON polygon
 */
export function h3ToGeoJSON(h3Index: string) {
  const boundary = cellToBoundary(h3Index, true); // true for GeoJSON format [lng, lat]
  
  // Close the polygon by adding the first point at the end
  const coordinates = [...boundary, boundary[0]];
  
  return {
    type: "Feature" as const,
    properties: {
      h3Index,
    },
    geometry: {
      type: "Polygon" as const,
      coordinates: [coordinates],
    },
  };
}

/**
 * Create a FeatureCollection from H3 cells with their measure values
 */
export function createH3FeatureCollection(
  data: Array<{ h3Index: string; value: number }>,
) {
  return {
    type: "FeatureCollection" as const,
    features: data.map(({ h3Index, value }) => ({
      ...h3ToGeoJSON(h3Index),
      properties: {
        h3Index,
        value,
      },
    })),
  };
}

/**
 * Calculate color based on value and min/max range
 */
export function getColorForValue(
  value: number,
  min: number,
  max: number,
  colorScheme: { start: string; end: string } = {
    start: "#f7fbff",
    end: "#08519c",
  },
): string {
  if (max === min) return colorScheme.end;
  
  const ratio = (value - min) / (max - min);
  
  // Simple linear interpolation for now
  // You can use chroma-js for more sophisticated color interpolation
  return `rgba(8, 81, 156, ${0.2 + ratio * 0.8})`;
}

