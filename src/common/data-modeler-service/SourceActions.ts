import { DataModelerActions } from "./DataModelerActions";
import {
  FILE_EXTENSION_TO_SOURCE_TYPE,
  ProfileColumn,
  SourceType,
} from "$lib/types";
import {
  getNewDerivedSource,
  getNewSource,
} from "$common/stateInstancesFactory";
import {
  extractFileExtension,
  getSourceNameFromFile,
  INVALID_CHARS,
} from "$lib/util/extract-source-name";
import type {
  PersistentSourceEntity,
  PersistentSourceStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/PersistentSourceEntityService";
import {
  EntityStatus,
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
  DerivedSourceEntity,
  DerivedSourceStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/DerivedSourceEntityService";
import { DatabaseActionQueuePriority } from "$common/priority-action-queue/DatabaseActionQueuePriority";
import { existsSync } from "fs";
import { ActionResponseFactory } from "$common/data-modeler-service/response/ActionResponseFactory";
import {
  ActionResponse,
  ActionStatus,
} from "$common/data-modeler-service/response/ActionResponse";

export interface ImportSourceOptions {
  csvDelimiter?: string;
}

export class SourceActions extends DataModelerActions {
  @DataModelerActions.PersistentSourceAction()
  public async clearAllSources({
    stateService,
  }: PersistentSourceStateActionArg): Promise<void> {
    stateService.getCurrentState().entities.forEach((source) => {
      this.dataModelerStateService.dispatch("deleteEntity", [
        EntityType.Source,
        StateType.Persistent,
        source.id,
      ]);
      this.dataModelerStateService.dispatch("deleteEntity", [
        EntityType.Source,
        StateType.Derived,
        source.id,
      ]);
    });
  }

  @DataModelerActions.PersistentSourceAction()
  public async addOrUpdateSourceFromFile(
    { stateService }: PersistentSourceStateActionArg,
    path: string,
    sourceName?: string,
    options: ImportSourceOptions = {}
  ): Promise<ActionResponse> {
    const name = getSourceNameFromFile(path, sourceName);
    const type = FILE_EXTENSION_TO_SOURCE_TYPE[extractFileExtension(path)];

    if (!existsSync(path)) {
      return ActionResponseFactory.getImportSourceError(
        `File ${path} does not exist`
      );
    }

    if (type === undefined) {
      return ActionResponseFactory.getImportSourceError(`Invalid file type`);
    }
    if (sourceName && INVALID_CHARS.test(sourceName)) {
      return ActionResponseFactory.getImportSourceError(
        `Input source name has invalid characters`
      );
    }

    const existingSource = stateService.getByField("sourceName", name);
    const source = existingSource ? { ...existingSource } : getNewSource();

    source.path = path;
    source.name = name;
    source.sourceName = name;
    source.sourceType = type;
    if (options.csvDelimiter) {
      source.csvDelimiter = options.csvDelimiter;
    }

    const existingModelResp = this.checkExistingModel(name);
    if (existingModelResp) {
      return existingModelResp;
    }

    source.lastUpdated = Date.now();

    return await this.addOrUpdateSource(source, !existingSource);
  }

  @DataModelerActions.PersistentSourceAction()
  public async addOrSyncSourceFromDB(
    { stateService }: PersistentSourceStateActionArg,
    sourceName?: string
  ) {
    const existingSource = stateService.getByField("sourceName", sourceName);
    const source = existingSource ? { ...existingSource } : getNewSource();

    source.name = source.sourceName = sourceName;
    source.sourceType = SourceType.DuckDB;

    await this.addOrUpdateSource(source, !existingSource);
  }

  @DataModelerActions.DerivedSourceAction()
  @DataModelerActions.ResetStateToIdle(EntityType.Source)
  public async collectSourceInfo(
    { stateService }: DerivedSourceStateActionArg,
    sourceId: string
  ): Promise<ActionResponse> {
    const persistentSource = this.dataModelerStateService.getEntityById(
      EntityType.Source,
      StateType.Persistent,
      sourceId
    );
    const newDerivedSource: DerivedSourceEntity = {
      id: sourceId,
      type: EntityType.Source,
      status: EntityStatus.Profiling,
      lastUpdated: Date.now(),
      profiled: false,
    };

    if (!persistentSource) {
      return ActionResponseFactory.getEntityError(
        `No source found for ${sourceId}`
      );
    }
    this.databaseActionQueue.clearQueue(sourceId);

    try {
      this.dataModelerStateService.dispatch("setEntityStatus", [
        EntityType.Source,
        sourceId,
        EntityStatus.Profiling,
      ]);
      await this.dataModelerStateService.dispatch("clearProfileSummary", [
        EntityType.Source,
        sourceId,
      ]);

      await Promise.all(
        [
          async () => {
            newDerivedSource.profile = await this.databaseActionQueue.enqueue(
              {
                id: sourceId,
                priority: DatabaseActionQueuePriority.SourceImport,
              },
              "getProfileColumns",
              [persistentSource.sourceName]
            );
            newDerivedSource.profile = newDerivedSource.profile.filter(
              (row) =>
                row.name !== "duckdb_schema" &&
                row.name !== "schema" &&
                row.name !== "root"
            );
          },
          async () =>
            (newDerivedSource.sizeInBytes =
              await this.databaseActionQueue.enqueue(
                {
                  id: sourceId,
                  priority: DatabaseActionQueuePriority.SourceProfile,
                },
                "getDestinationSize",
                [persistentSource.path]
              )),
          async () =>
            (newDerivedSource.cardinality =
              await this.databaseActionQueue.enqueue(
                {
                  id: sourceId,
                  priority: DatabaseActionQueuePriority.SourceProfile,
                },
                "getCardinalityOfTable",
                [persistentSource.sourceName]
              )),
          async () =>
            (newDerivedSource.preview = await this.databaseActionQueue.enqueue(
              {
                id: sourceId,
                priority: DatabaseActionQueuePriority.SourceProfile,
              },
              "getFirstNOfTable",
              [persistentSource.sourceName]
            )),
        ].map((asyncFunc) => asyncFunc())
      );

      this.dataModelerStateService.dispatch("updateEntity", [
        EntityType.Source,
        StateType.Derived,
        newDerivedSource,
      ]);
      await this.dataModelerService.dispatch("collectProfileColumns", [
        EntityType.Source,
        sourceId,
      ]);
      this.dataModelerStateService.dispatch("markAsProfiled", [
        EntityType.Source,
        sourceId,
        true,
      ]);
    } catch (err) {
      return ActionResponseFactory.getErrorResponse(err);
    }
  }

  @DataModelerActions.PersistentSourceAction()
  public async dropSource(
    { stateService }: PersistentSourceStateActionArg,
    sourceName: string,
    removeOnly = false
  ): Promise<ActionResponse> {
    const source = stateService.getByField("sourceName", sourceName);
    if (!source) {
      return ActionResponseFactory.getEntityError(
        `No source found for ${sourceName}`
      );
    }

    if (!removeOnly) {
      await this.databaseActionQueue.enqueue(
        { id: source.id, priority: DatabaseActionQueuePriority.SourceImport },
        "dropTable",
        [source.sourceName]
      );
    }
    this.notificationService.notify({
      message: `dropped source ${source.sourceName}`,
      type: "info",
    });

    await this.dataModelerService.dispatch("deleteEntity", [
      EntityType.Source,
      source.id,
    ]);
  }

  @DataModelerActions.DerivedSourceAction()
  public async syncSource(
    { stateService }: DerivedSourceStateActionArg,
    sourceId: string
  ) {
    let derivedSource = stateService.getById(sourceId);
    const persistentSource = this.dataModelerStateService
      .getEntityStateService(EntityType.Source, StateType.Persistent)
      .getById(sourceId);
    if (!derivedSource || !persistentSource) {
      return ActionResponseFactory.getEntityError(
        `No source found for ${sourceId}`
      );
    }
    if (derivedSource.status === EntityStatus.Profiling) return;

    // check row count
    const newCardinality = await this.databaseActionQueue.enqueue(
      { id: sourceId, priority: DatabaseActionQueuePriority.SourceProfile },
      "getCardinalityOfTable",
      [persistentSource.sourceName]
    );
    derivedSource = stateService.getById(sourceId);
    if (newCardinality === derivedSource.cardinality) return;

    // check column count and names
    const newProfiles: Array<ProfileColumn> =
      await this.databaseActionQueue.enqueue(
        { id: sourceId, priority: DatabaseActionQueuePriority.SourceImport },
        "getProfileColumns",
        [persistentSource.sourceName]
      );
    derivedSource = stateService.getById(sourceId);
    if (newProfiles.length === derivedSource.profile.length) {
      const existingColumns = new Map<string, ProfileColumn>();
      derivedSource.profile.forEach((column) =>
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
      EntityType.Source,
      sourceId,
      EntityStatus.Profiling,
    ]);

    await this.dataModelerService.dispatch("collectSourceInfo", [sourceId]);
  }

  private async addOrUpdateSource(
    source: PersistentSourceEntity,
    isNew: boolean
  ): Promise<ActionResponse> {
    // get the original Source state if not new.
    let originalPersistentSource: PersistentSourceEntity;
    if (!isNew) {
      originalPersistentSource = this.dataModelerStateService
        .getEntityStateService(EntityType.Source, StateType.Persistent)
        .getByField("sourceName", source.name);
    }

    // update the new state
    if (isNew) {
      this.dataModelerStateService.dispatch("addEntity", [
        EntityType.Source,
        StateType.Persistent,
        source,
      ]);
    } else {
      this.dataModelerStateService.dispatch("updateEntity", [
        EntityType.Source,
        StateType.Persistent,
        source,
      ]);
    }

    let derivedSource: DerivedSourceEntity;
    if (isNew) {
      derivedSource = getNewDerivedSource(source);
      derivedSource.status = EntityStatus.Importing;
      this.dataModelerStateService.dispatch("addEntity", [
        EntityType.Source,
        StateType.Derived,
        derivedSource,
      ]);
    } else {
      this.dataModelerStateService.dispatch("setEntityStatus", [
        EntityType.Source,
        source.id,
        EntityStatus.Importing,
      ]);
    }
    this.dataModelerStateService.dispatch("addOrUpdateSourceToState", [
      source,
      isNew,
    ]);

    const response = await this.importSourceDataByType(source);
    if (
      response?.status !== undefined &&
      response?.status === ActionStatus.Failure
    ) {
      if (isNew) {
        // Delete the source entirely.
        this.dataModelerStateService.dispatch("deleteEntity", [
          EntityType.Source,
          StateType.Derived,
          derivedSource.id,
        ]);
        // Fetch the persistent source in this instance
        // and delete
        const existingSource = this.dataModelerStateService
          .getEntityStateService(EntityType.Source, StateType.Persistent)
          .getByField("sourceName", source.name);
        this.dataModelerStateService.dispatch("deleteEntity", [
          EntityType.Source,
          StateType.Persistent,
          existingSource.id,
        ]);
      } else {
        this.dataModelerStateService.dispatch("updateEntity", [
          EntityType.Source,
          StateType.Persistent,
          originalPersistentSource,
        ]);
        // Reset entity status to idle in the case where the source already exists.
        this.dataModelerStateService.dispatch("setEntityStatus", [
          EntityType.Source,
          source.id,
          EntityStatus.Idle,
        ]);
      }
      return response;
    }

    if (this.config.profileWithUpdate) {
      await this.dataModelerService.dispatch("collectSourceInfo", [source.id]);
    } else {
      this.dataModelerStateService.dispatch("markAsProfiled", [
        EntityType.Source,
        source.id,
        false,
      ]);
    }
    this.dataModelerStateService.dispatch("setEntityStatus", [
      EntityType.Source,
      source.id,
      EntityStatus.Idle,
    ]);
  }

  private async importSourceDataByType(
    source: PersistentSourceEntity
  ): Promise<ActionResponse> {
    let response: ActionResponse;
    switch (source.sourceType) {
      case SourceType.ParquetFile:
        response = await this.databaseActionQueue.enqueue(
          { id: source.id, priority: DatabaseActionQueuePriority.SourceImport },
          "importParquetFile",
          [source.path, source.sourceName]
        );
        break;

      case SourceType.CSVFile:
        response = await this.databaseActionQueue.enqueue(
          { id: source.id, priority: DatabaseActionQueuePriority.SourceImport },
          "importCSVFile",
          [source.path, source.sourceName, source.csvDelimiter]
        );
        break;

      case SourceType.DuckDB:
        // source already exists. nothing to do here
        break;
    }
    if (response?.status === ActionStatus.Failure) {
      this.notificationService.notify({
        message: `failed to import ${source.name} from ${source.path}`,
        type: "error",
      });
    } else {
      this.notificationService.notify({
        message: `imported ${source.name}`,
        type: "info",
      });
    }
    return response;
  }

  private checkExistingModel(sourceName: string): ActionResponse {
    const existingModel = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getByField("sourceName", sourceName);
    if (existingModel) {
      this.notificationService.notify({
        message: `Another model with the sanitised source name ${sourceName} already exists`,
        type: "error",
      });
      return ActionResponseFactory.getExisingEntityError(
        `Another model with the sanitised source name ${sourceName} already exists`
      );
    }
    return undefined;
  }
}
