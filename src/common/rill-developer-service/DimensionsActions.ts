import { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";
import type { MetricsDefinitionContext } from "$common/rill-developer-service/MetricsDefinitionActions";
import type { DimensionDefinition } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

/**
 * select
 * count(*), date_trunc('HOUR', created_date) as inter
 * from nyc311_reduced
 * group by date_trunc('HOUR', created_date) order by inter;
 */

export class DimensionsActions extends RillDeveloperActions {
  public async addNewDimension(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    columnName: string
  ) {
    const dimensions: DimensionDefinition = {
      dimensionColumn: columnName,
      id: "",
      dimensionIsValid: ValidationState.OK,
      sqlNameIsValid: ValidationState.OK,
    };

    this.dataModelerStateService.dispatch("addNewDimension", [
      metricsDefId,
      dimensions,
    ]);
    rillRequestContext.actionsChannel.pushMessage("addNewDimension", [
      metricsDefId,
      dimensions,
    ]);
  }
}
