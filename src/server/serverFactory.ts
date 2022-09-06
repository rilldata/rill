import type { RootConfig } from "$common/config/RootConfig";
import { ApplicationActions } from "$common/data-modeler-service/ApplicationActions";
import { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import { ModelActions } from "$common/data-modeler-service/ModelActions";
import { ProfileColumnActions } from "$common/data-modeler-service/ProfileColumnActions";
import { TableActions } from "$common/data-modeler-service/TableActions";
import { ApplicationStateActions } from "$common/data-modeler-state-service/ApplicationStateActions";
import { CommonStateActions } from "$common/data-modeler-state-service/CommonStateActions";
import { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import { ApplicationStateService } from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import { DerivedModelEntityService } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import { DerivedTableEntityService } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import { DimensionDefinitionStateService } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { MeasureDefinitionStateService } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { MetricsDefinitionStateService } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { PersistentModelEntityService } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import { PersistentTableEntityService } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { MetricsDefinitionStateActions } from "$common/data-modeler-state-service/MetricsDefinitionStateActions";
import { ModelStateActions } from "$common/data-modeler-state-service/ModelStateActions";
import { ProfileColumnStateActions } from "$common/data-modeler-state-service/ProfileColumnStateActions";
import { TableStateActions } from "$common/data-modeler-state-service/TableStateActions";
import { DatabaseColumnActions } from "$common/database-service/DatabaseColumnActions";
import { DatabaseDataLoaderActions } from "$common/database-service/DatabaseDataLoaderActions";
import { DatabaseMetricsExplorerActions } from "$common/database-service/DatabaseMetricsExplorerActions";
import { DatabaseService } from "$common/database-service/DatabaseService";
import { DatabaseTableActions } from "$common/database-service/DatabaseTableActions";
import { DatabaseTimeSeriesActions } from "$common/database-service/DatabaseTimeSeriesActions";
import { DuckDBClient } from "$common/database-service/DuckDBClient";
import { BehaviourEventFactory } from "$common/metrics-service/BehaviourEventFactory";
import { MetricsService } from "$common/metrics-service/MetricsService";
import { ProductHealthEventFactory } from "$common/metrics-service/ProductHealthEventFactory";
import { RillIntakeClient } from "$common/metrics-service/RillIntakeClient";
import { DimensionsActions } from "$common/rill-developer-service/DimensionsActions";
import { MeasuresActions } from "$common/rill-developer-service/MeasuresActions";
import { MetricsDefinitionActions } from "$common/rill-developer-service/MetricsDefinitionActions";
import { MetricsViewActions } from "$common/rill-developer-service/MetricsViewActions";
import { RillDeveloperService } from "$common/rill-developer-service/RillDeveloperService";
import { SocketNotificationService } from "$common/socket/SocketNotificationService";
import { initLocalConfig } from "$common/utils/initLocalConfig";
import { ExpressServer } from "$server/ExpressServer";
import type { RillDeveloper } from "$server/RillDeveloper";
import { SocketServer } from "$server/SocketServer";
import { readFileSync } from "fs";

let PACKAGE_JSON = "";
try {
  PACKAGE_JSON = __dirname + "/../../package.json";
} catch (err) {
  PACKAGE_JSON = "package.json";
}

export function databaseServiceFactory(config: RootConfig) {
  const duckDbClient = DuckDBClient.getInstance(config.database);
  return new DatabaseService(
    duckDbClient,
    [
      DatabaseDataLoaderActions,
      DatabaseTableActions,
      DatabaseColumnActions,
      DatabaseMetricsExplorerActions,
      DatabaseTimeSeriesActions,
    ].map(
      (DatabaseActionsClass) =>
        new DatabaseActionsClass(config.database, duckDbClient)
    )
  );
}

export function dataModelerStateServiceFactory(config: RootConfig) {
  return new DataModelerStateService(
    [
      TableStateActions,
      ModelStateActions,
      ProfileColumnStateActions,
      CommonStateActions,
      ApplicationStateActions,
      MetricsDefinitionStateActions,
    ].map((StateActionsClass) => new StateActionsClass()),
    [
      PersistentTableEntityService,
      DerivedTableEntityService,
      PersistentModelEntityService,
      DerivedModelEntityService,
      ApplicationStateService,
      MetricsDefinitionStateService,
      MeasureDefinitionStateService,
      DimensionDefinitionStateService,
    ].map((EntityStateService) => new EntityStateService()),
    config
  );
}

export function metricsServiceFactory(
  config: RootConfig,
  dataModelerStateService: DataModelerStateService
) {
  return new MetricsService(
    config,
    dataModelerStateService,
    new RillIntakeClient(config),
    [new ProductHealthEventFactory(config), new BehaviourEventFactory(config)]
  );
}

export function dataModelerServiceFactory(config: RootConfig) {
  config.local = initLocalConfig(config.local);
  try {
    config.local.version = JSON.parse(
      readFileSync(PACKAGE_JSON).toString()
    ).version;
  } catch (err) {
    console.error(err);
  }

  const databaseService = databaseServiceFactory(config);

  const dataModelerStateService = dataModelerStateServiceFactory(config);

  const metricsService = metricsServiceFactory(config, dataModelerStateService);

  const notificationService = new SocketNotificationService();

  const dataModelerService = new DataModelerService(
    dataModelerStateService,
    databaseService,
    notificationService,
    metricsService,
    [TableActions, ModelActions, ProfileColumnActions, ApplicationActions].map(
      (DataModelerActionsClass) =>
        new DataModelerActionsClass(
          config,
          dataModelerStateService,
          databaseService
        )
    )
  );

  return {
    dataModelerStateService,
    dataModelerService,
    notificationService,
    metricsService,
  };
}

export function rillDeveloperServiceFactory(rillDeveloper: RillDeveloper) {
  return new RillDeveloperService(
    rillDeveloper.dataModelerStateService,
    rillDeveloper.dataModelerService,
    rillDeveloper.dataModelerService.getDatabaseService(),
    [
      MetricsDefinitionActions,
      DimensionsActions,
      MeasuresActions,
      MetricsViewActions,
    ].map(
      (RillDeveloperActionsClass) =>
        new RillDeveloperActionsClass(
          rillDeveloper.config,
          rillDeveloper.dataModelerStateService,
          rillDeveloper.dataModelerService.getDatabaseService()
        )
    )
  );
}

export function serverFactory(config: RootConfig) {
  const {
    dataModelerStateService,
    dataModelerService,
    notificationService,
    metricsService,
  } = dataModelerServiceFactory(config);

  const socketServer = new SocketServer(
    config,
    dataModelerService,
    dataModelerStateService,
    metricsService
  );
  notificationService.setSocketServer(socketServer.getSocketServer());

  return { dataModelerStateService, dataModelerService, socketServer };
}

export function expressServerFactory(
  config: RootConfig,
  rillDeveloper: RillDeveloper,
  rillDeveloperService: RillDeveloperService
) {
  return new ExpressServer(
    config,
    rillDeveloper.dataModelerService,
    rillDeveloperService,
    rillDeveloper.dataModelerStateService,
    rillDeveloper.notificationService as SocketNotificationService,
    rillDeveloper.metricsService
  );
}
