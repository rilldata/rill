import {DataModelerActions} from ".//DataModelerActions";
import type {DataModelerState, Table} from "$lib/types";
import {getNewTable} from "$common/stateInstancesFactory";
import {ColumnarItemType} from "$common/data-modeler-state-service/ProfileColumnStateActions";
import {IDLE_STATUS, RUNNING_STATUS} from "$common/constants";
import { extractFileExtension, extractTableName, sanitizeTableName } from "$lib/util/extract-table-name";
import {getParquetFiles} from "$common/utils/getParquetFiles";
import {stat} from "fs/promises";
import { FILE_EXTENSION_TO_TABLE_TYPE, TableSourceType } from "$lib/types";

export class TableActions extends DataModelerActions {
    public async updateTablesFromSource(currentState: DataModelerState, sourcePath: string): Promise<void> {
        const files = await getParquetFiles(sourcePath);
        const filePaths = new Set(files);
        const newTables = currentState.tables.filter((table, index, self) => {
            if (!filePaths.has(table.path)) return false;
            return index === self.findIndex(indexCheckTable => (indexCheckTable.path === table.path));
        });
        if (currentState.tables.length !== newTables.length) {
            this.dataModelerStateService.dispatch("pruneAndDedupeTables", [files]);
        }

        await this.dataModelerService.dispatch("addOrUpdateAllTables", [files]);
    }

    public async addOrUpdateAllTables(currentState: DataModelerState, files: Array<string>): Promise<void> {
        const filePaths = new Set(files);
        await Promise.all(currentState.tables.map(async (table) => {
            const fileStats = await stat(table.path);
            if (fileStats.mtimeMs < table.lastUpdated) filePaths.delete(table.path);
            else filePaths.add(table.path);
        }));
        if (filePaths.size > 0) {
            await Promise.all([...filePaths].map(filePath =>
              this.dataModelerService.dispatch("addOrUpdateTableFromFile", [filePath])));
        }
    }

    public async addOrUpdateTableFromFile(currentState: DataModelerState, path: string, tableName?: string): Promise<void> {
        const name = tableName ?? sanitizeTableName(extractTableName(path));
        const type = FILE_EXTENSION_TO_TABLE_TYPE[extractFileExtension(path)];
        if (type === undefined) {
            // TODO: Create a error response pipeline
            console.log("Invalid file type");
            return;
        }

        const tables = currentState.tables;
        const existingTable = tables.find(s => s.path === path);
        const table = {...(existingTable || getNewTable())};
        table.path = path;
        table.name = name;
        table.tableName = name;
        table.sourceType = type;

        // get stats of the file and update only if it changed since we last saw it
        const fileStats = await stat(path);
        if (fileStats.mtimeMs < table.lastUpdated) return;
        table.lastUpdated = Date.now();

        await this.dataModelerService.dispatch("addOrUpdateTable", [table, !existingTable]);
    }

    public async addOrUpdateTable(currentState: DataModelerState, table: Table, isNew: boolean): Promise<void> {
        this.dataModelerStateService.dispatch("addOrUpdateTableToState",
          [table, isNew]);
        this.dataModelerStateService.dispatch("setTableStatus",
          [ColumnarItemType.Table, table.id, RUNNING_STATUS]);

        try {
            await this.collectTableInfo(table);

            await this.dataModelerService.dispatch("collectProfileColumns",
              [table.id, ColumnarItemType.Table]);
        } catch (err) {
            console.log(err);
        }

        this.dataModelerStateService.dispatch("setTableStatus",
          [ColumnarItemType.Table, table.id, IDLE_STATUS]);
    }

    // TODO: move this to something more meaningful
    public async setActiveAsset(currentState: DataModelerState, id: string, assetType: string): Promise<void> {
        this.dataModelerStateService.dispatch("setActiveAsset", [id, assetType]);
    }
    public async unsetActiveAsset(currentState: DataModelerState): Promise<void> {
        this.dataModelerStateService.dispatch("unsetActiveAsset", []);
    }

    private async collectTableInfo(table: Table) {
        switch (table.sourceType) {
            case TableSourceType.ParquetFile:
                await this.databaseService.dispatch("importParquetFile", [table.path, table.tableName]);
                break;
        }

        // create new table as one passed in args is readonly from the state.
        const newTable: Table = {
            id: table.id,
            path: table.path,
            name: table.name,
            tableName: table.tableName,
            head: undefined,
        };

        await Promise.all([
            async () => {
                newTable.profile = await this.databaseService.dispatch(
                  "getProfileColumns", [table.tableName]);
                newTable.profile = newTable.profile
                    .filter(row => row.name !== "duckdb_schema" && row.name !== "schema" && row.name !== "root");
            },
            async () => newTable.sizeInBytes =
                await this.databaseService.dispatch("getDestinationSize", [table.path]),
            async () => newTable.cardinality =
                await this.databaseService.dispatch("getCardinalityOfTable", [table.tableName]),
            async () => newTable.head =
                await this.databaseService.dispatch("getFirstNOfTable", [table.tableName]),
        ].map(asyncFunc => asyncFunc()));

        this.dataModelerStateService.dispatch("addOrUpdateTableToState",
            [newTable, false]);
    }
}
