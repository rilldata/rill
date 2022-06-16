import { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import { DataModelerSocketService } from "$common/socket/DataModelerSocketService";
import type { RootConfig } from "$common/config/RootConfig";
import { PersistentSourceEntityService } from "$common/data-modeler-state-service/entity-state-service/PersistentSourceEntityService";
import { DerivedSourceEntityService } from "$common/data-modeler-state-service/entity-state-service/DerivedSourceEntityService";
import { PersistentModelEntityService } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import { DerivedModelEntityService } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import { ApplicationStateService } from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type { MetricsService } from "$common/metrics-service/MetricsService";
import { MetricsSocketService } from "$common/socket/MetricsSocketService";

export function dataModelerStateServiceClientFactory() {
  return new DataModelerStateService(
    [],
    [
      PersistentSourceEntityService,
      DerivedSourceEntityService,
      PersistentModelEntityService,
      DerivedModelEntityService,
      ApplicationStateService,
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
