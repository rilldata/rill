import type {DataModelerState, Dataset, Model, ProfileColumn} from "$lib/types";
import type {ColumnarItemType} from "$common/state-actions/ProfileColumnStateActions";
import {ColumnarItemTypeMap} from "$common/state-actions/ProfileColumnStateActions";
import {DataModelerActions} from "$common/data-modeler-actions/DataModelerActions";

export class ProfileColumnActions extends DataModelerActions {
    public async collectProfileColumns(currentState: DataModelerState,
                                       columnarItemId: string, columnarItemType: ColumnarItemType): Promise<void> {
        const item: Model | Dataset = (currentState[ColumnarItemTypeMap[columnarItemType]] as any[])
            .find(findItem => findItem.id === columnarItemId);
        await Promise.all(item.profile.map(column =>
            this.collectColumnInfo(columnarItemId, columnarItemType, item, column)));
    }

    private async collectColumnInfo(columnarItemId: string, columnarItemType: ColumnarItemType,
                                    item: Model | Dataset, column: ProfileColumn): Promise<void> {
        const promises = [];
        if (column.type.includes("VARCHAR")) {
            promises.push(this.collectTopKAndCardinality(columnarItemId, columnarItemType, item, column));
        } else {
            promises.push(this.collectNumericHistogram(columnarItemId, columnarItemType, item, column));
            if (column.type.includes("TIMESTAMP")) {
                promises.push(this.collectTimeRange(columnarItemId, columnarItemType, item, column));
            } else {
                promises.push(this.collectDescriptiveStatistics(columnarItemId, columnarItemType, item, column));
            }
        }
        promises.push(this.collectNullCount(columnarItemId, columnarItemType, item, column));
        await Promise.all(promises);
    }

    private async collectTopKAndCardinality(columnarItemId: string, columnarItemType: ColumnarItemType,
                                            item: Model | Dataset, column: ProfileColumn): Promise<void> {
        this.dataModelerStateManager.dispatch("updateColumnSummary", [
            columnarItemId, columnarItemType, column.name,
            await this.databaseColumnActions.getTopKAndCardinality(item.tableName, column.name),
        ]);
    }

    private async collectNumericHistogram(columnarItemId: string, columnarItemType: ColumnarItemType,
                                          item: Model | Dataset, column: ProfileColumn): Promise<void> {
        this.dataModelerStateManager.dispatch("updateColumnSummary", [
            columnarItemId, columnarItemType, column.name,
            await this.databaseColumnActions.getNumericHistogram(item.tableName, column.name, column.type),
        ]);
    }

    private async collectTimeRange(columnarItemId: string, columnarItemType: ColumnarItemType,
                                   item: Model | Dataset, column: ProfileColumn): Promise<void> {
        this.dataModelerStateManager.dispatch("updateColumnSummary", [
            columnarItemId, columnarItemType, column.name,
            await this.databaseColumnActions.getTimeRange(item.tableName, column.name),
        ]);
    }

    private async collectDescriptiveStatistics(columnarItemId: string, columnarItemType: ColumnarItemType,
                                               item: Model | Dataset, column: ProfileColumn): Promise<void> {
        this.dataModelerStateManager.dispatch("updateColumnSummary", [
            columnarItemId, columnarItemType, column.name,
            await this.databaseColumnActions.getDescriptiveStatistics(item.tableName, column.name),
        ]);
    }

    private async collectNullCount(columnarItemId: string, columnarItemType: ColumnarItemType,
                                   item: Model | Dataset, column: ProfileColumn): Promise<void> {
        this.dataModelerStateManager.dispatch("updateNullCount", [
            columnarItemId, columnarItemType, column.name,
            await this.databaseColumnActions.getNullCount(item.tableName, column.name),
        ]);
    }
}
