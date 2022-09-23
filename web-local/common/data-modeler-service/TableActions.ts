import { SOURCE_PREVIEW_COUNT } from "../constants";
import {
  ActionResponse,
  ActionStatus,
} from "./response/ActionResponse";
import { ActionResponseFactory } from "./response/ActionResponseFactory";
import type {
  DerivedTableEntity,
  DerivedTableStateActionArg,
} from "../data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import {
  EntityStatus,
  EntityType,
  StateType,
} from "../data-modeler-state-service/entity-state-service/EntityStateService";
import type {
  PersistentTableEntity,
  PersistentTableStateActionArg,
} from "../data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { DatabaseActionQueuePriority } from "../priority-action-queue/DatabaseActionQueuePriority";
import { getNewDerivedTable, getNewTable } from "../stateInstancesFactory";
import { getName } from "../utils/incrementName";
import {
  FILE_EXTENSION_TO_TABLE_TYPE,
  ProfileColumn,
  TableSourceType,
} from "../../lib/types";
import {
  extractFileExtension,
  extractTableName,
  getTableNameFromFile,
  INVALID_CHARS,
  sanitizeEntityName,
} from "../../lib/util/extract-table-name";
import { existsSync } from "fs";
import { DataModelerActions } from ".//DataModelerActions";

export interface ImportTableOptions {
  csvDelimiter?: string;
  shouldNotProfile?: boolean;
}

export class TableActions extends DataModelerActions {
  @DataModelerActions.PersistentTableAction()
  public async clearAllTables({
    stateService,
  }: PersistentTableStateActionArg): Promise<void> {
    stateService.getCurrentState().entities.forEach((table) => {
      this.dataModelerStateService.dispatch("deleteEntity", [
        EntityType.Table,
        StateType.Persistent,
        table.id,
      ]);
      this.dataModelerStateService.dispatch("deleteEntity", [
        EntityType.Table,
        StateType.Derived,
        table.id,
      ]);
    });
  }

  @DataModelerActions.PersistentTableAction()
  public async addOrUpdateTableFromFile(
    { stateService }: PersistentTableStateActionArg,
    path: string,
    tableName?: string,
    options: ImportTableOptions = {}
  ): Promise<ActionResponse> {
    const name = getTableNameFromFile(path, tableName);
    const type = FILE_EXTENSION_TO_TABLE_TYPE[extractFileExtension(path)];

    if (!existsSync(path)) {
      return ActionResponseFactory.getImportTableError(
        `File ${path} does not exist`
      );
    }

    if (type === undefined) {
      return ActionResponseFactory.getImportTableError(`Invalid file type`);
    }
    if (tableName && INVALID_CHARS.test(tableName)) {
      return ActionResponseFactory.getImportTableError(
        `Input table name has invalid characters`
      );
    }

    const existingTable = stateService.getByField("tableName", name);
    const table = existingTable ? { ...existingTable } : getNewTable();

    table.path = path;
    table.name = name;
    table.tableName = name;
    table.sourceType = type;
    if (options.csvDelimiter) {
      table.csvDelimiter = options.csvDelimiter;
    }

    const existingModelResp = this.checkExistingModel(name);
    if (existingModelResp) {
      return existingModelResp;
    }

    table.lastUpdated = Date.now();

    const response = await this.addOrUpdateTable(
      table,
      !existingTable,
      !options.shouldNotProfile
    );
    if (response) return response;
    return ActionResponseFactory.getSuccessResponse("", table);
  }

  @DataModelerActions.PersistentTableAction()
  public async addOrSyncTableFromDB(
    { stateService }: PersistentTableStateActionArg,
    tableName?: string
  ) {
    const existingTable = stateService.getByField("tableName", tableName);
    const table = existingTable ? { ...existingTable } : getNewTable();

    table.name = table.tableName = tableName;
    table.sourceType = TableSourceType.DuckDB;

    await this.addOrUpdateTable(table, !existingTable, true);
  }

  @DataModelerActions.DerivedTableAction()
  @DataModelerActions.ResetStateToIdle(EntityType.Table)
  public async collectTableInfo(
    _: DerivedTableStateActionArg,
    tableId: string
  ): Promise<ActionResponse> {
    const persistentTable = this.dataModelerStateService.getEntityById(
      EntityType.Table,
      StateType.Persistent,
      tableId
    );
    const newDerivedTable: DerivedTableEntity = {
      id: tableId,
      type: EntityType.Table,
      status: EntityStatus.Profiling,
      lastUpdated: Date.now(),
      profiled: false,
    };

    if (!persistentTable) {
      return ActionResponseFactory.getEntityError(
        `No table found for ${tableId}`
      );
    }
    this.databaseActionQueue.clearQueue(tableId);
    this.dataModelerService.dispatch("clearColumnProfilePriority", [
      EntityType.Table,
      tableId,
    ]);

    try {
      this.dataModelerStateService.dispatch("setEntityStatus", [
        EntityType.Table,
        tableId,
        EntityStatus.Profiling,
      ]);
      await this.dataModelerStateService.dispatch("clearProfileSummary", [
        EntityType.Table,
        tableId,
      ]);

      await Promise.all(
        [
          async () => {
            newDerivedTable.profile = await this.databaseActionQueue.enqueue(
              {
                id: tableId,
                priority: DatabaseActionQueuePriority.TableImport,
              },
              "getProfileColumns",
              [persistentTable.tableName]
            );
            newDerivedTable.profile = newDerivedTable.profile.filter(
              (row) =>
                row.name !== "duckdb_schema" &&
                row.name !== "schema" &&
                row.name !== "root"
            );
          },
          async () =>
            (newDerivedTable.sizeInBytes =
              await this.databaseActionQueue.enqueue(
                {
                  id: tableId,
                  priority: DatabaseActionQueuePriority.TableProfile,
                },
                "getDestinationSize",
                [persistentTable.path]
              )),
          async () =>
            (newDerivedTable.cardinality =
              await this.databaseActionQueue.enqueue(
                {
                  id: tableId,
                  priority: DatabaseActionQueuePriority.TableProfile,
                },
                "getCardinalityOfTable",
                [persistentTable.tableName]
              )),
          async () =>
            (newDerivedTable.preview = await this.databaseActionQueue.enqueue(
              {
                id: tableId,
                priority: DatabaseActionQueuePriority.TableProfile,
              },
              "getFirstNOfTable",
              [persistentTable.tableName, SOURCE_PREVIEW_COUNT]
            )),
        ].map((asyncFunc) => asyncFunc())
      );

      this.dataModelerStateService.dispatch("updateEntity", [
        EntityType.Table,
        StateType.Derived,
        newDerivedTable,
      ]);
      await this.dataModelerService.dispatch("collectProfileColumns", [
        EntityType.Table,
        tableId,
      ]);
      this.dataModelerStateService.dispatch("markAsProfiled", [
        EntityType.Table,
        tableId,
        true,
      ]);
    } catch (err) {
      return ActionResponseFactory.getErrorResponse(err);
    }
  }

  @DataModelerActions.DerivedTableAction()
  public async refreshPreview(
    _: DerivedTableStateActionArg,
    tableId: string,
    tableName: string
  ): Promise<void> {
    this.dataModelerStateService.dispatch("updateTablePreview", [
      tableId,
      await this.dataModelerService.databaseActionQueue.enqueue(
        {
          id: tableId,
          priority: DatabaseActionQueuePriority.TableProfile,
        },
        "getFirstNOfTable",
        [tableName, SOURCE_PREVIEW_COUNT]
      ),
    ]);
  }

  @DataModelerActions.PersistentTableAction()
  public async updateTableName(
    { stateService }: PersistentTableStateActionArg,
    tableId: string,
    name: string
  ): Promise<ActionResponse> {
    const sanitizedNewName = sanitizeEntityName(extractTableName(name));
    const existingTable = stateService.getByField(
      "tableName",
      sanitizedNewName
    );

    if (existingTable) {
      return ActionResponseFactory.getExisingEntityError(
        `another source named "${existingTable.tableName}" already exists`
      );
    }

    const table = stateService.getById(tableId);
    const currentName = table.tableName;

    this.dataModelerStateService.dispatch("updateTableName", [
      tableId,
      sanitizedNewName,
    ]);
    this.databaseService.dispatch("renameTable", [
      currentName,
      sanitizedNewName,
    ]);
    return ActionResponseFactory.getSuccessResponse(
      `source ${currentName} renamed to ${sanitizedNewName}`
    );
  }

  @DataModelerActions.PersistentTableAction()
  public async validateTableName(
    { stateService }: PersistentTableStateActionArg,
    tableName: string
  ): Promise<ActionResponse> {
    const sanitizedTableName = sanitizeEntityName(extractTableName(tableName));
    const existingNames = stateService
      .getCurrentState()
      .entities.map((table) => table.tableName);

    const nonDuplicateName = getName(sanitizedTableName, existingNames);

    if (nonDuplicateName === sanitizedTableName) {
      return ActionResponseFactory.getSuccessResponse();
    } else {
      return ActionResponseFactory.getSuccessResponse(nonDuplicateName);
    }
  }

  @DataModelerActions.PersistentTableAction()
  public async dropTable(
    { stateService }: PersistentTableStateActionArg,
    tableName: string,
    removeOnly = false
  ): Promise<ActionResponse> {
    const table = stateService.getByField("tableName", tableName);
    if (!table) {
      return ActionResponseFactory.getEntityError(
        `No table found for ${tableName}`
      );
    }

    if (!removeOnly) {
      await this.databaseActionQueue.enqueue(
        { id: table.id, priority: DatabaseActionQueuePriority.TableImport },
        "dropTable",
        [table.tableName]
      );
    }
    this.notificationService.notify({
      message: `dropped table ${table.tableName}`,
      type: "info",
    });

    await this.dataModelerService.dispatch("deleteEntity", [
      EntityType.Table,
      table.id,
    ]);
  }

  @DataModelerActions.DerivedTableAction()
  public async syncTable(
    { stateService }: DerivedTableStateActionArg,
    tableId: string
  ) {
    let derivedTable = stateService.getById(tableId);
    const persistentTable = this.dataModelerStateService
      .getEntityStateService(EntityType.Table, StateType.Persistent)
      .getById(tableId);
    if (!derivedTable || !persistentTable) {
      return ActionResponseFactory.getEntityError(
        `No table found for ${tableId}`
      );
    }
    if (derivedTable.status === EntityStatus.Profiling) return;

    // check row count
    const newCardinality = await this.databaseActionQueue.enqueue(
      { id: tableId, priority: DatabaseActionQueuePriority.TableProfile },
      "getCardinalityOfTable",
      [persistentTable.tableName]
    );
    derivedTable = stateService.getById(tableId);
    if (newCardinality === derivedTable.cardinality) return;

    // check column count and names
    const newProfiles: Array<ProfileColumn> =
      await this.databaseActionQueue.enqueue(
        { id: tableId, priority: DatabaseActionQueuePriority.TableImport },
        "getProfileColumns",
        [persistentTable.tableName]
      );
    derivedTable = stateService.getById(tableId);
    if (newProfiles.length === derivedTable.profile.length) {
      const existingColumns = new Map<string, ProfileColumn>();
      derivedTable.profile.forEach((column) =>
        existingColumns.set(column.name, column)
      );

      if (
        newProfiles.every(
          (newProfileColumn) =>
            existingColumns.has(newProfileColumn.name) &&
            existingColumns.get(newProfileColumn.name).type ===
              newProfileColumn.type
        )
      ) {
        return;
      }
    }

    this.dataModelerStateService.dispatch("setEntityStatus", [
      EntityType.Table,
      tableId,
      EntityStatus.Profiling,
    ]);

    await this.dataModelerService.dispatch("collectTableInfo", [tableId]);
  }

  private async addOrUpdateTable(
    table: PersistentTableEntity,
    isNew: boolean,
    shouldProfile: boolean
  ): Promise<ActionResponse> {
    // get the original Table state if not new.
    let originalPersistentTable: PersistentTableEntity;
    if (!isNew) {
      originalPersistentTable = this.dataModelerStateService
        .getEntityStateService(EntityType.Table, StateType.Persistent)
        .getByField("tableName", table.name);
    }

    // update the new state
    if (isNew) {
      this.dataModelerStateService.dispatch("addEntity", [
        EntityType.Table,
        StateType.Persistent,
        table,
      ]);
    } else {
      this.dataModelerStateService.dispatch("updateEntity", [
        EntityType.Table,
        StateType.Persistent,
        table,
      ]);
    }

    let derivedTable: DerivedTableEntity;
    if (isNew) {
      derivedTable = getNewDerivedTable(table);
      derivedTable.status = EntityStatus.Importing;
      this.dataModelerStateService.dispatch("addEntity", [
        EntityType.Table,
        StateType.Derived,
        derivedTable,
      ]);
    } else {
      this.dataModelerStateService.dispatch("setEntityStatus", [
        EntityType.Table,
        table.id,
        EntityStatus.Importing,
      ]);
    }

    const response = await this.importTableDataByType(table);
    if (
      response?.status !== undefined &&
      response?.status === ActionStatus.Failure
    ) {
      if (isNew) {
        // Delete the table entirely.
        this.dataModelerStateService.dispatch("deleteEntity", [
          EntityType.Table,
          StateType.Derived,
          derivedTable.id,
        ]);
        // Fetch the persistent table in this instance
        // and delete
        const existingTable = this.dataModelerStateService
          .getEntityStateService(EntityType.Table, StateType.Persistent)
          .getByField("tableName", table.name);
        this.dataModelerStateService.dispatch("deleteEntity", [
          EntityType.Table,
          StateType.Persistent,
          existingTable.id,
        ]);
      } else {
        this.dataModelerStateService.dispatch("updateEntity", [
          EntityType.Table,
          StateType.Persistent,
          originalPersistentTable,
        ]);
        // Reset entity status to idle in the case where the table already exists.
        this.dataModelerStateService.dispatch("setEntityStatus", [
          EntityType.Table,
          table.id,
          EntityStatus.Idle,
        ]);
      }
      return response;
    }

    if (this.config.profileWithUpdate) {
      // this check should not hit else on false. hence it is nested
      if (shouldProfile) {
        await this.dataModelerService.dispatch("collectTableInfo", [table.id]);
      }
    } else {
      this.dataModelerStateService.dispatch("markAsProfiled", [
        EntityType.Table,
        table.id,
        false,
      ]);
    }
    this.dataModelerStateService.dispatch("setEntityStatus", [
      EntityType.Table,
      table.id,
      EntityStatus.Idle,
    ]);
  }

  private async importTableDataByType(
    table: PersistentTableEntity
  ): Promise<ActionResponse> {
    let response: ActionResponse;
    switch (table.sourceType) {
      case TableSourceType.ParquetFile:
        response = await this.databaseActionQueue.enqueue(
          { id: table.id, priority: DatabaseActionQueuePriority.TableImport },
          "importParquetFile",
          [table.path, table.tableName]
        );
        break;

      case TableSourceType.CSVFile:
        response = await this.databaseActionQueue.enqueue(
          { id: table.id, priority: DatabaseActionQueuePriority.TableImport },
          "importCSVFile",
          [table.path, table.tableName, table.csvDelimiter]
        );
        break;

      case TableSourceType.DuckDB:
        // table already exists. nothing to do here
        break;
    }
    if (response?.status === ActionStatus.Failure) {
      this.notificationService.notify({
        message: `failed to import ${table.name} from ${table.path}`,
        type: "error",
      });
    } else {
      this.notificationService.notify({
        message: `imported ${table.name}`,
        type: "info",
      });
    }
    return response;
  }

  private checkExistingModel(tableName: string): ActionResponse {
    const existingModel = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getByField("tableName", tableName);
    if (existingModel) {
      this.notificationService.notify({
        message: `Another model with the sanitised table name ${tableName} already exists`,
        type: "error",
      });
      return ActionResponseFactory.getExisingEntityError(
        `Another model with the sanitised table name ${tableName} already exists`
      );
    }
    return undefined;
  }
}
