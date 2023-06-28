import { parseDocument } from "yaml";

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

export function runtimeErrorToLine(message: string, yaml: string) {
  const lines = yaml.split("\n");
  if (message === ConfigErrors.SouceNotSelected) {
    /** if this is undefined, then the field isn't here either. */
    const line = lines.findIndex((line) => line.startsWith("model: "));
    return { line: line + 1, message, level: "error" };
  }
  if (message.startsWith(ConfigErrors.InvalidTimeGrainForSmallest)) {
    const line = lines.findIndex((line) =>
      line.startsWith("smallest_time_grain:")
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

export function mapRuntimeErrorsToLines(errors, yaml: string) {
  if (!errors) return [];
  return errors
    .map((error) => {
      return runtimeErrorToLine(error.message, yaml);
    })
    .filter((error) => error.message !== ConfigErrors.Malformed);
}

export function getSyntaxErrors(yaml: string) {
  const doc = parseDocument(yaml);
  const syntaxErrors = doc.errors;
  if (syntaxErrors.length === 0) return [];
  return syntaxErrors.map((error) => {
    return {
      line: error.linePos[0].line,
      message: error.message,
      level: "error",
    };
  });
}
