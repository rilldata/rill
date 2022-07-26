import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { SourceModelValidationStatus } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

// Unified location for error messages
// TODO: move all errors here.

export const ExplorerSourceModelDoesntExist =
  "Selected source model for the active metrics no longer exists.";
export const ExplorerSourceModelIsInvalid =
  "Source model query has errors. Please fix those before using it for metrics.";
export const ExplorerTimeDimensionDoesntExist =
  "Selected time dimension for the active metrics no longer exists.";

export const MetricsSourceSelectionError = (
  metricsDefinition: MetricsDefinitionEntity
): string => {
  if (
    metricsDefinition.sourceModelValidationStatus !==
    SourceModelValidationStatus.OK
  ) {
    switch (metricsDefinition.sourceModelValidationStatus) {
      case SourceModelValidationStatus.EMPTY:
        return ""; // nothing as of now
      case SourceModelValidationStatus.INVALID:
        return ExplorerSourceModelIsInvalid;
      case SourceModelValidationStatus.MISSING:
        return ExplorerSourceModelDoesntExist;
    }
  }

  if (
    metricsDefinition.timeDimensionValidationStatus !==
    SourceModelValidationStatus.OK
  ) {
    switch (metricsDefinition.timeDimensionValidationStatus) {
      case SourceModelValidationStatus.EMPTY:
        return ""; // nothing as of now
      case SourceModelValidationStatus.INVALID:
        return ExplorerSourceModelIsInvalid;
      case SourceModelValidationStatus.MISSING:
        return ExplorerTimeDimensionDoesntExist;
    }
  }

  return "";
};
