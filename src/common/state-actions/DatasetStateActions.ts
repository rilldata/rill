import {StateActions} from "../state-actions/StateActions";
import type {Dataset} from "$lib/types";
import type {DataModelerState} from "$lib/types";

export class DatasetStateActions extends StateActions {
    public addOrUpdateDatasetToState(draftState: DataModelerState, dataset: Dataset, isNew: boolean): void {
        if (isNew) {
            const datasetToUpdate = DatasetStateActions.getDataset(draftState, dataset.id);
            DatasetStateActions.shallowCopy(dataset, datasetToUpdate);
        } else {
            draftState.sources.push(dataset);
        }
    }
}
