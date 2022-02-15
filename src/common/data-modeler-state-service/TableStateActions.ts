import {StateActions} from ".//StateActions";
import type {Table} from "$lib/types";
import type {DataModelerState} from "$lib/types";
import type {Model} from "$lib/types";
import {ColumnarItemType, ColumnarItemTypeMap} from "$common/data-modeler-state-service/ProfileColumnStateActions";

export class TableStateActions extends StateActions {
    public addOrUpdateTableToState(draftState: DataModelerState, table: Table, isNew: boolean): void {
        if (isNew) {
            draftState.tables.push(table);
        } else {
            const tableToUpdate = TableStateActions.getTable(draftState, table.id);
            TableStateActions.shallowCopy(table, tableToUpdate);
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

    public setTableStatus(draftState: DataModelerState, columnarItemType: ColumnarItemType, columnarItemId: string, status: string): void {
        const item: Model | Table = (draftState[ColumnarItemTypeMap[columnarItemType]] as any[])
            .find(findItem => findItem.id === columnarItemId);
        item.status = status;
    }

    public pruneAndDedupeTables(draftState: DataModelerState, files: Array<string>): void {
        const filePaths = new Set(files);

        const newSources = draftState.tables.filter((table, index, self) => {
           if (!filePaths.has(table.path)) return false;
           return index === self.findIndex(indexCheckTable => (indexCheckTable.path === table.path));
        });
        if (newSources.length !== draftState.tables.length) {
            draftState.tables = newSources;
        }
    }
}
