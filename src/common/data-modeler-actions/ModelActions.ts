import {DataModelerActions} from "$common/data-modeler-actions/DataModelerActions";
import type {DataModelerState, Model} from "$lib/types";
import {ColumnarItemType} from "$common/state-actions/ProfileColumnStateActions";
import {IDLE_STATUS, MODEL_PREVIEW_COUNT, RUNNING_STATUS} from "$common/constants";
import {sanitizeQuery} from "$lib/util/sanitize-query";

export class ModelActions extends DataModelerActions {
    public async updateQueryInformation(currentState: DataModelerState, id: string, query: string): Promise<void> {
        const model = currentState.queries.find(findModel => findModel.id === id);

        if (!model) {
            console.log(`No model found for ${id}`);
            return;
        }

        const sanitizedQuery = sanitizeQuery(query);

        this.dataModelerStateManager.dispatch("updateModelQuery", [id, query, sanitizedQuery]);

        this.dataModelerStateManager.dispatch("setDatasetStatus",
            [ColumnarItemType.Model, id, RUNNING_STATUS]);

        // validate query 1st
        if (!await this.validateModelQuery(model, sanitizedQuery)) {
            this.dataModelerStateManager.dispatch("setDatasetStatus",
                [ColumnarItemType.Model, id, IDLE_STATUS]);
            return;
        }
        this.dataModelerStateManager.dispatch("clearModelError", [model.id]);

        try {
            // create a view of the query for other analysis
            await this.databaseTableActions.createViewOfQuery(model.tableName, sanitizedQuery);

            await this.collectModelInfo(model);
        } catch (err) {
            console.log(err);
        }

        this.dataModelerStateManager.dispatch("setDatasetStatus",
            [ColumnarItemType.Model, id, IDLE_STATUS]);
    }

    private async validateModelQuery(model: Model, sanitizedQuery: string): Promise<boolean> {
        try {
            await this.databaseTableActions.validateQuery(sanitizedQuery);
        } catch (error) {
            if (error.message !== 'No statement to prepare!') {
                this.dataModelerStateManager.dispatch("addModelError", [model.id, error.message]);
            }  else {
                this.dataModelerStateManager.dispatch("clearModelQuery", [model.id]);
            }
            return false;
        }
        return true;
    }

    private async collectModelInfo(model: Model): Promise<void> {
        this.dataModelerStateManager.dispatch("updateModelProfileColumns",
            [model.id, await this.databaseTableActions.getProfileColumns(model.tableName)]);
        await Promise.all([
            async () => await this.dataModelerActionAPI.dispatch("collectProfileColumns",
                [model.id, ColumnarItemType.Model]),
            // TODO: add debouncing
            async () => this.dataModelerStateManager.dispatch("updateModelPreview",
                [model.id, await this.databaseTableActions.getFirstN(model.tableName, MODEL_PREVIEW_COUNT)]),
            async () => this.dataModelerStateManager.dispatch("updateModelCardinality",
                [model.id, await this.databaseTableActions.getCardinality(model.tableName)]),
            async () => this.dataModelerStateManager.dispatch("updateModelDestinationSize",
                [model.id, await this.databaseDataLoaderActions.getDestinationSize(model.tableName)]),
        ].map(asyncFunc => asyncFunc()));
    }
}
