import {DatasetActions} from "$common/data-modeler-actions/DatasetActions";
import type {DataModelerState, Dataset, Model, ProfileColumn} from "$lib/types";
import type {ColumnarItemType} from "$common/state-actions/ProfileColumnStateActions";
import {ColumnarItemTypeMap} from "$common/state-actions/ProfileColumnStateActions";

export class ProfileColumnActions extends DatasetActions {
    public async collectProfileColumns(currentState: DataModelerState,
                                       columnarItemId: string, columnarItemType: ColumnarItemType): Promise<void> {
        const item: Model | Dataset = (currentState[ColumnarItemTypeMap[columnarItemType]] as any[])
            .find(findItem => findItem.id === columnarItemId);
        await Promise.all(item.profile.map(column =>
            this.collectColumnInfo(columnarItemId, columnarItemType, item, column)));
    }

    private async collectColumnInfo(columnarItemId: string, columnarItemType: ColumnarItemType,
                                    item: Model | Dataset, column: ProfileColumn): Promise<void> {
        if (column.type.includes("VARCHAR")) {
            this.dataModelerStateManager.dispatch("updateColumnSummary", [
                columnarItemId, columnarItemType, column.name,
                await this.duckDBColumnAPI.getTopKAndCardinality(item.name, column.name),
            ]);
        } else {
            this.dataModelerStateManager.dispatch("updateColumnSummary", [
                columnarItemId, columnarItemType, column.name,
                await this.duckDBColumnAPI.getNumericHistogram(item.name, column.name, column.type),
            ]);
            if (column.type.includes("TIMESTAMP")) {
                this.dataModelerStateManager.dispatch("updateColumnSummary", [
                    columnarItemId, columnarItemType, column.name,
                    await this.duckDBColumnAPI.getTimeRange(item.name, column.name),
                ]);
            } else {
                this.dataModelerStateManager.dispatch("updateColumnSummary", [
                    columnarItemId, columnarItemType, column.name,
                    await this.duckDBColumnAPI.getDescriptiveStatistics(item.name, column.name),
                ]);
            }
        }
        this.dataModelerStateManager.dispatch("updateNullCount", [
            columnarItemId, columnarItemType, column.name,
            await this.duckDBColumnAPI.getNullCount(item.name, column.name),
        ]);
    }
}
