import {DataModelerActions} from ".//DataModelerActions";
import {FILE_EXTENSION_TO_TABLE_TYPE, ProfileColumn, TableSourceType} from "$lib/types";
import {getNewDerivedTable, getNewTable} from "$common/stateInstancesFactory";
import {extractFileExtension, getTableNameFromFile, INVALID_CHARS} from "$lib/util/extract-table-name";
import type {
    PersistentTableEntity,
    PersistentTableStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import {
    EntityStatus,
    EntityType,
    StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
    DerivedTableEntity,
    DerivedTableStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import {DatabaseActionQueuePriority} from "$common/priority-action-queue/DatabaseActionQueuePriority";
import {existsSync} from "fs";
import {ActionResponseFactory} from "$common/data-modeler-service/response/ActionResponseFactory";
import type {ActionResponse} from "$common/data-modeler-service/response/ActionResponse";

export interface ImportTableOptions {
    csvDelimiter?: string;
}

export class TableActions extends DataModelerActions {
    @DataModelerActions.PersistentTableAction()
    public async clearAllTables({stateService}: PersistentTableStateActionArg): Promise<void> {
        stateService.getCurrentState().entities.forEach((table) => {
            this.dataModelerStateService.dispatch("deleteEntity",
                [EntityType.Table, StateType.Persistent, table.id]);
            this.dataModelerStateService.dispatch("deleteEntity",
                [EntityType.Table, StateType.Derived, table.id]);
        });
    }

    @DataModelerActions.PersistentTableAction()
    public async addOrUpdateTableFromFile({stateService}: PersistentTableStateActionArg, path: string,
                                          tableName?: string, options: ImportTableOptions = {}): Promise<ActionResponse> {
        const name = getTableNameFromFile(path, tableName);
        const type = FILE_EXTENSION_TO_TABLE_TYPE[extractFileExtension(path)];

        if (!existsSync(path)) {
            return ActionResponseFactory.getImportTableError(
                `File ${path} does not exist`);
        }

        if (type === undefined) {
            return ActionResponseFactory.getImportTableError(
                `Invalid file type`);
        }
        if (tableName && INVALID_CHARS.test(tableName)) {
            return ActionResponseFactory.getImportTableError(
                `Input table name has invalid characters`);
        }

        const existingTable = stateService.getByField("tableName", name);
        const table = existingTable ? {...existingTable} : getNewTable();

        table.path = path;
        table.name = name;
        table.tableName = name;
        table.sourceType = type;
        if (options.csvDelimiter) {
            table.csvDelimiter = options.csvDelimiter;
        }

        table.lastUpdated = Date.now();

        await this.addOrUpdateTable(table, !existingTable);
    }

    @DataModelerActions.PersistentTableAction()
    public async addOrSyncTableFromDB(
        {stateService}: PersistentTableStateActionArg, tableName?: string
    ) {
        const existingTable = stateService.getByField("tableName", tableName);
        const table = existingTable ? {...existingTable} : getNewTable();

        table.name = table.tableName = tableName;
        table.sourceType = TableSourceType.DuckDB;

        await this.addOrUpdateTable(table, !existingTable);
    }

    @DataModelerActions.DerivedTableAction()
    @DataModelerActions.ResetStateToIdle(EntityType.Table)
    public async collectTableInfo({stateService}: DerivedTableStateActionArg, tableId: string): Promise<ActionResponse> {
        const persistentTable = this.dataModelerStateService
            .getEntityById(EntityType.Table, StateType.Persistent, tableId);
        const newDerivedTable: DerivedTableEntity = {
            id: tableId,
            type: EntityType.Table,
            status: EntityStatus.Profiling,
            lastUpdated: Date.now(),
            profiled: false,
        };

        if (!persistentTable) {
            return ActionResponseFactory.getEntityError(`No table found for ${tableId}`);
        }
        this.databaseActionQueue.clearQueue(tableId);

        try {
            this.dataModelerStateService.dispatch("setEntityStatus",
                [EntityType.Table, tableId, EntityStatus.Profiling]);
            await this.dataModelerStateService.dispatch("clearProfileSummary",
                [EntityType.Table, tableId]);

            await Promise.all([
                async () => {
                    newDerivedTable.profile = await this.databaseActionQueue.enqueue(
                        {id: tableId, priority: DatabaseActionQueuePriority.TableImport},
                        "getProfileColumns", [persistentTable.tableName]);
                    newDerivedTable.profile = newDerivedTable.profile
                        .filter(row => row.name !== "duckdb_schema" && row.name !== "schema" && row.name !== "root");
                },
                async () => newDerivedTable.sizeInBytes = await this.databaseActionQueue.enqueue(
                    {id: tableId, priority: DatabaseActionQueuePriority.TableProfile},
                    "getDestinationSize", [persistentTable.path]),
                async () => newDerivedTable.cardinality = await this.databaseActionQueue.enqueue(
                    {id: tableId, priority: DatabaseActionQueuePriority.TableProfile},
                    "getCardinalityOfTable", [persistentTable.tableName]),
                async () => newDerivedTable.preview = await this.databaseActionQueue.enqueue(
                    {id: tableId, priority: DatabaseActionQueuePriority.TableProfile},
                    "getFirstNOfTable", [persistentTable.tableName]),
            ].map(asyncFunc => asyncFunc()));

            this.dataModelerStateService.dispatch("updateEntity",
                [EntityType.Table, StateType.Derived, newDerivedTable])
            await this.dataModelerService.dispatch("collectProfileColumns",
                [EntityType.Table, tableId]);
            this.dataModelerStateService.dispatch("markAsProfiled",
                [EntityType.Table, tableId, true]);
        } catch (err) {
            return ActionResponseFactory.getErrorResponse(err);
        }
    }

    @DataModelerActions.PersistentTableAction()
    public async dropTable({stateService}: PersistentTableStateActionArg,
                           tableName: string, removeOnly = false): Promise<ActionResponse> {
        const table = stateService.getByField("tableName", tableName);
        if (!table) {
            return ActionResponseFactory.getEntityError(`No table found for ${tableName}`);
        }

        if (!removeOnly) {
            await this.databaseActionQueue.enqueue(
                {id: table.id, priority: DatabaseActionQueuePriority.TableImport},
                "dropTable", [table.tableName]);
        }
        this.notificationService.notify({ message: `dropped table ${table.tableName}`, type: "info"});

        await this.dataModelerService.dispatch("deleteEntity",
            [EntityType.Table, table.id]);
    }

    @DataModelerActions.DerivedTableAction()
    @DataModelerActions.ResetStateToIdle(EntityType.Table)
    public async syncTable({stateService}: DerivedTableStateActionArg, tableId: string) {
        const derivedTable = stateService.getById(tableId);
        const persistentTable = this.dataModelerStateService
            .getEntityStateService(EntityType.Table, StateType.Persistent)
            .getById(tableId);
        if (!derivedTable || !persistentTable) {
            return ActionResponseFactory.getEntityError(`No table found for ${tableId}`);
        }
        this.dataModelerStateService.dispatch("setEntityStatus",
            [EntityType.Table, tableId, EntityStatus.Profiling]);

        // check row count
        const newCardinality = await this.databaseActionQueue.enqueue(
            {id: tableId, priority: DatabaseActionQueuePriority.TableProfile},
            "getCardinalityOfTable", [persistentTable.tableName]);
        if (newCardinality === derivedTable.cardinality) return;

        // check column count and names
        const newProfiles: Array<ProfileColumn> = await this.databaseActionQueue.enqueue(
            {id: tableId, priority: DatabaseActionQueuePriority.TableImport},
            "getProfileColumns", [persistentTable.tableName]);
        if (newProfiles.length === derivedTable.profile.length) {
            const existingColumns = new Map<string, ProfileColumn>();
            derivedTable.profile.forEach(column => existingColumns.set(column.name, column));

            if (newProfiles.every(newProfileColumn =>
                existingColumns.has(newProfileColumn.name) &&
                existingColumns.get(newProfileColumn.name).type === newProfileColumn.type
            )) {
                return;
            }
        }

        await this.dataModelerService.dispatch("collectTableInfo", [tableId]);
    }

    private async addOrUpdateTable(table: PersistentTableEntity, isNew: boolean): Promise<void> {
        if (isNew) {
            const derivedTable = getNewDerivedTable(table);
            derivedTable.status = EntityStatus.Importing;
            this.dataModelerStateService.dispatch("addEntity",
                [EntityType.Table, StateType.Derived, derivedTable]);
        } else {
            this.dataModelerStateService.dispatch("setEntityStatus",
                [EntityType.Table, table.id, EntityStatus.Importing]);
        }
        this.dataModelerStateService.dispatch("addOrUpdateTableToState",
            [table, isNew]);

        await this.importTableDataByType(table);

        if (this.config.profileWithUpdate) {
            await this.dataModelerService.dispatch("collectTableInfo", [table.id]);
        } else {
            this.dataModelerStateService.dispatch("markAsProfiled",
                [EntityType.Table, table.id, false]);
        }
        this.dataModelerStateService.dispatch("setEntityStatus",
            [EntityType.Table, table.id, EntityStatus.Idle]);
    }

    private async importTableDataByType(table: PersistentTableEntity) {
        switch (table.sourceType) {
            case TableSourceType.ParquetFile:
                await this.databaseActionQueue.enqueue(
                    {id: table.id, priority: DatabaseActionQueuePriority.TableImport},
                    "importParquetFile", [table.path, table.tableName]);
                break;

            case TableSourceType.CSVFile:
                await this.databaseActionQueue.enqueue(
                    {id: table.id, priority: DatabaseActionQueuePriority.TableImport},
                    "importCSVFile", [table.path, table.tableName, table.csvDelimiter]);
                break;

            case TableSourceType.DuckDB:
                // table already exists. nothing to do here
                break;
        }
        this.notificationService.notify({ message: `imported ${table.name}`, type: "info"});
    }
}
