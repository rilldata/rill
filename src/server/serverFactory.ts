import { DuckDBClient } from "$common/database-service/DuckDBClient";
import { DatabaseDataLoaderActions } from "$common/database-service/DatabaseDataLoaderActions";
import { DatabaseTableActions } from "$common/database-service/DatabaseTableActions";
import { DatabaseColumnActions } from "$common/database-service/DatabaseColumnActions";
import { TableStateActions } from "$common/data-modeler-state-service/TableStateActions";
import { ModelStateActions } from "$common/data-modeler-state-service/ModelStateActions";
import { TableActions } from "$common/data-modeler-service/TableActions";
import { ModelActions } from "$common/data-modeler-service/ModelActions";
import { ProfileColumnStateActions } from "$common/data-modeler-state-service/ProfileColumnStateActions";
import { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import { ProfileColumnActions } from "$common/data-modeler-service/ProfileColumnActions";
import { SocketServer } from "$server/SocketServer";
import { DatabaseService } from "$common/database-service/DatabaseService";
import type { RootConfig } from "$common/config/RootConfig";
import { SocketNotificationService } from "$common/socket/SocketNotificationService";
import { PersistentTableEntityService } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { DerivedTableEntityService } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import { PersistentModelEntityService } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import { DerivedModelEntityService } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import { CommonStateActions } from "$common/data-modeler-state-service/CommonStateActions";
import { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import { ApplicationStateService } from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import { ApplicationActions } from "$common/data-modeler-service/ApplicationActions";
import { ApplicationStateActions } from "$common/data-modeler-state-service/ApplicationStateActions";
import { ProductHealthEventFactory } from "$common/metrics-service/ProductHealthEventFactory";
import { MetricsService } from "$common/metrics-service/MetricsService";
import { RillIntakeClient } from "$common/metrics-service/RillIntakeClient";
import { existsSync, readFileSync } from "fs";
import { LocalConfigFile } from "$common/config/ConfigFolders";
import { MetricsDefinitionStateActions } from "$common/data-modeler-state-service/MetricsDefinitionStateActions";
import { MetricsDefinitionStateService } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { RillDeveloper } from "$server/RillDeveloper";
import { RillDeveloperService } from "$common/rill-developer-service/RillDeveloperService";
import { MetricsDefinitionActions } from "$common/rill-developer-service/MetricsDefinitionActions";
import { MeasuresActions } from "$common/rill-developer-service/MeasuresActions";
import { DimensionsActions } from "$common/rill-developer-service/DimensionsActions";
import { DatabaseMetricsExploreActions } from "$common/database-service/DatabaseMetricsExploreActions";
import { MetricsExploreActions } from "$common/rill-developer-service/MetricsExploreActions";
import { MeasureDefinitionStateService } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { DimensionDefinitionStateService } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { DatabaseTimeSeriesActions } from "$common/database-service/DatabaseTimeSeriesActions";
import { ExpressServer } from "$server/ExpressServer";

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
      DatabaseMetricsExploreActions,
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
  const productHealthEventFactory = new ProductHealthEventFactory(config);

  return new MetricsService(
    config,
    dataModelerStateService,
    new RillIntakeClient(config),
    [productHealthEventFactory]
  );
}

export function dataModelerServiceFactory(config: RootConfig) {
  if (existsSync(LocalConfigFile)) {
    config.local = JSON.parse(readFileSync(LocalConfigFile).toString());
  }
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
      MetricsExploreActions,
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
