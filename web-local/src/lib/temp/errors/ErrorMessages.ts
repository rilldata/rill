import type { V1ReconcileError } from "@rilldata/web-common/runtime-client";

export const MetricsSourceSelectionError = (
  errors: Array<V1ReconcileError>
): string => {
  return (
    errors?.find((error) => error.propertyPath.length === 0)?.message ?? ""
  );
};
