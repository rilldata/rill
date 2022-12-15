import type { V1ReconcileError } from "@rilldata/web-common/runtime-client";

// Unified location for error messages
// TODO: move all errors here.

export const ExplorerMetricsDefinitionDoesntExist =
  "Metrics definition doesn't exist";
export const ExplorerSourceModelDoesntExist =
  "Previously selected source model does not exist.";
export const ExplorerSourceModelIsInvalid = "Model query has errors.";
export const ExplorerTimeDimensionDoesntExist =
  "Previously selected timestamp column does not exist.";
export const ExplorerSourceColumnDoesntExist = "not found in FROM clause!"; // the full DuckDB error message is `Binder Error: Referenced column "COLUMN_NAME" not found in FROM clause!`

export const MetricsSourceSelectionError = (
  errors: Array<V1ReconcileError>
): string => {
  return (
    errors?.find((error) => error.propertyPath.length === 0)?.message ?? ""
  );

  // if (
  //   metricsDefinition.sourceModelValidationStatus !==
  //   SourceModelValidationStatus.OK
  // ) {
  //   switch (metricsDefinition.sourceModelValidationStatus) {
  //     case SourceModelValidationStatus.EMPTY:
  //       return ""; // nothing as of now
  //     case SourceModelValidationStatus.INVALID:
  //       return ExplorerSourceModelIsInvalid;
  //     case SourceModelValidationStatus.MISSING:
  //       return ExplorerSourceModelDoesntExist;
  //   }
  // }
  //
  // if (
  //   metricsDefinition.timeDimensionValidationStatus !==
  //   SourceModelValidationStatus.OK
  // ) {
  //   switch (metricsDefinition.timeDimensionValidationStatus) {
  //     case SourceModelValidationStatus.EMPTY:
  //       return ""; // nothing as of now
  //     case SourceModelValidationStatus.INVALID:
  //       return ExplorerSourceModelIsInvalid;
  //     case SourceModelValidationStatus.MISSING:
  //       return ExplorerTimeDimensionDoesntExist;
  //   }
  // }

  return "";
};
