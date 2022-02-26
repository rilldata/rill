import type {DataModelerState, Table, Model, ProfileColumn} from "$lib/types";
import type {ColumnarItemType} from "$common/data-modeler-state-service/ProfileColumnStateActions";
import {ColumnarItemTypeMap} from "$common/data-modeler-state-service/ProfileColumnStateActions";
import {DataModelerActions} from "$common/data-modeler-service/DataModelerActions";
import {TIMESTAMPS, CATEGORICALS} from "$lib/duckdb-data-types";

export class ProfileColumnActions extends DataModelerActions {
    public async collectProfileColumns(currentState: DataModelerState,
                                       columnarItemId: string, columnarItemType: ColumnarItemType): Promise<void> {
        const item: Model | Table = (currentState[ColumnarItemTypeMap[columnarItemType]] as any[])
            .find(findItem => findItem.id === columnarItemId);
        await Promise.all(item.profile.map(column =>
            this.collectColumnInfo(columnarItemId, columnarItemType, item, column)));
    }

    private async collectColumnInfo(columnarItemId: string, columnarItemType: ColumnarItemType,
                                    item: Model | Table, column: ProfileColumn): Promise<void> {
        const promises = [];
        if (CATEGORICALS.has(column.type)) {
            promises.push(this.collectTopKAndCardinality(columnarItemId, columnarItemType, item, column));
        } else {
            promises.push(this.collectNumericHistogram(columnarItemId, columnarItemType, item, column));
            if (TIMESTAMPS.has(column.type)) {
                promises.push(this.collectTimeRange(columnarItemId, columnarItemType, item, column));
            } else {
                promises.push(this.collectDescriptiveStatistics(columnarItemId, columnarItemType, item, column));
            }
        }
        promises.push(this.collectNullCount(columnarItemId, columnarItemType, item, column));
        await Promise.all(promises);
    }

    private async collectTopKAndCardinality(columnarItemId: string, columnarItemType: ColumnarItemType,
                                            item: Model | Table, column: ProfileColumn): Promise<void> {
        this.dataModelerStateService.dispatch("updateColumnSummary", [
            columnarItemId, columnarItemType, column.name,
            await this.databaseService.dispatch("getTopKAndCardinality", [item.tableName, column.name]),
        ]);
    }

    private async collectNumericHistogram(columnarItemId: string, columnarItemType: ColumnarItemType,
                                          item: Model | Table, column: ProfileColumn): Promise<void> {
        this.dataModelerStateService.dispatch("updateColumnSummary", [
            columnarItemId, columnarItemType, column.name,
            await this.databaseService.dispatch("getNumericHistogram", [item.tableName, column.name, column.type]),
        ]);
    }

    private async collectTimeRange(columnarItemId: string, columnarItemType: ColumnarItemType,
                                   item: Model | Table, column: ProfileColumn): Promise<void> {
        this.dataModelerStateService.dispatch("updateColumnSummary", [
            columnarItemId, columnarItemType, column.name,
            await this.databaseService.dispatch("getTimeRange", [item.tableName, column.name]),
        ]);
    }

    private async collectDescriptiveStatistics(columnarItemId: string, columnarItemType: ColumnarItemType,
                                               item: Model | Table, column: ProfileColumn): Promise<void> {
        this.dataModelerStateService.dispatch("updateColumnSummary", [
            columnarItemId, columnarItemType, column.name,
            await this.databaseService.dispatch("getDescriptiveStatistics", [item.tableName, column.name]),
        ]);
    }

    private async collectNullCount(columnarItemId: string, columnarItemType: ColumnarItemType,
                                   item: Model | Table, column: ProfileColumn): Promise<void> {
        this.dataModelerStateService.dispatch("updateNullCount", [
            columnarItemId, columnarItemType, column.name,
            await this.databaseService.dispatch("getNullCount", [item.tableName, column.name]),
        ]);
    }
}
