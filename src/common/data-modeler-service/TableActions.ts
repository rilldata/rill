import { DataModelerActions } from ".//DataModelerActions";
import { FILE_EXTENSION_TO_TABLE_TYPE, TableSourceType } from "$lib/types";
import { getNewDerivedTable, getNewTable } from "$common/stateInstancesFactory";
import { extractFileExtension, extractTableName, INVALID_CHARS, sanitizeTableName } from "$lib/util/extract-table-name";
import { stat } from "fs/promises";
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

export interface ImportTableOptions {
    csvDelimiter?: string;
}

export class TableActions extends DataModelerActions {
    @DataModelerActions.PersistentTableAction()
    public async addOrUpdateTableFromFile({stateService}: PersistentTableStateActionArg, path: string,
                                          tableName?: string, options: ImportTableOptions = {}): Promise<void> {
        const name = tableName ?? sanitizeTableName(extractTableName(path));
        const type = FILE_EXTENSION_TO_TABLE_TYPE[extractFileExtension(path)];
        if (type === undefined) {
            // TODO: Create a error response pipeline
            console.error("Invalid file type");
            return;
        }
        if (tableName && INVALID_CHARS.test(tableName)) {
            console.error("Input table name has invalid characters");
            return;
        }

        const existingTable = stateService.getByField("path", path);
        const table = existingTable ? {...existingTable} : getNewTable();

        if (existingTable && existingTable.tableName !== name) {
            console.error("New table name doesnt match existing. Renaming is not supported at the moment.");
            return;
        }
        const existingByName = stateService.getByField("tableName", name);
        if (existingByName && existingByName.path !== path) {
            console.error(`Another table with ${name} already exists.`);
            return;
        }

        table.path = path;
        table.name = name;
        table.tableName = name;
        table.sourceType = type;
        if (options.csvDelimiter) {
            table.csvDelimiter = options.csvDelimiter;
        }

        // get stats of the file and update only if it changed since we last saw it
        const fileStats = await stat(path);
        if (fileStats.mtimeMs < table.lastUpdated) return;
        table.lastUpdated = Date.now();

        await this.addOrUpdateTable(table, !existingTable);
    }

    @DataModelerActions.DerivedTableAction()
    public async collectTableInfo({stateService}: DerivedTableStateActionArg, tableId: string): Promise<void> {
        const persistentTable = this.dataModelerStateService
            .getEntityById(EntityType.Table, StateType.Persistent, tableId);
        const newDerivedTable: DerivedTableEntity = {
            id: tableId,
            type: EntityType.Table,
            status: EntityStatus.Profiling,
            lastUpdated: Date.now(),
        };

        this.dataModelerStateService.dispatch("setTableStatus",
            [EntityType.Table, tableId, EntityStatus.Profiling]);
        await this.dataModelerStateService.dispatch("clearProfileSummary",
            [EntityType.Table, tableId]);
        // TODO: should there be an update to lastUpdated here?

        await Promise.all([
            async () => {
                newDerivedTable.profile = await this.databaseService.dispatch(
                    "getProfileColumns", [persistentTable.tableName]);
                newDerivedTable.profile = newDerivedTable.profile
                    .filter(row => row.name !== "duckdb_schema" && row.name !== "schema" && row.name !== "root");
            },
            async () => newDerivedTable.sizeInBytes =
                await this.databaseService.dispatch("getDestinationSize", [persistentTable.path]),
            async () => newDerivedTable.cardinality =
                await this.databaseService.dispatch("getCardinalityOfTable", [persistentTable.tableName]),
            async () => newDerivedTable.preview =
                await this.databaseService.dispatch("getFirstNOfTable", [persistentTable.tableName]),
        ].map(asyncFunc => asyncFunc()));

        this.dataModelerStateService.dispatch("updateEntity",
            [EntityType.Table, StateType.Derived, newDerivedTable])
        await this.dataModelerService.dispatch("collectProfileColumns",
            [EntityType.Table, tableId]);
        this.dataModelerStateService.dispatch("setTableStatus",
            [EntityType.Table, tableId, EntityStatus.Idle]);
    }

    private async addOrUpdateTable(table: PersistentTableEntity, isNew: boolean): Promise<void> {
        if (isNew) {
            this.dataModelerStateService.dispatch("addEntity",
                [EntityType.Table, StateType.Derived, getNewDerivedTable(table)]);
        }
        this.dataModelerStateService.dispatch("addOrUpdateTableToState",
            [table, isNew]);

        try {
            await this.importTableDataByType(table);
            await this.dataModelerService.dispatch("collectTableInfo", [table.id]);
        } catch (err) {
            console.error(err);
        }
    }

    private async importTableDataByType(table: PersistentTableEntity) {
        switch (table.sourceType) {
            case TableSourceType.ParquetFile:
                await this.databaseService.dispatch("importParquetFile", [table.path, table.tableName]);
                break;

            case TableSourceType.CSVFile:
                await this.databaseService.dispatch("importCSVFile", [table.path, table.tableName, table.csvDelimiter]);
                break;
        }
    }
}
