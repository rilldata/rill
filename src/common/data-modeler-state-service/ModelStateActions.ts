import {StateActions} from "./StateActions";
import { extractSourceTables } from "$lib/util/model-structure";
import type {
    PersistentModelStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type { ProfileColumn, SourceTable } from "$lib/types";
import type {
    DerivedModelStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";

export interface NewModelParams {
    query?: string;
    name?: string;
    at?: number;
    makeActive?: boolean;
}

export class ModelStateActions extends StateActions {
    @StateActions.PersistentModelAction()
    public incrementModelNumber({draftState}: PersistentModelStateActionArg) {
        draftState.modelNumber++;
    }

    @StateActions.DerivedModelAction()
    public addModelError({stateService, draftState}: DerivedModelStateActionArg,
                         modelId: string, message: string): void {
        stateService.updateEntityField(draftState, modelId, "error", message);
    }

    @StateActions.DerivedModelAction()
    public clearModelError({stateService, draftState}: DerivedModelStateActionArg,
                           modelId: string): void {
        stateService.updateEntityField(draftState, modelId, "error", undefined);
    }

    @StateActions.DerivedModelAction()
    public clearModelProfile({stateService, draftState}: DerivedModelStateActionArg,
                             modelId: string): void {
        const model = stateService.getById(modelId, draftState);
        model.sizeInBytes = undefined;
        model.preview = undefined;
        model.profile = undefined;
        // TODO: update this automatically
        model.lastUpdated = Date.now();
    }

    @StateActions.PersistentModelAction()
    public updateModelQuery({stateService, draftState}: PersistentModelStateActionArg,
                            modelId: string, query: string, sanitizedQuery: string): void {
        stateService.updateEntityField(draftState, modelId, "query", query);
    }

    @StateActions.DerivedModelAction()
    public updateModelSanitizedQuery({stateService, draftState}: DerivedModelStateActionArg,
                                     modelId: string, sanitizedQuery: string): void {
        stateService.updateEntityField(draftState, modelId, "sanitizedQuery", sanitizedQuery);
    }

    @StateActions.DerivedModelAction()
    public updateModelProfileColumns({stateService, draftState}: DerivedModelStateActionArg,
                                     modelId: string, profileColumns: Array<ProfileColumn>): void {
        stateService.updateEntityField(draftState, modelId, "profile", profileColumns);
    }

    @StateActions.DerivedModelAction()
    public getModelSourceTables({stateService, draftState}: DerivedModelStateActionArg,
                                modelId: string, query: string): void {
        const sourceTables = extractSourceTables(query) as SourceTable[];
        stateService.updateEntityField(draftState, modelId, "sources", sourceTables);
    }

    @StateActions.DerivedModelAction()
    public clearSourceTables({stateService, draftState}: DerivedModelStateActionArg,
        modelId: string): void {
        stateService.updateEntityField(draftState, modelId, "sources", []);
    }

    @StateActions.DerivedModelAction()
    public updateModelPreview({stateService, draftState}: DerivedModelStateActionArg,
                              modelId: string, preview: Array<any>): void {
        stateService.updateEntityField(draftState, modelId, "preview", preview);
    }

    @StateActions.DerivedModelAction()
    public updateModelCardinality({stateService, draftState}: DerivedModelStateActionArg,
                                  modelId: string, cardinality: number): void {
        stateService.updateEntityField(draftState, modelId, "cardinality", cardinality);
    }

    @StateActions.DerivedModelAction()
    public updateModelDestinationSize({stateService, draftState}: DerivedModelStateActionArg,
                                      modelId: string, sizeInBytes: number): void {
        stateService.updateEntityField(draftState, modelId, "sizeInBytes", sizeInBytes);
    }

    @StateActions.PersistentModelAction()
    public updateModelName({stateService, draftState}: PersistentModelStateActionArg,
                           modelId: string, name: string): void {
        stateService.updateEntityField(draftState, modelId, "name", `${name}.sql`);
        stateService.updateEntityField(draftState, modelId, "tableName", name);
    }
}
