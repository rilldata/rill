import {DuckDBClient} from "$common/database-service/DuckDBClient";
import {DatabaseDataLoaderActions} from "$common/database-service/DatabaseDataLoaderActions";
import {DatabaseTableActions} from "$common/database-service/DatabaseTableActions";
import {DatabaseColumnActions} from "$common/database-service/DatabaseColumnActions";
import {TableStateActions} from "$common/data-modeler-state-service/TableStateActions";
import {ModelStateActions} from "$common/data-modeler-state-service/ModelStateActions";
import {TableActions} from "$common/data-modeler-service/TableActions";
import {ModelActions} from "$common/data-modeler-service/ModelActions";
import {ProfileColumnStateActions} from "$common/data-modeler-state-service/ProfileColumnStateActions";
import {DataModelerService} from "$common/data-modeler-service/DataModelerService";
import {ProfileColumnActions} from "$common/data-modeler-service/ProfileColumnActions";
import {SocketServer} from "$common/socket/SocketServer";
import {DatabaseService} from "$common/database-service/DatabaseService";
import type {RootConfig} from "$common/config/RootConfig";
import { SocketNotificationService } from "$common/socket/SocketNotificationService";
import {
    PersistentTableEntityService
} from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import {
    DerivedTableEntityService
} from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import {
    PersistentModelEntityService
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import {
    DerivedModelEntityService
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import { CommonStateActions } from "$common/data-modeler-state-service/CommonStateActions";
import { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import {
    ApplicationStateService
} from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import { ApplicationActions } from "$common/data-modeler-service/ApplicationActions";
import { ApplicationStateActions } from "$common/data-modeler-state-service/ApplicationStateActions";

export function databaseServiceFactory(config: RootConfig) {
    const duckDbClient = new DuckDBClient(config.database);
    const databaseDataLoaderActions = new DatabaseDataLoaderActions(config.database, duckDbClient);
    const databaseTableActions = new DatabaseTableActions(config.database, duckDbClient);
    const databaseColumnActions = new DatabaseColumnActions(config.database, duckDbClient);
    return new DatabaseService(duckDbClient,
        [databaseDataLoaderActions, databaseTableActions, databaseColumnActions]);
}

export function dataModelerStateServiceFactory(config: RootConfig) {
    return new DataModelerStateService(
        [
            TableStateActions, ModelStateActions,
            ProfileColumnStateActions, CommonStateActions,
            ApplicationStateActions,
        ].map(StateActionsClass => new StateActionsClass()),
        [
            PersistentTableEntityService, DerivedTableEntityService,
            PersistentModelEntityService, DerivedModelEntityService,
            ApplicationStateService,
        ].map(EntityStateService => new EntityStateService()),
        config);
}

export function dataModelerServiceFactory(config: RootConfig) {
    const databaseService = databaseServiceFactory(config);

    const dataModelerStateService = dataModelerStateServiceFactory(config);

    const notificationService = new SocketNotificationService();

    const tableActions = new TableActions(config, dataModelerStateService, databaseService);
    const modelActions = new ModelActions(config, dataModelerStateService, databaseService);
    const profileColumnActions = new ProfileColumnActions(config, dataModelerStateService, databaseService);
    const applicationActions = new ApplicationActions(config, dataModelerStateService, databaseService);
    const dataModelerService = new DataModelerService(dataModelerStateService, databaseService, notificationService,
        [tableActions, modelActions, profileColumnActions, applicationActions]);

    return {dataModelerStateService, dataModelerService, notificationService};
}

export function serverFactory(config: RootConfig) {
    const {dataModelerStateService, dataModelerService, notificationService} = dataModelerServiceFactory(config);

    const socketServer = new SocketServer(config, dataModelerService, dataModelerStateService);
    notificationService.setSocketServer(socketServer.getSocketServer());

    return {dataModelerStateService, dataModelerService, socketServer};
}
