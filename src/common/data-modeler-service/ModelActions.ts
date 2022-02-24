import { DataModelerActions } from "$common/data-modeler-service/DataModelerActions";
import type { DataModelerState } from "$lib/types";
import { MODEL_PREVIEW_COUNT } from "$common/constants";
import { sanitizeQuery } from "$lib/util/sanitize-query";
import type { NewModelParams } from "$common/data-modeler-state-service/ModelStateActions";
import type {
    PersistentModelEntity,
    PersistentModelStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import {
    AllStateTypes,
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
        this.dataModelerStateService.addEntities(EntityType.Model, [
            [StateType.Persistent, persistentModel],
            [StateType.Derived, getNewDerivedModel(persistentModel)],
        ], params.at);
    }

    @DataModelerActions.PersistentModelAction()
    public async updateModelQuery({stateService}: PersistentModelStateActionArg,
                                  modelId: string, query: string): Promise<void> {
        const model = stateService.getById(modelId);
        if (!model) {
            console.error(`No model found for ${modelId}`);
            return;
        }

        this.dataModelerStateService.dispatch("updateModelQuery",
            [modelId, query, sanitizeQuery(query)]);

        // validate query with the original query first.
        if (!await this.validateModelQuery(model, query)) {
            return;
        }
        this.dataModelerStateService.dispatch("clearModelError", [model.id]);

        try {
            // create a view of the query for other analysis
            // re-sanitize query but do not remove casing, in case there is case-sensitive syntax 
            // in the query e.g. strftime(dt, '%I:%M:%S')
            await this.databaseService.dispatch("createViewOfQuery",
                [model.tableName, sanitizeQuery(query, false)]);
            await this.dataModelerService.dispatch("collectModelInfo", [modelId]);
        } catch (err) {
            console.error(err);
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
            this.dataModelerStateService.dispatch("addModelError", [modelId, error.message]);
            return;
        }
        // clear any model error if we get this far.
        this.dataModelerStateService.dispatch("clearModelError", [modelId]);

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
        this.dataModelerStateService.deleteEntities(EntityType.Model, AllStateTypes, modelId);
    }

    @DataModelerActions.PersistentModelAction()
    public async moveModelDown(args: PersistentModelStateActionArg,
                               modelId: string): Promise<void> {
        this.dataModelerStateService.moveEntitiesDown(EntityType.Model, AllStateTypes, modelId);
    }

    @DataModelerActions.PersistentModelAction()
    public async moveModelUp(args: PersistentModelStateActionArg,
                             modelId: string): Promise<void> {
        this.dataModelerStateService.moveEntitiesUp(EntityType.Model, AllStateTypes, modelId);
    }

    private async validateModelQuery(model: PersistentModelEntity, sanitizedQuery: string): Promise<boolean> {
        try {
            await this.databaseService.dispatch("validateQuery", [sanitizedQuery]);
        } catch (error) {
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
