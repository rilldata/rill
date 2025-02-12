import type { KPIGridSpec } from ".";

export function validateKPIGridSchema(spec: KPIGridSpec) {
  if (typeof spec.metrics_view !== "string" || !spec.metrics_view) {
    return {
      isValid: false,
      error: "A metrics view must be specified",
    };
  }

  if (!Array.isArray(spec.measures) || spec.measures.length === 0) {
    return {
      isValid: false,
      error: "At least one measure must be specified",
    };
  }

  return {
    isValid: true,
  };
}
