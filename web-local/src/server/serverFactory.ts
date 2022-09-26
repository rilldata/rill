import type { RootConfig } from "$web-local/common/config/RootConfig";
import { ApplicationActions } from "$web-local/common/data-modeler-service/ApplicationActions";
import { DataModelerService } from "$web-local/common/data-modeler-service/DataModelerService";
import { ModelActions } from "$web-local/common/data-modeler-service/ModelActions";
import { ProfileColumnActions } from "$web-local/common/data-modeler-service/ProfileColumnActions";
import { TableActions } from "$web-local/common/data-modeler-service/TableActions";
import { ApplicationStateActions } from "$web-local/common/data-modeler-state-service/ApplicationStateActions";
import { CommonStateActions } from "$web-local/common/data-modeler-state-service/CommonStateActions";
import { DataModelerStateService } from "$web-local/common/data-modeler-state-service/DataModelerStateService";
import { ApplicationStateService } from "$web-local/common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import { DerivedModelEntityService } from "$web-local/common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import { DerivedTableEntityService } from "$web-local/common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import { DimensionDefinitionStateService } from "$web-local/common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { MeasureDefinitionStateService } from "$web-local/common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { MetricsDefinitionStateService } from "$web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { PersistentModelEntityService } from "$web-local/common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import { PersistentTableEntityService } from "$web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { MetricsDefinitionStateActions } from "$web-local/common/data-modeler-state-service/MetricsDefinitionStateActions";
import { ModelStateActions } from "$web-local/common/data-modeler-state-service/ModelStateActions";
import { ProfileColumnStateActions } from "$web-local/common/data-modeler-state-service/ProfileColumnStateActions";
import { TableStateActions } from "$web-local/common/data-modeler-state-service/TableStateActions";
import { DatabaseColumnActions } from "$web-local/common/database-service/DatabaseColumnActions";
import { DatabaseDataLoaderActions } from "$web-local/common/database-service/DatabaseDataLoaderActions";
import { DatabaseMetricsExplorerActions } from "$web-local/common/database-service/DatabaseMetricsExplorerActions";
import { DatabaseService } from "$web-local/common/database-service/DatabaseService";
import { DatabaseTableActions } from "$web-local/common/database-service/DatabaseTableActions";
import { DatabaseTimeSeriesActions } from "$web-local/common/database-service/DatabaseTimeSeriesActions";
import { DuckDBClient } from "$web-local/common/database-service/DuckDBClient";
import { BehaviourEventFactory } from "$web-local/common/metrics-service/BehaviourEventFactory";
import { MetricsService } from "$web-local/common/metrics-service/MetricsService";
import { ProductHealthEventFactory } from "$web-local/common/metrics-service/ProductHealthEventFactory";
import { RillIntakeClient } from "$web-local/common/metrics-service/RillIntakeClient";
import { DimensionsActions } from "$web-local/common/rill-developer-service/DimensionsActions";
import { MeasuresActions } from "$web-local/common/rill-developer-service/MeasuresActions";
import { MetricsDefinitionActions } from "$web-local/common/rill-developer-service/MetricsDefinitionActions";
import { MetricsViewActions } from "$web-local/common/rill-developer-service/MetricsViewActions";
import { RillDeveloperService } from "$web-local/common/rill-developer-service/RillDeveloperService";
import { SocketNotificationService } from "$web-local/common/socket/SocketNotificationService";
import { initLocalConfig } from "$web-local/common/utils/initLocalConfig";
import { ExpressServer } from "./ExpressServer";
import type { RillDeveloper } from "./RillDeveloper";
import { SocketServer } from "./SocketServer";
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
