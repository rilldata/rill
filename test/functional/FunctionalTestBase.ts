import {TestBase} from "@adityahegde/typescript-test-utils";
import {JestTestLibrary} from "@adityahegde/typescript-test-utils/dist/jest/JestTestLibrary";
import {DataModelerStateService} from "$common/data-modeler-state-service/DataModelerStateService";
import type {DataModelerService} from "$common/data-modeler-service/DataModelerService";
import {dataModelerServiceFactory} from "$common/serverFactory";
import {asyncWait, waitUntil} from "$common/utils/waitUtils";
import {IDLE_STATUS} from "$common/constants";
import type {ColumnarTypeKeys, ProfileColumn} from "$lib/types";
import type {TestDataColumns} from "../data/DataLoader.data";
import {DataModelerSocketServiceMock} from "./DataModelerSocketServiceMock";
import {SocketServerMock} from "./SocketServerMock";
import {ParquetFileTestData} from "../data/DataLoader.data";
import {DATA_FOLDER} from "../data/generator/data-constants";
import {RootConfig} from "$common/config/RootConfig";
import { ColumnarItemType, ColumnarItemTypeMap } from "$common/data-modeler-state-service/ProfileColumnStateActions";
import { DatabaseConfig } from "$common/config/DatabaseConfig";

@TestBase.TestLibrary(JestTestLibrary)
export class FunctionalTestBase extends TestBase {
    protected clientDataModelerStateService: DataModelerStateService;
    protected clientDataModelerService: DataModelerService;

    protected serverDataModelerStateService: DataModelerStateService;
    protected serverDataModelerService: DataModelerService;
    protected socketServer: SocketServerMock;

    @TestBase.BeforeSuite()
    public async setup(): Promise<void> {
        this.clientDataModelerStateService = new DataModelerStateService([]);
        this.clientDataModelerService = new DataModelerSocketServiceMock(this.clientDataModelerStateService);

        const serverInstances = dataModelerServiceFactory(new RootConfig({
            database: new DatabaseConfig({ parquetFolder: "data", databaseName: ":memory:" }),
        }));
        this.serverDataModelerStateService = serverInstances.dataModelerStateService;
        this.serverDataModelerService = serverInstances.dataModelerService;
        this.socketServer = new SocketServerMock(this.serverDataModelerService, this.serverDataModelerStateService,
            this.clientDataModelerService as DataModelerSocketServiceMock);
        (this.clientDataModelerService as DataModelerSocketServiceMock).socketServerMock = this.socketServer;

        await this.clientDataModelerService.init();
        await this.socketServer.init();
    }

    protected async loadTestTables(): Promise<void> {
        await Promise.all(ParquetFileTestData.subData.map(async (parquetFileData) => {
            await this.clientDataModelerService.dispatch("addOrUpdateTable", [`${DATA_FOLDER}/${parquetFileData.title}`]);
        }));
        await this.waitForTables();
    }

    protected async waitForTables(): Promise<void> {
        await this.waitForColumnar(ColumnarItemTypeMap[ColumnarItemType.Table]);
    }

    protected async waitForModels(): Promise<void> {
        await this.waitForColumnar(ColumnarItemTypeMap[ColumnarItemType.Model]);
    }

    private async waitForColumnar(columnarKey: ColumnarTypeKeys): Promise<void> {
        await asyncWait(200);
        await waitUntil(() => {
            const currentState = this.clientDataModelerStateService.getCurrentState();
            return (currentState[columnarKey] as any[]).every(item => item.status === IDLE_STATUS);
        });
    }
}
