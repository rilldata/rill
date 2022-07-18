import { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import { DataModelerSocketService } from "$common/socket/DataModelerSocketService";
import type { RootConfig } from "$common/config/RootConfig";
import { PersistentTableEntityService } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { DerivedTableEntityService } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import { PersistentModelEntityService } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import { DerivedModelEntityService } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import { ApplicationStateService } from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type { MetricsService } from "$common/metrics-service/MetricsService";
import { MetricsSocketService } from "$common/socket/MetricsSocketService";
import { MetricsDefinitionStateService } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { MeasureDefinitionStateService } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { DimensionDefinitionStateService } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";

export function dataModelerStateServiceClientFactory() {
  return new DataModelerStateService(
    [],
    [
      PersistentTableEntityService,
      DerivedTableEntityService,
      PersistentModelEntityService,
      DerivedModelEntityService,
      ApplicationStateService,
      MetricsDefinitionStateService,
      MeasureDefinitionStateService,
      DimensionDefinitionStateService,
    ].map((EntityStateService) => new EntityStateService())
  );
}

export function clientFactory(config: RootConfig): {
  dataModelerStateService: DataModelerStateService;
  metricsService: MetricsService;
  dataModelerService: DataModelerService;
} {
  const dataModelerStateService = dataModelerStateServiceClientFactory();
  const metricsService = new MetricsSocketService(config);
  const dataModelerService = new DataModelerSocketService(
    dataModelerStateService,
    metricsService,
    config.server
  );

  return { dataModelerStateService, metricsService, dataModelerService };
}
