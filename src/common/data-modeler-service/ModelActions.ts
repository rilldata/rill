import {DataModelerActions} from "$common/data-modeler-service/DataModelerActions";
import type {DataModelerState, Model} from "$lib/types";
import {ColumnarItemType} from "$common/data-modeler-state-service/ProfileColumnStateActions";
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

        this.dataModelerStateService.dispatch("updateModelQuery", [id, query, sanitizedQuery]);

        this.dataModelerStateService.dispatch("setDatasetStatus",
            [ColumnarItemType.Model, id, RUNNING_STATUS]);

        // validate query 1st
        if (!await this.validateModelQuery(model, sanitizedQuery)) {
            this.dataModelerStateService.dispatch("setDatasetStatus",
                [ColumnarItemType.Model, id, IDLE_STATUS]);
            return;
        }
        this.dataModelerStateService.dispatch("clearModelError", [model.id]);

        try {
            // create a view of the query for other analysis
            await this.databaseService.dispatch("createViewOfQuery", [model.tableName, sanitizedQuery]);

            await this.collectModelInfo(model);
        } catch (err) {
            console.log(err);
        }

        this.dataModelerStateService.dispatch("setDatasetStatus",
            [ColumnarItemType.Model, id, IDLE_STATUS]);
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
        this.dataModelerStateService.dispatch("updateModelProfileColumns",
            [model.id, await this.databaseService.dispatch("getProfileColumns", [model.tableName])]);
        await Promise.all([
            async () => await this.dataModelerActionAPI.dispatch("collectProfileColumns",
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
