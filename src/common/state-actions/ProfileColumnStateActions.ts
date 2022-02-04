import {StateActions} from "../state-actions/StateActions";
import type {ColumnarItem, ColumnarTypeKeys, DataModelerState, ProfileColumn, ProfileColumnSummary} from "$lib/types";

export enum ColumnarItemType {
    Dataset,
    Model,
}
export const ColumnarItemTypeMap: {
    [type in ColumnarItemType]: ColumnarTypeKeys
} = {
    [ColumnarItemType.Dataset]: "sources",
    [ColumnarItemType.Model]: "queries",
}

export class ProfileColumnStateActions extends StateActions {
    public clearProfileSummary(draftState: DataModelerState,
                               columnarItemId: string, columnarItemType: ColumnarItemType): void {
        const modelToUpdate = ProfileColumnStateActions.getByID<ColumnarItem>(
            draftState[ColumnarItemTypeMap[columnarItemType]], columnarItemId);
        modelToUpdate.profile?.forEach((profile) => {
            profile.summary = null;
            profile.nullCount = null;
        });
    }

    public updateColumnSummary(draftState: DataModelerState,
                               columnarItemId: string, columnarItemType: ColumnarItemType,
                               columnName: string, summary: ProfileColumnSummary): void {
        const profileToUpdate = ProfileColumnStateActions.getProfile(
            draftState[ColumnarItemTypeMap[columnarItemType]],
            columnarItemId, columnName);
        profileToUpdate.summary ??= {};
        ProfileColumnStateActions.shallowCopy(summary, profileToUpdate.summary)
    }

    public updateNullCount(draftState: DataModelerState,
                           columnarItemId: string, columnarItemType: ColumnarItemType,
                           columnName: string, nullCount: number): void {
        const profileToUpdate = ProfileColumnStateActions.getProfile(
            draftState[ColumnarItemTypeMap[columnarItemType]],
            columnarItemId, columnName);
        profileToUpdate.nullCount = nullCount;
    }

    public updateProfiles(draftState: DataModelerState,
                          columnarItemId: string, columnarItemType: ColumnarItemType,
                          profiles: ProfileColumn[]): void {
        const modelToUpdate = ProfileColumnStateActions.getByID<ColumnarItem>(
            draftState[ColumnarItemTypeMap[columnarItemType]], columnarItemId);
        modelToUpdate.profile = profiles;
    }
}
