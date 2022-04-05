import { EntityType, StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
    DataProfileStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/DataProfileEntity";
import { DataModelerActions } from "$common/data-modeler-service/DataModelerActions";
import { BOOLEANS, CATEGORICALS, TIMESTAMPS } from "$lib/duckdb-data-types";
import type { ProfileColumn } from "$lib/types";
import { DatabaseActionQueuePriority } from "$common/priority-action-queue/DatabaseActionQueuePriority";

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
        try {
            await Promise.all(entity.profile.map(column =>
                this.collectColumnInfo(entityType, entityId, persistentEntity.tableName, column)));
        } catch (err) {}
    }

    private async collectColumnInfo(entityType: EntityType, entityId: string,
                                    tableName: string, column: ProfileColumn): Promise<void> {
        const promises = [];
        if (CATEGORICALS.has(column.type) || BOOLEANS.has(column.type)) {
            promises.push(this.collectTopKAndCardinality(entityType, entityId, tableName, column));
        } else {
            promises.push(this.collectNumericHistogram(entityType, entityId, tableName, column));
            if (TIMESTAMPS.has(column.type)) {
                promises.push(this.collectTimeRange(entityType, entityId, tableName, column));
            } else {
                promises.push(this.collectDescriptiveStatistics(entityType, entityId, tableName, column));
            }
        }
        promises.push(this.collectNullCount(entityType, entityId, tableName, column));
        await Promise.all(promises);
    }

    private async collectTopKAndCardinality(entityType: EntityType, entityId: string,
                                            tableName: string, column: ProfileColumn): Promise<void> {
        this.dataModelerStateService.dispatch("updateColumnSummary",[
            entityType, entityId, column.name,
            await this.databaseActionQueue.enqueue(
                {id: entityId, priority: ColumnProfilePriorityMap[entityType]},
                "getTopKAndCardinality", [tableName, column.name]),
        ]);
    }

    private async estimateTimeGrain(entityType: EntityType, entityId: string,
                                            tableName: string, column: ProfileColumn): Promise<void> {
            this.dataModelerStateService.dispatch("updateColumnSummary",[
            entityType, entityId, column.name,
            await this.databaseActionQueue.enqueue(
            {id: entityId, priority: ColumnProfilePriorityMap[entityType]},
            "estimateTimeGrain", [tableName, column.name]),
        ]);
    }

    private async collectNumericHistogram(entityType: EntityType, entityId: string,
                                          tableName: string, column: ProfileColumn): Promise<void> {
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
                                               tableName: string, column: ProfileColumn): Promise<void> {
        this.dataModelerStateService.dispatch("updateColumnSummary", [
            entityType, entityId, column.name,
            await this.databaseActionQueue.enqueue(
                {id: entityId, priority: ColumnProfilePriorityMap[entityType]},
                "getDescriptiveStatistics", [tableName, column.name]),
        ]);
    }

    private async collectNullCount(entityType: EntityType, entityId: string,
                                   tableName: string, column: ProfileColumn): Promise<void> {
        this.dataModelerStateService.dispatch("updateNullCount", [
            entityType, entityId, column.name,
            await this.databaseActionQueue.enqueue(
                {id: entityId, priority: ColumnProfilePriorityMap[entityType]},
                "getNullCount", [tableName, column.name]),
        ]);
    }
}
