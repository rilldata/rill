import { DataModelerActions } from "$common/data-modeler-service/DataModelerActions";
import { MODEL_PREVIEW_COUNT } from "$common/constants";
import { sanitizeQuery } from "$lib/util/sanitize-query";
import type { NewModelParams } from "$common/data-modeler-state-service/ModelStateActions";
import type {
    PersistentModelEntity,
    PersistentModelStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import {
    EntityStatus,
    EntityType,
    StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
    DerivedModelStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import { getNewDerivedModel, getNewModel } from "$common/stateInstancesFactory";

export class ModelActions extends DataModelerActions {
    @DataModelerActions.PersistentModelAction()
    public async addModel(args: PersistentModelStateActionArg, params: NewModelParams) {
        const persistentModel = getNewModel(params);
        this.dataModelerStateService.dispatch("addEntity",
            [EntityType.Model, StateType.Persistent,
                persistentModel, params.at]);
        this.dataModelerStateService.dispatch("addEntity",
            [EntityType.Model, StateType.Derived,
                getNewDerivedModel(persistentModel), params.at]);
    }

    @DataModelerActions.PersistentModelAction()
    public async updateModelQuery({stateService}: PersistentModelStateActionArg,
                                  modelId: string, query: string): Promise<void> {
        const model = stateService.getById(modelId);
        const derivedModel = this.dataModelerStateService
            .getEntityById(EntityType.Model, StateType.Derived, modelId);
        if (!model) {
            console.error(`No model found for ${modelId}`);
            return;
        }

        const sanitizedQuery = sanitizeQuery(query);
        if (sanitizedQuery === derivedModel.sanitizedQuery) {
            return;
        }

        this.dataModelerStateService.dispatch("updateModelQuery", [modelId, query, sanitizedQuery]);
        this.dataModelerStateService.dispatch("updateModelSanitizedQuery", [modelId, sanitizedQuery]);


        // validate query with the original query first.
        if (!await this.validateModelQuery(model, query)) {
            return;
        }
        this.dataModelerStateService.dispatch("clearModelError", [model.id]);

        if (this.config.profileWithUpdate) {
            await this.dataModelerService.dispatch("collectModelInfo", [modelId]);
        } else {
            this.dataModelerStateService.dispatch("markAsProfiled",
                [EntityType.Model, modelId, false]);
        }
    }

    @DataModelerActions.DerivedModelAction()
    public async collectModelInfo({stateService}: DerivedModelStateActionArg,
                                  modelId: string): Promise<void> {
        const persistentModel = this.dataModelerStateService
            .getEntityById(EntityType.Model, StateType.Persistent, modelId);
        const model = stateService.getById(modelId);
        if (!model) {
            console.error(`No model found for ${modelId}`);
            return;
        }

        try {
            // create a view of the query for other analysis
            // re-sanitize query but do not remove casing, in case there is case-sensitive syntax
            // in the query e.g. strftime(dt, '%I:%M:%S')
            await this.databaseService.dispatch("createViewOfQuery",
                [persistentModel.tableName, sanitizeQuery(persistentModel.query, false)]);
        } catch (err) {
            console.error(err);
            return;
        }

        this.dataModelerStateService.dispatch("setTableStatus",
            [EntityType.Model, modelId, EntityStatus.Profiling]);

        let profileColumns;
        try {
            // To get the profile columns, we'll select a single  value out of
            // the view. This is also a good place to _test_ whether this query has any runtime errors, since
            // to get one result of the view, we'll need to run the underlying query itself.
            // FIXME: We should really start writing tests here!
            profileColumns = await this.databaseService.dispatch("getProfileColumns", [persistentModel.tableName])
        } catch (error) {
            console.log(error);
            this.dataModelerStateService.dispatch("addModelError", [modelId, error.message]);
            return;
        }
        // clear any model error if we get this far.
        this.dataModelerStateService.dispatch("clearModelError", [modelId]);

        // retrieve the source table references from the query directly.
        this.dataModelerStateService.dispatch("getModelSourceTables",
            [model.id, persistentModel.query]);

        this.dataModelerStateService.dispatch("updateModelProfileColumns",
            [modelId, profileColumns]);
        await Promise.all([
            async () => await this.dataModelerService.dispatch("collectProfileColumns",
                [EntityType.Model, modelId]),
            // TODO: add debouncing
            async () => this.dataModelerStateService.dispatch("updateModelPreview", [modelId,
                await this.databaseService.dispatch("getFirstNOfTable", [persistentModel.tableName, MODEL_PREVIEW_COUNT])]),
            async () => this.dataModelerStateService.dispatch("updateModelCardinality", [modelId,
                await this.databaseService.dispatch("getCardinalityOfTable", [persistentModel.tableName])]),
            async () => this.dataModelerStateService.dispatch("updateModelDestinationSize", [modelId,
                await this.databaseService.dispatch("getDestinationSize", [persistentModel.tableName])]),
        ].map(asyncFunc => asyncFunc()));

        this.dataModelerStateService.dispatch("markAsProfiled",
            [EntityType.Model, modelId, true]);
        this.dataModelerStateService.dispatch("setTableStatus",
            [EntityType.Model, modelId, EntityStatus.Idle]);
    }

    @DataModelerActions.PersistentModelAction()
    public async exportToParquet({stateService}: PersistentModelStateActionArg,
                                 modelId: string, exportFile: string): Promise<void> {
        const model = stateService.getById(modelId);
        const exportPath = await this.databaseService.dispatch("exportToParquet",
            [sanitizeQuery(model.query), exportFile]);
        await this.dataModelerStateService.dispatch("updateModelDestinationSize",
          [modelId, await this.databaseService.dispatch("getDestinationSize", [exportPath])]);
        this.notificationService.notify({ message: `exported ${exportPath}`, type: "info"})
    }

    @DataModelerActions.PersistentModelAction()
    public async updateModelName(args: PersistentModelStateActionArg,
                                 modelId: string, name: string): Promise<void> {
        this.dataModelerStateService.dispatch("updateModelName", [modelId, name]);
    }

    @DataModelerActions.PersistentModelAction()
    public async deleteModel(args: PersistentModelStateActionArg,
                             modelId: string): Promise<void> {
        this.dataModelerStateService.dispatch("deleteEntity",
            [EntityType.Model, StateType.Persistent, modelId]);
        this.dataModelerStateService.dispatch("deleteEntity",
            [EntityType.Model, StateType.Derived, modelId]);
    }

    @DataModelerActions.PersistentModelAction()
    public async moveModelDown(args: PersistentModelStateActionArg,
                               modelId: string): Promise<void> {
        this.dataModelerStateService.dispatch("moveEntityDown",
            [EntityType.Model, StateType.Persistent, modelId]);
        this.dataModelerStateService.dispatch("moveEntityDown",
            [EntityType.Model, StateType.Derived, modelId]);
    }

    @DataModelerActions.PersistentModelAction()
    public async moveModelUp(args: PersistentModelStateActionArg,
                             modelId: string): Promise<void> {
        this.dataModelerStateService.dispatch("moveEntityUp",
            [EntityType.Model, StateType.Persistent, modelId]);
        this.dataModelerStateService.dispatch("moveEntityUp",
            [EntityType.Model, StateType.Derived, modelId]);
    }

    private async validateModelQuery(model: PersistentModelEntity, sanitizedQuery: string): Promise<boolean> {
        try {
            await this.databaseService.dispatch("validateQuery", [sanitizedQuery]);
        } catch (error) {
            console.log(error);
            if (error.message !== 'No statement to prepare!') {
                this.dataModelerStateService.dispatch("addModelError", [model.id, error.message]);
            }  else {
                this.dataModelerStateService.dispatch("clearModelProfile", [model.id]);
            }
            return false;
        }
        return true;
    }
}
