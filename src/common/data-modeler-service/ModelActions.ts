import {DataModelerActions} from "$common/data-modeler-service/DataModelerActions";
import type {DataModelerState, Model} from "$lib/types";
import {ColumnarItemType} from "$common/data-modeler-state-service/ProfileColumnStateActions";
import {IDLE_STATUS, MODEL_PREVIEW_COUNT, RUNNING_STATUS} from "$common/constants";
import {sanitizeQuery} from "$lib/util/sanitize-query";
import type {NewModelParams} from "$common/data-modeler-state-service/ModelStateActions";

export class ModelActions extends DataModelerActions {
    public async addModel(currentState: DataModelerState, params: NewModelParams) {
        this.dataModelerStateService.dispatch("addModel", [params]);
    }

    public async updateModelQuery(currentState: DataModelerState, modelId: string, query: string): Promise<void> {
        const model = currentState.models.find(findModel => findModel.id === modelId);

        if (!model) {
            console.error(`No model found for ${modelId}`);
            return;
        }

        const sanitizedQuery = sanitizeQuery(query);

        this.dataModelerStateService.dispatch("updateModelQuery", [modelId, query, sanitizedQuery]);

        this.dataModelerStateService.dispatch("setTableStatus",
            [ColumnarItemType.Model, modelId, RUNNING_STATUS]);

        // validate query with the original query first.
        if (!await this.validateModelQuery(model, query)) {
            this.dataModelerStateService.dispatch("setTableStatus",
                [ColumnarItemType.Model, modelId, IDLE_STATUS]);
            return;
        }
        this.dataModelerStateService.dispatch("clearModelError", [model.id]);

        try {
            // create a view of the query for other analysis
            // re-sanitize query but do not remove casing, in case there is case-sensitive syntax 
            // in the query e.g. strftime(dt, '%I:%M:%S')
            await this.databaseService.dispatch("createViewOfQuery", [model.tableName, sanitizeQuery(query, false)]);
            await this.collectModelInfo(model);
        } catch (err) {
            console.error(err);
        }

        this.dataModelerStateService.dispatch("setTableStatus",
            [ColumnarItemType.Model, modelId, IDLE_STATUS]);
    }

    public async exportToParquet(currentState: DataModelerState, modelId: string, exportFile: string): Promise<void> {
        const model = currentState.models.find(findModel => findModel.id === modelId);
        const exportPath = await this.databaseService.dispatch("exportToParquet", [model.sanitizedQuery, exportFile]);
        await this.dataModelerStateService.dispatch("updateModelDestinationSize",
          [modelId, await this.databaseService.dispatch("getDestinationSize", [exportPath])]);
        this.notificationService.notify({ message: `exported ${exportPath}`, type: "info"})
    }

    public async updateModelName(currentState: DataModelerState, modelId: string, name: string): Promise<void> {
        this.dataModelerStateService.dispatch("updateModelName", [modelId, name]);
    }

    public async deleteModel(currentState: DataModelerState, modelId: string): Promise<void> {
        this.dataModelerStateService.dispatch("deleteModel", [modelId]);
    }

    public async moveModelDown(currentState: DataModelerState, modelId: string): Promise<void> {
        this.dataModelerStateService.dispatch("moveModelDown", [modelId]);
    }

    public async moveModelUp(currentState: DataModelerState, modelId: string): Promise<void> {
        this.dataModelerStateService.dispatch("moveModelUp", [modelId]);
    }

    private async validateModelQuery(model: Model, sanitizedQuery: string): Promise<boolean> {
        try {
            await this.databaseService.dispatch("validateQuery", [sanitizedQuery]);
        } catch (error) {
            if (error.message !== 'No statement to prepare!') {
                this.dataModelerStateService.dispatch("addModelError", [model.id, error.message]);
            }  else {
                this.dataModelerStateService.dispatch("clearModelQuery", [model.id]);
            }
            return false;
        }
        return true;
    }

    private async collectModelInfo(model: Model): Promise<void> {
        let profileColumns;
        try {
            // To get the profile columns, we'll select a single  value out of 
            // the view. This is also a good place to _test_ whether this query has any runtime errors, since
            // to get one result of the view, we'll need to run the underlying query itself.
            // FIXME: We should really start writing tests here!
            profileColumns = await this.databaseService.dispatch("getProfileColumns", [model.tableName])
        } catch (error) {
            this.dataModelerStateService.dispatch("addModelError", [model.id, error.message]);
            return;
        }
        // clear any model error if we get this far.
        this.dataModelerStateService.dispatch("clearModelError", [model.id]);
        
        this.dataModelerStateService.dispatch("updateModelProfileColumns",
            [model.id, profileColumns]);
        
        // retrieve the source table references from the query directly.
        this.dataModelerStateService.dispatch("getModelSourceTables", [model.id]);

        await Promise.all([
            async () => await this.dataModelerService.dispatch("collectProfileColumns",
                [model.id, ColumnarItemType.Model]),
            // TODO: add debouncing
            async () => this.dataModelerStateService.dispatch("updateModelPreview", [model.id,
                await this.databaseService.dispatch("getFirstNOfTable", [model.tableName, MODEL_PREVIEW_COUNT])]),
            async () => this.dataModelerStateService.dispatch("updateModelCardinality", [model.id,
                await this.databaseService.dispatch("getCardinalityOfTable", [model.tableName])]),
            async () => this.dataModelerStateService.dispatch("updateModelDestinationSize", [model.id,
                await this.databaseService.dispatch("getDestinationSize", [model.tableName])]),
        ].map(asyncFunc => asyncFunc()));
    }
}
