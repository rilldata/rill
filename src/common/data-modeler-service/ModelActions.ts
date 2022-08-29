import { MODEL_PREVIEW_COUNT } from "$common/constants";
import { DataModelerActions } from "$common/data-modeler-service/DataModelerActions";
import type { ActionResponse } from "$common/data-modeler-service/response/ActionResponse";
import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";
import { ActionResponseFactory } from "$common/data-modeler-service/response/ActionResponseFactory";
import { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";
import type { DerivedModelStateActionArg } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import {
  EntityStatus,
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
  PersistentModelEntity,
  PersistentModelEntityService,
  PersistentModelStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type { NewModelParams } from "$common/data-modeler-state-service/ModelStateActions";
import { DatabaseActionQueuePriority } from "$common/priority-action-queue/DatabaseActionQueuePriority";
import {
  cleanModelName,
  getNewDerivedModel,
  getNewModel,
} from "$common/stateInstancesFactory";
import {
  extractTableName,
  sanitizeEntityName,
} from "$lib/util/extract-table-name";
import { sanitizeQuery } from "$lib/util/sanitize-query";

export enum FileExportType {
  Parquet = "exportToParquet",
  CSV = "exportToCsv",
}

export class ModelActions extends DataModelerActions {
  @DataModelerActions.PersistentModelAction()
  public async clearAllModels({ stateService }: PersistentModelStateActionArg) {
    stateService.getCurrentState().entities.forEach((table) => {
      this.dataModelerStateService.dispatch("deleteEntity", [
        EntityType.Model,
        StateType.Persistent,
        table.id,
      ]);
      this.dataModelerStateService.dispatch("deleteEntity", [
        EntityType.Model,
        StateType.Derived,
        table.id,
      ]);
    });
  }

  // Load all model queries and views. This is not persisted by duckdb!
  @DataModelerActions.PersistentModelAction()
  public async loadModels({ stateService }: PersistentModelStateActionArg) {
    const models = stateService.getCurrentState().entities;
    await Promise.all(
      models.map((model) =>
        this.databaseActionQueue.enqueue(
          { id: model.id, priority: DatabaseActionQueuePriority.ActiveModel },
          "createViewOfQuery",
          [model.tableName, sanitizeQuery(model.query, false)]
        )
      )
    );
  }

  @DataModelerActions.PersistentModelAction()
  public async addModel(
    { stateService }: PersistentModelStateActionArg,
    params: NewModelParams
  ) {
    const persistentModel = getNewModel(
      params,
      stateService.getCurrentState().modelNumber + 1
    );
    const duplicateResp = this.checkDuplicateModel(
      stateService,
      persistentModel.name,
      persistentModel.id
    );
    if (duplicateResp) {
      return duplicateResp;
    }

    this.dataModelerStateService.dispatch("incrementModelNumber", []);
    this.dataModelerStateService.dispatch("addEntity", [
      EntityType.Model,
      StateType.Persistent,
      persistentModel,
      params.at,
    ]);
    this.dataModelerStateService.dispatch("addEntity", [
      EntityType.Model,
      StateType.Derived,
      getNewDerivedModel(persistentModel),
      params.at,
    ]);
    if (persistentModel.query) {
      await this.dataModelerService.dispatch("updateModelQuery", [
        persistentModel.id,
        params.query,
      ]);
    }
    return persistentModel;
  }

  @DataModelerActions.PersistentModelAction()
  @DataModelerActions.ResetStateToIdle(EntityType.Model)
  public async updateModelQuery(
    { stateService }: PersistentModelStateActionArg,
    modelId: string,
    query: string,
    // force update the query even if the query didn't change
    // this is to update model when associated sources change
    force = false
  ): Promise<ActionResponse> {
    const model = stateService.getById(modelId);
    const derivedModel = this.dataModelerStateService.getEntityById(
      EntityType.Model,
      StateType.Derived,
      modelId
    );
    if (!model) {
      return ActionResponseFactory.getEntityError(
        `No model found for ${modelId}`
      );
    }

    const sanitizedQuery = sanitizeQuery(query);
    if (!force && sanitizedQuery === derivedModel.sanitizedQuery) {
      if (derivedModel.error) {
        return ActionResponseFactory.getModelQueryError(derivedModel.error);
      }
      return;
    }

    this.databaseActionQueue.clearQueue(modelId);
    this.dataModelerService.dispatch("clearColumnProfilePriority", [
      EntityType.Model,
      modelId,
    ]);
    await this.setModelStatus(modelId, EntityStatus.Validating);

    this.dataModelerStateService.dispatch("updateModelQuery", [modelId, query]);
    this.dataModelerStateService.dispatch("updateModelSanitizedQuery", [
      modelId,
      sanitizedQuery,
    ]);

    // validate query with the original query first.
    const validationResponse = await this.validateModelQuery(model, query);
    if (validationResponse) {
      return this.setModelError(modelId, validationResponse);
    }
    this.dataModelerStateService.dispatch("clearSourceTables", [modelId]);
    this.dataModelerStateService.dispatch("clearModelError", [modelId]);

    if (this.config.profileWithUpdate) {
      return await this.dataModelerService.dispatch("collectModelInfo", [
        modelId,
      ]);
    } else {
      this.dataModelerStateService.dispatch("markAsProfiled", [
        EntityType.Model,
        modelId,
        false,
      ]);
    }
  }

  @DataModelerActions.DerivedModelAction()
  @DataModelerActions.ResetStateToIdle(EntityType.Model)
  public async collectModelInfo(
    { stateService }: DerivedModelStateActionArg,
    modelId: string
  ): Promise<ActionResponse> {
    const persistentModel = this.dataModelerStateService.getEntityById(
      EntityType.Model,
      StateType.Persistent,
      modelId
    );
    const model = stateService.getById(modelId);
    if (!model) {
      return ActionResponseFactory.getEntityError(
        `No model found for ${modelId}`
      );
    }
    if (!model.sanitizedQuery) return;
    this.databaseActionQueue.clearQueue(modelId);

    this.dataModelerService.dispatch("clearColumnProfilePriority", [
      EntityType.Model,
      modelId,
    ]);

    try {
      // create a view of the query for other analysis
      // re-sanitize query but do not remove casing, in case there is case-sensitive syntax
      // in the query e.g. strftime(dt, '%I:%M:%S')
      await this.databaseActionQueue.enqueue(
        { id: modelId, priority: DatabaseActionQueuePriority.ActiveModel },
        "createViewOfQuery",
        [persistentModel.tableName, sanitizeQuery(persistentModel.query, false)]
      );
    } catch (error) {
      return this.setModelError(
        modelId,
        ActionResponseFactory.getModelQueryError(error.message)
      );
    }

    await this.setModelStatus(modelId, EntityStatus.Profiling);

    let profileColumns;
    try {
      // To get the profile columns, we'll select a single  value out of
      // the view. This is also a good place to _test_ whether this query has any runtime errors, since
      // to get one result of the view, we'll need to run the underlying query itself.
      // FIXME: We should really start writing tests here!
      profileColumns = await this.databaseActionQueue.enqueue(
        { id: modelId, priority: DatabaseActionQueuePriority.ActiveModel },
        "getProfileColumns",
        [persistentModel.tableName]
      );
    } catch (error) {
      return this.setModelError(
        modelId,
        ActionResponseFactory.getModelQueryError(error.message)
      );
    }
    // clear any model error if we get this far.
    this.dataModelerStateService.dispatch("clearModelError", [modelId]);

    // retrieve the source table references from the query directly.
    this.dataModelerStateService.dispatch("getModelSourceTables", [
      model.id,
      persistentModel.query,
    ]);

    // check the sanitizeQuery. If it matches a simple select *, copy over the
    // source profile
    const derivedModel = this.dataModelerStateService.getEntityById(
      EntityType.Model,
      StateType.Derived,
      modelId
    );

    const sourceTableName = derivedModel.sources?.[0]?.name;
    const canCopyExistingProfile =
      sourceTableName &&
      derivedModel?.sanitizedQuery === `select * from ${sourceTableName}`;

    if (canCopyExistingProfile) {
      /** copy over the source profile columns here if profiling is done */
      // get the associated derived table
      //const persistentTable
      const table = this.dataModelerStateService
        .getEntityStateService(EntityType.Table, StateType.Persistent)
        .getByField("tableName", sanitizeEntityName(sourceTableName));

      const derivedTable = this.dataModelerStateService.getEntityById(
        EntityType.Table,
        StateType.Derived,
        table.id
      );

      /** if the source table has been profiled, we will copy over the relevant
       * state parts from the derived source.
       */
      if (derivedTable.profiled) {
        this.dataModelerStateService.dispatch("updateModelCardinality", [
          modelId,
          derivedTable.cardinality,
        ]);
        this.dataModelerStateService.dispatch("updateModelProfileColumns", [
          modelId,
          derivedTable.profile,
        ]);
        /** enqueue a preview table since we don't have one yet. */
        this.dataModelerStateService.dispatch("updateModelPreview", [
          modelId,
          await this.databaseActionQueue.enqueue(
            {
              id: modelId,
              priority: DatabaseActionQueuePriority.ActiveModel,
            },
            "getFirstNOfTable",
            [persistentModel.tableName, MODEL_PREVIEW_COUNT]
          ),
        ]),
          this.dataModelerStateService.dispatch("markAsProfiled", [
            EntityType.Model,
            modelId,
            true,
          ]);

        return;
      }
    }

    this.dataModelerStateService.dispatch("updateModelProfileColumns", [
      modelId,
      profileColumns,
    ]);
    // catch "cancelled query" error.
    try {
      await Promise.all(
        [
          // We start the query queue by first updating the model preview. This is the user's
          // first bit of intuition building while the rest of the dataset profiles.
          // If there is something obviously wrong, they can catch it here first.
          // TODO: add debouncing
          async () =>
            this.dataModelerStateService.dispatch("updateModelPreview", [
              modelId,
              await this.databaseActionQueue.enqueue(
                {
                  id: modelId,
                  priority: DatabaseActionQueuePriority.ActiveModel,
                },
                "getFirstNOfTable",
                [persistentModel.tableName, MODEL_PREVIEW_COUNT]
              ),
            ]),
          // get the total number of rows first, since many parts of the iterative profiling
          // require this number as the denominator (e.g. the top k and the null %s)
          async () =>
            this.dataModelerStateService.dispatch("updateModelCardinality", [
              modelId,
              await this.databaseActionQueue.enqueue(
                {
                  id: modelId,
                  priority: DatabaseActionQueuePriority.ActiveModelProfile,
                },
                "getCardinalityOfTable",
                [persistentModel.tableName]
              ),
            ]),
          async () =>
            await this.dataModelerService.dispatch("collectProfileColumns", [
              EntityType.Model,
              modelId,
            ]),
          async () =>
            this.dataModelerStateService.dispatch(
              "updateModelDestinationSize",
              [
                modelId,
                await this.databaseActionQueue.enqueue(
                  {
                    id: modelId,
                    priority: DatabaseActionQueuePriority.ActiveModelProfile,
                  },
                  "getDestinationSize",
                  [persistentModel.tableName]
                ),
              ]
            ),
        ].map((asyncFunc) => asyncFunc())
      );
    } catch (err) {
      return this.setModelError(
        modelId,
        ActionResponseFactory.getErrorResponse(err)
      );
    }

    this.dataModelerStateService.dispatch("markAsProfiled", [
      EntityType.Model,
      modelId,
      true,
    ]);
  }

  @DataModelerActions.PersistentModelAction()
  @DataModelerActions.ResetStateToIdle(EntityType.Model)
  public async exportToParquet(
    { stateService }: PersistentModelStateActionArg,
    modelId: string,
    exportFile: string
  ): Promise<void> {
    await this.exportToFile(
      stateService,
      modelId,
      exportFile,
      FileExportType.Parquet
    );
  }

  @DataModelerActions.PersistentModelAction()
  @DataModelerActions.ResetStateToIdle(EntityType.Model)
  public async exportToCsv(
    { stateService }: PersistentModelStateActionArg,
    modelId: string,
    exportFile: string
  ): Promise<void> {
    await this.exportToFile(
      stateService,
      modelId,
      exportFile,
      FileExportType.CSV
    );
  }

  @DataModelerActions.PersistentModelAction()
  public async updateModelName(
    { stateService }: PersistentModelStateActionArg,
    modelId: string,
    name: string
  ): Promise<ActionResponse> {
    const existingModel = stateService
      .getCurrentState()
      .entities.find(
        (model) =>
          cleanModelName(model.name).toLowerCase() === name.toLowerCase() &&
          model.id !== modelId
      );

    if (existingModel) {
      return ActionResponseFactory.getExisingEntityError(
        `Another model with the name ${name} already exists`
      );
    }

    const existingTable = this.dataModelerStateService
      .getEntityStateService(EntityType.Table, StateType.Persistent)
      .getByField("tableName", sanitizeEntityName(extractTableName(name)));

    if (existingTable) {
      return ActionResponseFactory.getExisingEntityError(
        `Another table with the sanitised table name ${existingTable.tableName} already exists`
      );
    }

    const model = stateService.getById(modelId);
    const currentName = model.tableName;
    const sanitizedModelName = cleanModelName(name);

    this.dataModelerStateService.dispatch("updateModelName", [
      modelId,
      sanitizedModelName,
    ]);

    return ActionResponseFactory.getSuccessResponse(
      `model ${currentName} renamed to ${sanitizedModelName}`
    );
  }

  @DataModelerActions.PersistentModelAction()
  public async deleteModel(
    { stateService }: PersistentModelStateActionArg,
    modelId: string
  ): Promise<ActionResponseFactory> {
    const model = stateService.getById(modelId);
    if (!model) {
      return ActionResponseFactory.getEntityError(
        `No model found for ${modelId}`
      );
    }
    await this.dataModelerService.dispatch("deleteEntity", [
      EntityType.Model,
      modelId,
    ]);
  }

  @DataModelerActions.PersistentModelAction()
  public async moveModelDown(
    args: PersistentModelStateActionArg,
    modelId: string
  ): Promise<void> {
    this.dataModelerStateService.dispatch("moveEntityDown", [
      EntityType.Model,
      StateType.Persistent,
      modelId,
    ]);
    this.dataModelerStateService.dispatch("moveEntityDown", [
      EntityType.Model,
      StateType.Derived,
      modelId,
    ]);
  }

  @DataModelerActions.PersistentModelAction()
  public async moveModelUp(
    args: PersistentModelStateActionArg,
    modelId: string
  ): Promise<void> {
    this.dataModelerStateService.dispatch("moveEntityUp", [
      EntityType.Model,
      StateType.Persistent,
      modelId,
    ]);
    this.dataModelerStateService.dispatch("moveEntityUp", [
      EntityType.Model,
      StateType.Derived,
      modelId,
    ]);
  }

  private async validateModelQuery(
    model: PersistentModelEntity,
    sanitizedQuery: string
  ): Promise<ActionResponse> {
    try {
      await this.databaseActionQueue.enqueue(
        { id: model.id, priority: DatabaseActionQueuePriority.ActiveModel },
        "validateQuery",
        [sanitizedQuery]
      );
    } catch (error) {
      if (error.message !== "No statement to prepare!") {
        return ActionResponseFactory.getModelQueryError(error.message);
      } else {
        this.dataModelerStateService.dispatch("clearModelProfile", [model.id]);
        return ActionResponseFactory.getSuccessResponse();
      }
    }
    return undefined;
  }

  private async exportToFile(
    stateService: PersistentModelEntityService,
    modelId: string,
    exportFile: string,
    exportType: FileExportType
  ) {
    const model = stateService.getById(modelId);
    await this.setModelStatus(modelId, EntityStatus.Exporting);
    const exportPath = (await this.databaseService.dispatch(exportType, [
      sanitizeQuery(model.query, false),
      exportFile,
    ])) as string;
    await this.dataModelerStateService.dispatch("updateModelDestinationSize", [
      modelId,
      (await this.databaseService.dispatch("getDestinationSize", [
        exportPath,
      ])) as number,
    ]);
  }

  private setModelStatus(modelId: string, status: EntityStatus) {
    return this.dataModelerStateService.dispatch("setEntityStatus", [
      EntityType.Model,
      modelId,
      status,
    ]);
  }

  private async setModelError(modelId: string, response: ActionResponse) {
    if (
      response.status === ActionStatus.Failure &&
      response.messages[0]?.errorType === ActionErrorType.ModelQuery
    ) {
      // store only model errors. other errors are not to be seen by the user
      this.dataModelerStateService.dispatch("addModelError", [
        modelId,
        response.messages[0].message,
      ]);
    } else {
      this.dataModelerStateService.dispatch("clearModelError", [modelId]);
    }
    return response;
  }

  private checkDuplicateModel(
    stateService: PersistentModelEntityService,
    name: string,
    id: string
  ): ActionResponse {
    name = cleanModelName(name);
    const existing = stateService
      .getCurrentState()
      .entities.find(
        (model) =>
          cleanModelName(model.name).toLowerCase() === name.toLowerCase() &&
          model.id !== id
      );
    const existingTable = this.dataModelerStateService
      .getEntityStateService(EntityType.Table, StateType.Persistent)
      .getByField("tableName", sanitizeEntityName(extractTableName(name)));

    if (existing) {
      this.notificationService.notify({
        message: `Another model with the name ${name} already exists`,
        type: "error",
      });
      return ActionResponseFactory.getExisingEntityError(
        `Another model with the name ${name} already exists`
      );
    } else if (existingTable) {
      this.notificationService.notify({
        message: `Another table with the sanitised table name ${existingTable.tableName} already exists`,
        type: "error",
      });
      return ActionResponseFactory.getExisingEntityError(
        `Another table with the sanitised table name ${existingTable.tableName} already exists`
      );
    }
    return undefined;
  }
}
