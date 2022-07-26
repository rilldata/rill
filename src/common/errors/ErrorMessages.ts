import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { SourceModelValidationStatus } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

// Unified location for error messages
// TODO: move all errors here.

export const ExplorerSourceModelDoesntExist =
  "Previously selected source model does not exist.";
export const ExplorerSourceModelIsInvalid = "Model query has errors.";
export const ExplorerTimeDimensionDoesntExist =
  "Missing a valid timestamp column.";

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
