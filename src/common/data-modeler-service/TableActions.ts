import { DataModelerActions } from ".//DataModelerActions";
import { FILE_EXTENSION_TO_TABLE_TYPE, TableSourceType } from "$lib/types";
import { getNewDerivedTable, getNewTable } from "$common/stateInstancesFactory";
import { extractFileExtension, getTableNameFromFile, INVALID_CHARS } from "$lib/util/extract-table-name";
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
import { DatabaseActionQueuePriority } from "$common/priority-action-queue/DatabaseActionQueuePriority";
import { existsSync } from "fs";
import { ActionResponseFactory } from "$common/data-modeler-service/response/ActionResponseFactory";
import { ActionResponse, ActionStatus } from "$common/data-modeler-service/response/ActionResponse";

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

        const response = await this.addOrUpdateTable(table, !existingTable);
        if (response?.status !== undefined && response?.status === ActionStatus.Failure) {
            return response;
        }
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
                           tableName: string): Promise<ActionResponse> {
        const table = stateService.getByField("tableName", tableName);
        if (!table) {
            return ActionResponseFactory.getEntityError(`No table found for ${tableName}`);
        }

        await this.databaseService.dispatch("dropTable", [table.tableName]);
        this.notificationService.notify({ message: `dropped table ${table.tableName}`, type: "info"});

        await this.dataModelerService.dispatch("deleteEntity",
            [EntityType.Table, table.id]);
    }

    
    private async addOrUpdateTable(table: PersistentTableEntity, isNew: boolean): Promise<ActionResponse> {
        
        // get the original Table state if not new.
        let originalPersistentTable:PersistentTableEntity;
        if (!isNew) {
            originalPersistentTable = this.dataModelerStateService
            .getEntityStateService(EntityType.Table, StateType.Persistent)
            .getByField("tableName", table.name)
        }

        // update the new state
        if (isNew) {
            this.dataModelerStateService.dispatch("addEntity",
                [EntityType.Table, StateType.Persistent, table]);
        } else {
            this.dataModelerStateService.dispatch("updateEntity",
                [EntityType.Table, StateType.Persistent, table]);
            //stateService.updateEntity(draftState, table.id, table);
        }


        let derivedTable:DerivedTableEntity;
        if (isNew) {
            derivedTable = getNewDerivedTable(table);
            derivedTable.status = EntityStatus.Importing;
            this.dataModelerStateService.dispatch("addEntity",
                [EntityType.Table, StateType.Derived, derivedTable]);
        } else {
            this.dataModelerStateService.dispatch("setEntityStatus",
                [EntityType.Table, table.id, EntityStatus.Importing]);
        }
        this.dataModelerStateService.dispatch("addOrUpdateTableToState",
            [table, isNew]);

        const response = await this.importTableDataByType(table);
        if (response?.status !== undefined && (response?.status === ActionStatus.Failure)) {
            if (isNew) {
                // Delete the table entirely.
                this.dataModelerStateService.dispatch("deleteEntity",
                [EntityType.Table, StateType.Derived, derivedTable.id]);
                // Fetch the persistent table in this instance
                // and delete
                const existingTable = this.dataModelerStateService
                    .getEntityStateService(EntityType.Table, StateType.Persistent)
                    .getByField("tableName", table.name);
                this.dataModelerStateService.dispatch("deleteEntity",
                [EntityType.Table, StateType.Persistent, existingTable.id]);
            } else {
                this.dataModelerStateService.dispatch("updateEntity",
                [EntityType.Table, StateType.Persistent, originalPersistentTable]);
                // Reset entity status to idle in the case where the table already exists.
                // nothing has updated here I think?
                this.dataModelerStateService.dispatch("setEntityStatus",
                [EntityType.Table, table.id, EntityStatus.Idle]);
            }
            return response;
        }

        if (this.config.profileWithUpdate) {
            await this.dataModelerService.dispatch("collectTableInfo", [table.id]);
        } else {
            this.dataModelerStateService.dispatch("markAsProfiled",
                [EntityType.Table, table.id, false]);
        }
        this.dataModelerStateService.dispatch("setEntityStatus",
            [EntityType.Table, table.id, EntityStatus.Idle]);
    }

    private async importTableDataByType(table: PersistentTableEntity) : Promise<ActionResponse> {
        let response:ActionResponse;
        switch (table.sourceType) {
            case TableSourceType.ParquetFile:
                response = await this.databaseActionQueue.enqueue(
                    {id: table.id, priority: DatabaseActionQueuePriority.TableImport},
                    "importParquetFile", [table.path, table.tableName]);
                break;

            case TableSourceType.CSVFile:
                response = await this.databaseActionQueue.enqueue(
                    {id: table.id, priority: DatabaseActionQueuePriority.TableImport},
                    "importCSVFile", [table.path, table.tableName, table.csvDelimiter]);
                break;
        }
        if (response) {
            this.notificationService.notify({ message: `failed to import ${table.name} from ${table.path}`, type: "error"});
            return response;
        } else {
            this.notificationService.notify({ message: `imported ${table.name}`, type: "info"});
        }
            

    }
}
