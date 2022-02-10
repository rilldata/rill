import {TestBase} from "@adityahegde/typescript-test-utils";
import {JestTestLibrary} from "@adityahegde/typescript-test-utils/dist/jest/JestTestLibrary";
import {DataModelerStateService} from "$common/data-modeler-state-service/DataModelerStateService";
import type {DataModelerService} from "$common/data-modeler-service/DataModelerService";
import type {SocketServer} from "$common/SocketServer";
import {dataModelerServiceFactory} from "$common/serverFactory";
import {asyncWait, waitUntil} from "$common/utils/waitUtils";
import {IDLE_STATUS} from "$common/constants";
import type {ColumnarTypeKeys, ProfileColumn} from "$lib/types";
import type {TestDataColumns} from "../data/DataLoader.data";
import {DataModelerSocketServiceMock} from "./DataModelerSocketServiceMock";
import {SocketServerMock} from "./SocketServerMock";

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

        const serverInstances = dataModelerServiceFactory();
        this.serverDataModelerStateService = serverInstances.dataModelerStateService;
        this.serverDataModelerService = serverInstances.dataModelerService;
        this.socketServer = new SocketServerMock(this.serverDataModelerService, this.serverDataModelerStateService,
            this.clientDataModelerService as DataModelerSocketServiceMock);
        (this.clientDataModelerService as DataModelerSocketServiceMock).socketServerMock = this.socketServer;

        await this.clientDataModelerService.init();
        await this.socketServer.init();
    }

    protected async waitForDatasets(): Promise<void> {
        await this.waitForColumnar("sources");
    }

    protected async waitForModels(): Promise<void> {
        await this.waitForColumnar("queries");
    }

    protected assertColumns(profileColumns: ProfileColumn[], columns: TestDataColumns): void {
        profileColumns.forEach((profileColumn, idx) => {
            expect(profileColumn.name).toBe(columns[idx].name);
            expect(profileColumn.type).toBe(columns[idx].type);
            expect(profileColumn.nullCount > 0).toBe(columns[idx].isNull);
            // TODO: assert summary
        });
    }

    private async waitForColumnar(columnarKey: ColumnarTypeKeys): Promise<void> {
        await asyncWait(200);
        await waitUntil(() => {
            const currentState = this.clientDataModelerStateService.getCurrentState();
            return currentState.status === IDLE_STATUS &&
                (currentState[columnarKey] as any[]).every(item => item.status === IDLE_STATUS);
        });
    }
}
