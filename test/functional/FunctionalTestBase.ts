import { TestBase } from "@adityahegde/typescript-test-utils";
import { JestTestLibrary } from "@adityahegde/typescript-test-utils/dist/jest/JestTestLibrary";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import { dataModelerServiceFactory } from "$common/serverFactory";
import { asyncWait, waitUntil } from "$common/utils/waitUtils";
import { IDLE_STATUS } from "$common/constants";
import type { ProfileColumn } from "$lib/types";
import type { TestDataColumns } from "../data/DataLoader.data";
import { ParquetFileTestData } from "../data/DataLoader.data";
import { DataModelerSocketServiceMock } from "./DataModelerSocketServiceMock";
import { SocketServerMock } from "./SocketServerMock";
import { DATA_FOLDER } from "../data/generator/data-constants";
import { RootConfig } from "$common/config/RootConfig";
import { DatabaseConfig } from "$common/config/DatabaseConfig";
import { StateConfig } from "$common/config/StateConfig";
import {
    EntityRecord,
    EntityStateService, EntityStatus,
    EntityType,
    StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
    PersistentTableEntity
} from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import type {
    DerivedTableEntity
} from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import { dataModelerStateServiceClientFactory } from "$common/clientFactory";

@TestBase.TestLibrary(JestTestLibrary)
export class FunctionalTestBase extends TestBase {
    protected clientDataModelerStateService: DataModelerStateService;
    protected clientDataModelerService: DataModelerService;

    protected serverDataModelerStateService: DataModelerStateService;
    protected serverDataModelerService: DataModelerService;
    protected socketServer: SocketServerMock;

    @TestBase.BeforeSuite()
    public async setup(): Promise<void> {
        this.clientDataModelerStateService = dataModelerStateServiceClientFactory()
        this.clientDataModelerService = new DataModelerSocketServiceMock(this.clientDataModelerStateService);

        const serverInstances = dataModelerServiceFactory(new RootConfig({
            database: new DatabaseConfig({ parquetFolder: "data", databaseName: ":memory:" }),
            state: new StateConfig({ autoSync: false }),
        }));
        this.serverDataModelerStateService = serverInstances.dataModelerStateService;
        this.serverDataModelerService = serverInstances.dataModelerService;
        this.socketServer = new SocketServerMock(this.serverDataModelerService, this.serverDataModelerStateService,
            this.clientDataModelerService as DataModelerSocketServiceMock);
        (this.clientDataModelerService as DataModelerSocketServiceMock).socketServerMock = this.socketServer;

        await this.clientDataModelerService.init();
        await this.socketServer.init();
    }

    @TestBase.AfterSuite()
    public async teardown(): Promise<void> {
        await this.serverDataModelerService?.destroy();
    }

    protected async loadTestTables(): Promise<void> {
        await Promise.all(ParquetFileTestData.subData.map(async (parquetFileData) => {
            await this.clientDataModelerService.dispatch("addOrUpdateTableFromFile", [`${DATA_FOLDER}/${parquetFileData.title}`]);
        }));
        await this.waitForTables();
    }

    protected async waitForTables(): Promise<void> {
        await this.waitForEntity(EntityType.Table);
    }
    protected async waitForModels(): Promise<void> {
        await this.waitForEntity(EntityType.Model);
    }

    protected getTables(field: string, value: any): [PersistentTableEntity, DerivedTableEntity] {
        const persistent = this.getEntityByField(
            EntityType.Table, StateType.Persistent, field, value) as PersistentTableEntity;
        return [
            persistent,
            this.getEntityByField(EntityType.Table, StateType.Derived, "id", persistent.id) as DerivedTableEntity,
        ];
    }

    protected assertColumns(profileColumns: ProfileColumn[], columns: TestDataColumns): void {
        profileColumns.forEach((profileColumn, idx) => {
            expect(profileColumn.name).toBe(columns[idx].name);
            expect(profileColumn.type).toBe(columns[idx].type);
            expect(profileColumn.nullCount > 0).toBe(columns[idx].isNull);
            // TODO: assert summary
            // console.log(profileColumn.name, profileColumn.summary);
        });
    }

    private async waitForEntity(entityType: EntityType): Promise<void> {
        await asyncWait(200);
        await waitUntil(() => {
            const currentState = this.clientDataModelerStateService
                .getEntityStateService(entityType, StateType.Derived)
                .getCurrentState();
            return (currentState.entities as any[]).every(item => item.status === EntityStatus.Idle);
        });
    }

    private getEntityByField(entityType: EntityType, stateType: StateType,
                             field: string, value: any): EntityRecord {
        return (this.clientDataModelerStateService
            .getEntityStateService(entityType, stateType) as EntityStateService<any>)
            .getByField(field, value);
    }
}
