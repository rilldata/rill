import { DuckDBClient } from "$common/database-service/DuckDBClient";
import { DatabaseDataLoaderActions } from "$common/database-service/DatabaseDataLoaderActions";
import { DatabaseTableActions } from "$common/database-service/DatabaseTableActions";
import { DatabaseColumnActions } from "$common/database-service/DatabaseColumnActions";
import { SourceStateActions } from "$common/data-modeler-state-service/SourceStateActions";
import { ModelStateActions } from "$common/data-modeler-state-service/ModelStateActions";
import { SourceActions } from "$common/data-modeler-service/SourceActions";
import { ModelActions } from "$common/data-modeler-service/ModelActions";
import { ProfileColumnStateActions } from "$common/data-modeler-state-service/ProfileColumnStateActions";
import { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import { ProfileColumnActions } from "$common/data-modeler-service/ProfileColumnActions";
import { SocketServer } from "$server/SocketServer";
import { DatabaseService } from "$common/database-service/DatabaseService";
import type { RootConfig } from "$common/config/RootConfig";
import { SocketNotificationService } from "$common/socket/SocketNotificationService";
import { PersistentSourceEntityService } from "$common/data-modeler-state-service/entity-state-service/PersistentSourceEntityService";
import { DerivedSourceEntityService } from "$common/data-modeler-state-service/entity-state-service/DerivedSourceEntityService";
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

let PACKAGE_JSON = "";
try {
  PACKAGE_JSON = __dirname + "/../../package.json";
} catch (err) {
  PACKAGE_JSON = "package.json";
}

export function databaseServiceFactory(config: RootConfig) {
  const duckDbClient = DuckDBClient.getInstance(config.database);
  const databaseDataLoaderActions = new DatabaseDataLoaderActions(
    config.database,
    duckDbClient
  );
  const databaseTableActions = new DatabaseTableActions(
    config.database,
    duckDbClient
  );
  const databaseColumnActions = new DatabaseColumnActions(
    config.database,
    duckDbClient
  );
  return new DatabaseService(duckDbClient, [
    databaseDataLoaderActions,
    databaseTableActions,
    databaseColumnActions,
  ]);
}

export function dataModelerStateServiceFactory(config: RootConfig) {
  return new DataModelerStateService(
    [
      SourceStateActions,
      ModelStateActions,
      ProfileColumnStateActions,
      CommonStateActions,
      ApplicationStateActions,
    ].map((StateActionsClass) => new StateActionsClass()),
    [
      PersistentSourceEntityService,
      DerivedSourceEntityService,
      PersistentModelEntityService,
      DerivedModelEntityService,
      ApplicationStateService,
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

  const sourceActions = new SourceActions(
    config,
    dataModelerStateService,
    databaseService
  );
  const modelActions = new ModelActions(
    config,
    dataModelerStateService,
    databaseService
  );
  const profileColumnActions = new ProfileColumnActions(
    config,
    dataModelerStateService,
    databaseService
  );
  const applicationActions = new ApplicationActions(
    config,
    dataModelerStateService,
    databaseService
  );
  const dataModelerService = new DataModelerService(
    dataModelerStateService,
    databaseService,
    notificationService,
    metricsService,
    [sourceActions, modelActions, profileColumnActions, applicationActions]
  );

  return {
    dataModelerStateService,
    dataModelerService,
    notificationService,
    metricsService,
  };
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
