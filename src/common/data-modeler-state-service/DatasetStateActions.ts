import {StateActions} from ".//StateActions";
import type {Dataset} from "$lib/types";
import type {DataModelerState} from "$lib/types";
import type {Model} from "$lib/types";
import {ColumnarItemType, ColumnarItemTypeMap} from "$common/data-modeler-state-service/ProfileColumnStateActions";

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
    public setActiveAsset(draftState: DataModelerState, id: string, assetType: string): void {
        draftState.activeAsset = { id, assetType };
    }
    public unsetActiveAsset(draftState: DataModelerState): void {
        draftState.activeAsset = undefined;
    }

    public setDatasetStatus(draftState: DataModelerState, columnarItemType: ColumnarItemType, columnarItemId: string, status: string): void {
        const item: Model | Dataset = (draftState[ColumnarItemTypeMap[columnarItemType]] as any[])
            .find(findItem => findItem.id === columnarItemId);
        item.status = status;
    }

    public pruneAndDedupeDatasets(draftState: DataModelerState, files: Array<string>): void {
        const filePaths = new Set(files);

        draftState.sources = draftState.sources.filter((dataset, index, self) => {
           if (!filePaths.has(dataset.path)) return false;
           return index === self.findIndex(indexCheckDataset => (indexCheckDataset.path === dataset.path));
        });
    }
}
