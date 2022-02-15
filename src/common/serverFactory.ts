import {DuckDBClient} from "$common/database-service/DuckDBClient";
import {DatabaseDataLoaderActions} from "$common/database-service/DatabaseDataLoaderActions";
import {DatabaseTableActions} from "$common/database-service/DatabaseTableActions";
import {DatabaseColumnActions} from "$common/database-service/DatabaseColumnActions";
import {DataModelerStateService} from "$common/data-modeler-state-service/DataModelerStateService";
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

export function databaseServiceFactory(config: RootConfig) {
    const duckDbClient = new DuckDBClient(config.database);
    const databaseDataLoaderActions = new DatabaseDataLoaderActions(config.database, duckDbClient);
    const databaseTableActions = new DatabaseTableActions(config.database, duckDbClient);
    const databaseColumnActions = new DatabaseColumnActions(config.database, duckDbClient);
    return new DatabaseService(duckDbClient,
        [databaseDataLoaderActions, databaseTableActions, databaseColumnActions]);
}

export function dataModelerStateServiceFactory() {
    const tableStateActions = new TableStateActions();
    const modelStateActions = new ModelStateActions();
    const profileColumnStateActions = new ProfileColumnStateActions();
    return new DataModelerStateService(
        [tableStateActions, modelStateActions, profileColumnStateActions]);
}

export function dataModelerServiceFactory(config: RootConfig) {
    const databaseService = databaseServiceFactory(config);

    const dataModelerStateService = dataModelerStateServiceFactory();

    const notificationService = new SocketNotificationService();

    const tableActions = new TableActions(dataModelerStateService, databaseService);
    const modelActions = new ModelActions(dataModelerStateService, databaseService);
    const profileColumnActions = new ProfileColumnActions(dataModelerStateService, databaseService);
    const dataModelerService = new DataModelerService(dataModelerStateService, databaseService, notificationService,
        [tableActions, modelActions, profileColumnActions]);

    return {dataModelerStateService, dataModelerService, notificationService};
}

export function serverFactory(config: RootConfig) {
    const {dataModelerStateService, dataModelerService, notificationService} = dataModelerServiceFactory(config);

    const socketServer = new SocketServer(dataModelerService, dataModelerStateService, config);
    notificationService.setSocketServer(socketServer.getSocketServer());

    return {dataModelerStateService, dataModelerService, socketServer};
}
