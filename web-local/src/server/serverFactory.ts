import type { RootConfig } from "@rilldata/web-local/common/config/RootConfig";
import { ApplicationActions } from "@rilldata/web-local/common/data-modeler-service/ApplicationActions";
import { DataModelerService } from "@rilldata/web-local/common/data-modeler-service/DataModelerService";
import { ModelActions } from "@rilldata/web-local/common/data-modeler-service/ModelActions";
import { ProfileColumnActions } from "@rilldata/web-local/common/data-modeler-service/ProfileColumnActions";
import { TableActions } from "@rilldata/web-local/common/data-modeler-service/TableActions";
import { ApplicationStateActions } from "@rilldata/web-local/common/data-modeler-state-service/ApplicationStateActions";
import { CommonStateActions } from "@rilldata/web-local/common/data-modeler-state-service/CommonStateActions";
import { DataModelerStateService } from "@rilldata/web-local/common/data-modeler-state-service/DataModelerStateService";
import { ApplicationStateService } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import { DerivedModelEntityService } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import { DerivedTableEntityService } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import { DimensionDefinitionStateService } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { MeasureDefinitionStateService } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { MetricsDefinitionStateService } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { PersistentModelEntityService } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import { PersistentTableEntityService } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { MetricsDefinitionStateActions } from "@rilldata/web-local/common/data-modeler-state-service/MetricsDefinitionStateActions";
import { ModelStateActions } from "@rilldata/web-local/common/data-modeler-state-service/ModelStateActions";
import { ProfileColumnStateActions } from "@rilldata/web-local/common/data-modeler-state-service/ProfileColumnStateActions";
import { TableStateActions } from "@rilldata/web-local/common/data-modeler-state-service/TableStateActions";
import { DatabaseColumnActions } from "@rilldata/web-local/common/database-service/DatabaseColumnActions";
import { DatabaseDataLoaderActions } from "@rilldata/web-local/common/database-service/DatabaseDataLoaderActions";
import { DatabaseMetricsExplorerActions } from "@rilldata/web-local/common/database-service/DatabaseMetricsExplorerActions";
import { DatabaseService } from "@rilldata/web-local/common/database-service/DatabaseService";
import { DatabaseTableActions } from "@rilldata/web-local/common/database-service/DatabaseTableActions";
import { DatabaseTimeSeriesActions } from "@rilldata/web-local/common/database-service/DatabaseTimeSeriesActions";
import { DuckDBClient } from "@rilldata/web-local/common/database-service/DuckDBClient";
import { BehaviourEventFactory } from "@rilldata/web-local/common/metrics-service/BehaviourEventFactory";
import { MetricsService } from "@rilldata/web-local/common/metrics-service/MetricsService";
import { ProductHealthEventFactory } from "@rilldata/web-local/common/metrics-service/ProductHealthEventFactory";
import { RillIntakeClient } from "@rilldata/web-local/common/metrics-service/RillIntakeClient";
import { DimensionsActions } from "@rilldata/web-local/common/rill-developer-service/DimensionsActions";
import { MeasuresActions } from "@rilldata/web-local/common/rill-developer-service/MeasuresActions";
import { MetricsDefinitionActions } from "@rilldata/web-local/common/rill-developer-service/MetricsDefinitionActions";
import { MetricsViewActions } from "@rilldata/web-local/common/rill-developer-service/MetricsViewActions";
import { RillDeveloperService } from "@rilldata/web-local/common/rill-developer-service/RillDeveloperService";
import { SocketNotificationService } from "@rilldata/web-local/common/socket/SocketNotificationService";
import { initLocalConfig } from "@rilldata/web-local/common/utils/initLocalConfig";
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
  const duckDbClient = DuckDBClient.getInstance(config);
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

export function metricsServiceFactory(config: RootConfig) {
  return new MetricsService(config, new RillIntakeClient(config), [
    new ProductHealthEventFactory(config),
    new BehaviourEventFactory(config),
  ]);
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

  const metricsService = metricsServiceFactory(config);

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
