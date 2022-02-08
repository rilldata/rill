import {DuckDBClient} from "$common/database-service/DuckDBClient";
import {DatabaseDataLoaderActions} from "$common/database-service/DatabaseDataLoaderActions";
import {DatabaseTableActions} from "$common/database-service/DatabaseTableActions";
import {DatabaseColumnActions} from "$common/database-service/DatabaseColumnActions";
import {DataModelerStateService} from "$common/data-modeler-state-service/DataModelerStateService";
import {DatasetStateActions} from "$common/data-modeler-state-service/DatasetStateActions";
import {ModelStateActions} from "$common/data-modeler-state-service/ModelStateActions";
import {DatasetActions} from "$common/data-modeler-service/DatasetActions";
import {ModelActions} from "$common/data-modeler-service/ModelActions";
import {ProfileColumnStateActions} from "$common/data-modeler-state-service/ProfileColumnStateActions";
import {DataModelerService} from "$common/data-modeler-service/DataModelerService";
import {ProfileColumnActions} from "$common/data-modeler-service/ProfileColumnActions";
import {SocketServer} from "$common/SocketServer";
import {DatabaseService} from "$common/database-service/DatabaseService";

export function databaseServiceFactory() {
    const duckDbClient = new DuckDBClient();
    const databaseDataLoaderActions = new DatabaseDataLoaderActions(duckDbClient);
    const databaseTableActions = new DatabaseTableActions(duckDbClient);
    const databaseColumnActions = new DatabaseColumnActions(duckDbClient);
    return new DatabaseService(duckDbClient,
        [databaseDataLoaderActions, databaseTableActions, databaseColumnActions]);
}

export function dataModelerStateServiceFactory() {
    const datasetStateActions = new DatasetStateActions();
    const modelStateActions = new ModelStateActions();
    const profileColumnStateActions = new ProfileColumnStateActions();
    return new DataModelerStateService(
        [datasetStateActions, modelStateActions, profileColumnStateActions]);
}

export function dataModelerServiceFactory() {
    const databaseService = databaseServiceFactory();

    const dataModelerStateService = dataModelerStateServiceFactory()

    const datasetActions = new DatasetActions(dataModelerStateService, databaseService);
    const modelActions = new ModelActions(dataModelerStateService, databaseService);
    const profileColumnActions = new ProfileColumnActions(dataModelerStateService, databaseService);
    const dataModelerService = new DataModelerService(dataModelerStateService, databaseService,
        [datasetActions, modelActions, profileColumnActions]);

    return {dataModelerStateService, dataModelerService};
}

export function serverFactory() {
    const {dataModelerStateService, dataModelerService} = dataModelerServiceFactory();

    const socketServer = new SocketServer(dataModelerService, dataModelerStateService,
        "http://localhost:3000", 3001);

    return {dataModelerStateService, dataModelerService, socketServer};
}
