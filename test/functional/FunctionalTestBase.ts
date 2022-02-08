import {TestBase} from "@adityahegde/typescript-test-utils";
import {JestTestLibrary} from "@adityahegde/typescript-test-utils/dist/jest/JestTestLibrary";
import type {DataModelerStateManager} from "$common/state-actions/DataModelerStateManager";
import type {DataModelerActionAPI} from "$common/data-modeler-actions/DataModelerActionAPI";
import type {SocketServer} from "$common/SocketServer";
import {serverFactory} from "$common/serverFactory";
import {clientFactory} from "$common/clientFactory";
import {asyncWait, waitUntil} from "$common/utils/waitUtils";
import {IDLE_STATUS} from "$common/constants";
import type {ColumnarTypeKeys, ProfileColumn} from "$lib/types";
import type {TestDataColumns} from "../data/DataLoader.data";

@TestBase.TestLibrary(JestTestLibrary)
export class FunctionalTestBase extends TestBase {
    protected serverDataModelerStateManager: DataModelerStateManager;
    protected serverDataModelerActionAPI: DataModelerActionAPI;
    protected socketServer: SocketServer;

    protected clientDataModelerStateManager: DataModelerStateManager;
    protected clientDataModelerActionAPI: DataModelerActionAPI;

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

    @TestBase.AfterSuite()
    public async teardown(): Promise<void> {
        await this.clientDataModelerActionAPI.destroy();
        await this.socketServer.destroy();
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
            const currentState = this.clientDataModelerStateManager.getCurrentState();
            return currentState.status === IDLE_STATUS &&
                (currentState[columnarKey] as any[]).every(item => item.status === IDLE_STATUS);
        });
    }
}
