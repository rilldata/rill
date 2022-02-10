import {StateActions} from ".//StateActions";
import type {DataModelerState, Model, ProfileColumn} from "$lib/types";

export class ModelStateActions extends StateActions {
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

    private static updateModelField<Field extends keyof Model>(draftState: DataModelerState, modelId: string,
                                                               field: Field, value: Model[Field]): void {
        const model = ModelStateActions.getModel(draftState, modelId);
        model[field] = value;
    }
}
