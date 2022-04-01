import { EntityType, StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
    DataProfileStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/DataProfileEntity";
import { DataModelerActions } from "$common/data-modeler-service/DataModelerActions";
import { BOOLEANS, CATEGORICALS, TIMESTAMPS } from "$lib/duckdb-data-types";
import type { ProfileColumn } from "$lib/types";
import { DatabaseActionQueuePriority } from "$common/priority-action-queue/DatabaseActionQueuePriority";
import type {DuckDBColumnSummary, DuckDBTableSummary} from "$common/duckdbTypes";
import {parseNumber} from "$common/utils/parseNumber";

const ColumnProfilePriorityMap = {
    [EntityType.Table]: DatabaseActionQueuePriority.TableProfile,
    [EntityType.Model]: DatabaseActionQueuePriority.ActiveModelProfile,
}

export class ProfileColumnActions extends DataModelerActions {
    @DataModelerActions.DerivedAction()
    public async collectProfileColumns({stateService}: DataProfileStateActionArg,
                                       entityType: EntityType, entityId: string): Promise<void> {
        const persistentEntity = this.dataModelerStateService
            .getEntityById(entityType, StateType.Persistent, entityId);
        const entity = stateService.getById(entityId);
        if (!entity) {
            console.error(`Entity not found. entityType=${entityType} entityId=${entityId}`);
            return;
        }
        const tableSummary: DuckDBTableSummary = await this.databaseActionQueue.enqueue(
            {id: entityId, priority: ColumnProfilePriorityMap[entityType]},
            "getTableSummary", [persistentEntity.tableName]);
        try {
            await Promise.all(entity.profile.map(column =>
                this.collectColumnInfo(entityType, entityId, persistentEntity.tableName, column,
                    tableSummary.find(summary => summary.column_name === column.name))));
        } catch (err) {}
    }

    private async collectColumnInfo(entityType: EntityType, entityId: string,
                                    tableName: string, column: ProfileColumn,
                                    columnSummary: DuckDBColumnSummary): Promise<void> {
        const promises = [];
        if (CATEGORICALS.has(column.type) || BOOLEANS.has(column.type)) {
            promises.push(this.collectTopKAndCardinality(entityType, entityId, tableName, column, columnSummary));
        } else {
            promises.push(this.collectNumericHistogram(entityType, entityId, tableName, column, columnSummary));
            if (TIMESTAMPS.has(column.type)) {
                promises.push(this.collectTimeRange(entityType, entityId, tableName, column));
            } else {
                promises.push(this.collectDescriptiveStatistics(entityType, entityId, tableName, column, columnSummary));
            }
        }
        promises.push(this.collectNullCount(entityType, entityId, tableName, column, columnSummary));
        await Promise.all(promises);
    }

    private async collectTopKAndCardinality(entityType: EntityType, entityId: string,
                                            tableName: string, column: ProfileColumn,
                                            columnSummary: DuckDBColumnSummary): Promise<void> {
        this.dataModelerStateService.dispatch("updateColumnSummary",[
            entityType, entityId, column.name,
            {
                topK: await this.databaseActionQueue.enqueue(
                    {id: entityId, priority: ColumnProfilePriorityMap[entityType]},
                    "getTopKOfColumn", [tableName, column.name]),
                cardinality: parseNumber(columnSummary.approx_unique),
            }
        ]);
    }

    private async collectNumericHistogram(entityType: EntityType, entityId: string,
                                          tableName: string, column: ProfileColumn,
                                          columnSummary: DuckDBColumnSummary): Promise<void> {
        this.dataModelerStateService.dispatch("updateColumnSummary", [
            entityType, entityId, column.name,
            await this.databaseActionQueue.enqueue(
                {id: entityId, priority: ColumnProfilePriorityMap[entityType]},
                "getNumericHistogram", [tableName, column.name, column.type]),
        ]);
    }

    private async collectTimeRange(entityType: EntityType, entityId: string,
                                   tableName: string, column: ProfileColumn): Promise<void> {
        this.dataModelerStateService.dispatch("updateColumnSummary", [
            entityType, entityId, column.name,
            await this.databaseActionQueue.enqueue(
                {id: entityId, priority: ColumnProfilePriorityMap[entityType]},
                "getTimeRange", [tableName, column.name]),
        ]);
    }

    private async collectDescriptiveStatistics(entityType: EntityType, entityId: string,
                                               tableName: string, column: ProfileColumn,
                                               columnSummary: DuckDBColumnSummary): Promise<void> {
        this.dataModelerStateService.dispatch("updateColumnSummary", [
            entityType, entityId, column.name,
            {
                statistics: {
                    max: parseNumber(columnSummary.max),
                    min: parseNumber(columnSummary.min),
                    mean: parseNumber(columnSummary.avg),
                    q25: parseNumber(columnSummary.q25),
                    q50: parseNumber(columnSummary.q50),
                    q75: parseNumber(columnSummary.q75),
                    sd: parseNumber(columnSummary.std),
                }
            },
        ]);
    }

    private async collectNullCount(entityType: EntityType, entityId: string,
                                   tableName: string, column: ProfileColumn,
                                   columnSummary: DuckDBColumnSummary): Promise<void> {
        const nullPercent = columnSummary.null_percentage ?
            Number(columnSummary.null_percentage.replace("%", "")) : 0;
        this.dataModelerStateService.dispatch("updateNullCount", [
            entityType, entityId, column.name,
            Math.round(nullPercent * columnSummary.count / 100),
        ]);
    }
}
