import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
import type { V1ParseError } from "@rilldata/web-common/runtime-client";

export enum ValidationState {
  OK = "OK",
  WARNING = "WARNING",
  ERROR = "ERROR",
}

export enum SourceModelValidationStatus {
  OK = "OK",
  // No source model selected.
  EMPTY = "EMPTY",
  // Source model query is invalid.
  INVALID = "INVALID",
  // Selected source model is no longer present.
  MISSING = "MISSING",
}

/** a temporary set of enums that shoul be emitted by orval's codegen */
export enum ConfigErrors {
  SourceNotSelected = "metrics view source not selected",
  SourceNotFound = "metrics view source not found",
  SouceNotSelected = "metrics view source not selected",
  TimestampNotSelected = "metrics view timestamp not selected",
  TimestampNotFound = "metrics view selected timestamp not found",
  MissingDimension = "at least one dimension should be present",
  MissingMeasure = "at least one measure should be present",
  Malformed = "did not find expected key",
  InvalidTimeGrainForSmallest = "invalid time grain",
}

export function runtimeErrorToLine(message: string, yaml: string): LineStatus {
  const lines = yaml.split("\n");
  if (message === ConfigErrors.SouceNotSelected) {
    /** if this is undefined, then the field isn't here either. */
    const line = lines.findIndex((line) => line.startsWith("model: "));
    return { line: line + 1, message, level: "error" };
  }
  if (message.startsWith(ConfigErrors.InvalidTimeGrainForSmallest)) {
    const line = lines.findIndex((line) =>
      line.startsWith("smallest_time_grain:"),
    );
    return { line: line + 1, message, level: "error" };
  }
  if (message === ConfigErrors.TimestampNotFound) {
    const line = lines.findIndex((line) => line.startsWith("timeseries:")) + 1;
    return { line: line, message, level: "error" };
  }
  if (message === ConfigErrors.MissingMeasure) {
    const line = lines.findIndex((line) => line.startsWith("measures:"));
    return { line: line + 1, message, level: "error" };
  }
  if (message === ConfigErrors.MissingDimension) {
    const line = lines.findIndex((line) => line.startsWith("dimensions:"));
    return { line: line + 1, message, level: "error" };
  }
  if (message.startsWith("yaml: line")) {
    const line = parseInt(message.split("yaml: line ")[1].split(":")[0]);
    return { line: line, message, level: "error" };
  }
  return { line: null, message, level: "error" };
}

// TODO: double check error
export function mapParseErrorsToLines(
  errors: Array<V1ParseError>,
  yaml: string,
): LineStatus[] {
  if (!errors) return [];
  return errors
    .map((error) => {
      if (error.startLocation) {
        // if line is provided, no need to parse
        // TODO: check if we need to strip anything
        return {
          line: error.startLocation.line,
          message: error.message,
          level: "error",
        };
      }
      return runtimeErrorToLine(error.message, yaml);
    })
    .filter((error) => error.message !== ConfigErrors.Malformed);
}
