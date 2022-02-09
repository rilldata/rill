import {StateActions} from "./StateActions";
import type {DataModelerState, Model, ProfileColumn} from "$lib/types";
import {newQuery} from "$common/dataFactory";

export interface NewModelParams {
    query?: string;
    name?: string;
    at?: number;
    makeActive?: boolean;
}

export class ModelStateActions extends StateActions {
    public addModel(draftState: DataModelerState, params: NewModelParams): void {
        const newModel = newQuery({query: params.query, name: params.name});
        if (params.at !== undefined) {
            draftState.queries.splice(params.at, 0, newModel);
        } else {
            draftState.queries.push(newModel);
            if (params.makeActive) {
                draftState.activeAsset = {
                    id: newModel.id,
                    assetType: "model"
                };
            }
        }
    }

    public addModelError(draftState: DataModelerState, modelId: string, message: string): void {
        ModelStateActions.updateModelField(draftState, modelId, "error", message);
    }

    public clearModelError(draftState: DataModelerState, modelId: string): void {
        ModelStateActions.updateModelField(draftState, modelId, "error", undefined);
    }

    public clearModelQuery(draftState: DataModelerState, modelId: string): void {
        const model = ModelStateActions.getModel(draftState, modelId);
        model.sizeInBytes = undefined;
        model.destinationProfile = undefined;
        model.preview = undefined;
        model.profile = undefined;
    }

    public updateModelQuery(draftState: DataModelerState, modelId: string, query: string, sanitizedQuery: string): void {
        const model = ModelStateActions.getModel(draftState, modelId);
        model.query = query;
        model.sanitizedQuery = sanitizedQuery;
    }

    public updateModelProfileColumns(draftState: DataModelerState, modelId: string, profileColumns: Array<ProfileColumn>): void {
        ModelStateActions.updateModelField(draftState, modelId, "profile", profileColumns);
    }

    public updateModelPreview(draftState: DataModelerState, modelId: string, preview: Array<any>): void {
        ModelStateActions.updateModelField(draftState, modelId, "preview", preview);
    }

    public updateModelCardinality(draftState: DataModelerState, modelId: string, cardinality: number): void {
        ModelStateActions.updateModelField(draftState, modelId, "cardinality", cardinality);
    }

    public updateModelDestinationSize(draftState: DataModelerState, modelId: string, sizeInBytes: number): void {
        ModelStateActions.updateModelField(draftState, modelId, "sizeInBytes", sizeInBytes);
    }

    public updateModelName(draftState: DataModelerState, modelId: string, name: string): void {
        ModelStateActions.updateModelField(draftState, modelId, "name", `${name}.sql`);
    }

    public deleteModel(draftState: DataModelerState, modelId: string): void {
        const index = draftState.queries.findIndex(model => model.id === modelId);
        if (index === -1) return;
        draftState.queries.splice(index, 1);
    }

    public moveModelDown(draftState: DataModelerState, modelId: string): void {
        const index = draftState.queries.findIndex(model => model.id === modelId);
        if (index === -1 || index === draftState.queries.length - 1) return;

        [draftState.queries[index], draftState.queries[index + 1]] =
            [draftState.queries[index + 1], draftState.queries[index]];
    }

    public moveModelUp(draftState: DataModelerState, modelId: string): void {
        const index = draftState.queries.findIndex(model => model.id === modelId);
        if (index === -1 || index === 0) return;

        [draftState.queries[index], draftState.queries[index - 1]] =
            [draftState.queries[index - 1], draftState.queries[index]];
    }

    private static updateModelField<Field extends keyof Model>(draftState: DataModelerState, modelId: string,
                                                               field: Field, value: Model[Field]): void {
        const model = ModelStateActions.getModel(draftState, modelId);
        model[field] = value;
    }
}
