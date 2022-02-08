import {DuckDBClient} from "$common/database/DuckDBClient";
import {DatabaseDataLoaderActions} from "$common/database/DatabaseDataLoaderActions";
import {DatabaseTableActions} from "$common/database/DatabaseTableActions";
import {DatabaseColumnActions} from "$common/database/DatabaseColumnActions";
import {DataModelerStateManager} from "$common/state-actions/DataModelerStateManager";
import {DatasetStateActions} from "$common/state-actions/DatasetStateActions";
import {ModelStateActions} from "$common/state-actions/ModelStateActions";
import {DatasetActions} from "$common/data-modeler-actions/DatasetActions";
import {ModelActions} from "$common/data-modeler-actions/ModelActions";
import {ProfileColumnStateActions} from "$common/state-actions/ProfileColumnStateActions";
import {DataModelerActionAPI} from "$common/data-modeler-actions/DataModelerActionAPI";
import {ProfileColumnActions} from "$common/data-modeler-actions/ProfileColumnActions";
import {SocketServer} from "$common/SocketServer";

export function serverFactory(): {
    dataModelerStateManager: DataModelerStateManager,
    dataModelerActionAPI: DataModelerActionAPI,
    socketServer: SocketServer,
} {
    const duckDbClient = new DuckDBClient();
    const duckDBDataLoaderAPI = new DatabaseDataLoaderActions(duckDbClient);
    const duckDbTableAPI = new DatabaseTableActions(duckDbClient);
    const duckDBColumnAPI = new DatabaseColumnActions(duckDbClient);

    const datasetStateActions = new DatasetStateActions();
    const modelStateActions = new ModelStateActions();
    const profileColumnStateActions = new ProfileColumnStateActions();
    const dataModelerStateManager = new DataModelerStateManager(
        [datasetStateActions, modelStateActions, profileColumnStateActions]);

    const datasetActions = new DatasetActions(
        dataModelerStateManager, duckDbTableAPI, duckDBColumnAPI, duckDBDataLoaderAPI);
    const modelActions = new ModelActions(
        dataModelerStateManager, duckDbTableAPI, duckDBColumnAPI, duckDBDataLoaderAPI);
    const profileColumnActions = new ProfileColumnActions(
        dataModelerStateManager, duckDbTableAPI, duckDBColumnAPI, duckDBDataLoaderAPI);
    const dataModelerActionAPI = new DataModelerActionAPI(dataModelerStateManager, duckDbClient,
        [datasetActions, modelActions, profileColumnActions]);

    const socketServer = new SocketServer(dataModelerActionAPI, dataModelerStateManager,
        "http://localhost:3000", 3001);

    return {dataModelerStateManager, dataModelerActionAPI, socketServer};
}
