import {StateActions} from "../state-actions/StateActions";
import type {Dataset} from "$lib/types";
import type {DataModelerState} from "$lib/types";
import type {Model} from "$lib/types";
import {ColumnarItemType, ColumnarItemTypeMap} from "$common/state-actions/ProfileColumnStateActions";

export class DatasetStateActions extends StateActions {
    public addOrUpdateDatasetToState(draftState: DataModelerState, dataset: Dataset, isNew: boolean): void {
        if (isNew) {
            draftState.sources.push(dataset);
        } else {
            const datasetToUpdate = DatasetStateActions.getDataset(draftState, dataset.id);
            DatasetStateActions.shallowCopy(dataset, datasetToUpdate);
        }
    }

    // TODO: find a better place for this
    public setStatus(draftState: DataModelerState, status: string): void {
        draftState.status = status;
    }

    public setDatasetStatus(draftState: DataModelerState, columnarItemType: ColumnarItemType, columnarItemId: string, status: string): void {
        const item: Model | Dataset = (draftState[ColumnarItemTypeMap[columnarItemType]] as any[])
            .find(findItem => findItem.id === columnarItemId);
        item.status = status;
    }
}
