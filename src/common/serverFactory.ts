import {DuckDBClient} from "$common/duckdb/DuckDBClient";
import {DuckDBTableAPI} from "$common/duckdb/DuckDBTableAPI";
import {DuckDBColumnAPI} from "$common/duckdb/DuckDBColumnAPI";
import {DuckDBDataLoaderAPI} from "$common/duckdb/DuckDBDataLoaderAPI";
import {DataModelerStateManager} from "$common/state-actions/DataModelerStateManager";
import {DatasetStateActions} from "$common/state-actions/DatasetStateActions";
import {ModelStateActions} from "$common/state-actions/ModelStateActions";
import {ProfileColumnStateActions} from "$common/state-actions/ProfileColumnStateActions";
import {DatasetActions} from "$common/data-modeler-actions/DatasetActions";
import {DataModelerActionAPI} from "$common/data-modeler-actions/DataModelerActionAPI";
import {ProfileColumnActions} from "$common/data-modeler-actions/ProfileColumnActions";
import {SocketServer} from "$common/SocketServer";

export function serverFactory(): {
    dataModelerStateManager: DataModelerStateManager,
    dataModelerActionAPI: DataModelerActionAPI,
    socketServer: SocketServer,
} {
    const duckDbClient = new DuckDBClient();
    const duckDbTableAPI = new DuckDBTableAPI(duckDbClient);
    const duckDBColumnAPI = new DuckDBColumnAPI(duckDbClient);
    const duckDBDataLoaderAPI = new DuckDBDataLoaderAPI(duckDbClient);

    const datasetStateActions = new DatasetStateActions();
    const modelStateActions = new ModelStateActions();
    const profileColumnStateActions = new ProfileColumnStateActions();
    const dataModelerStateManager = new DataModelerStateManager(
        [datasetStateActions, modelStateActions, profileColumnStateActions]);

    const datasetActions = new DatasetActions(
        dataModelerStateManager, duckDbTableAPI, duckDBColumnAPI, duckDBDataLoaderAPI);
    const profileColumnActions = new ProfileColumnActions(
        dataModelerStateManager, duckDbTableAPI, duckDBColumnAPI, duckDBDataLoaderAPI);
    const dataModelerActionAPI = new DataModelerActionAPI(dataModelerStateManager, duckDbClient,
        [datasetActions, profileColumnActions]);

    const socketServer = new SocketServer(dataModelerActionAPI, dataModelerStateManager,
        "http://localhost:3000", 3001);

    return {dataModelerStateManager, dataModelerActionAPI, socketServer};
}
