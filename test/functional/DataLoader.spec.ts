import {TestBase} from "@adityahegde/typescript-test-utils";
import {JestTestLibrary} from "@adityahegde/typescript-test-utils/dist/jest/JestTestLibrary";
import type {DataModelerActionAPI} from "$common/data-modeler-actions/DataModelerActionAPI";
import type {DataModelerStateManager} from "$common/state-actions/DataModelerStateManager";
import type {SocketServer} from "$common/SocketServer";
import {serverFactory} from "$common/serverFactory";
import {clientFactory} from "$common/clientFactory";
import {asyncWait} from "$common/utils/waitUtils";

@TestBase.Suite
@TestBase.TestLibrary(JestTestLibrary)
export class DataLoaderSpec extends TestBase {
    private serverDataModelerStateManager: DataModelerStateManager;
    private serverDataModelerActionAPI: DataModelerActionAPI;
    private socketServer: SocketServer;

    private clientDataModelerStateManager: DataModelerStateManager;
    private clientDataModelerActionAPI: DataModelerActionAPI;

    @TestBase.BeforeSuite()
    public async setup(): Promise<void> {
        const serverInstances = serverFactory();
        this.serverDataModelerStateManager = serverInstances.dataModelerStateManager;
        this.serverDataModelerActionAPI = serverInstances.dataModelerActionAPI;
        this.socketServer = serverInstances.socketServer;

        const clientInstances = clientFactory();
        this.clientDataModelerStateManager = clientInstances.dataModelerStateManager;
        this.clientDataModelerActionAPI = clientInstances.dataModelerActionAPI;

        await this.socketServer.init();
        await this.clientDataModelerActionAPI.init();
    }

    @TestBase.Test()
    public async shouldLoadParquetFile(): Promise<void> {
        const parquetFile = "AdBids.parquet"

        await this.clientDataModelerActionAPI.dispatch("addOrUpdateDataset", [parquetFile]);

        await asyncWait(2500);

        const dataset = this.clientDataModelerStateManager.getCurrentState().sources
            .find(datasetFind => datasetFind.path === parquetFile);
        console.log(dataset?.profile);
    }

    @TestBase.AfterSuite()
    public async teardown(): Promise<void> {
        await this.clientDataModelerActionAPI.destroy();
        await this.socketServer.destroy();
    }
}
