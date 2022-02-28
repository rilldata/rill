import { EntityType, StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
    DataProfileStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/DataProfileEntity";
import { DataModelerActions } from "$common/data-modeler-service/DataModelerActions";
import { CATEGORICALS, TIMESTAMPS } from "$lib/duckdb-data-types";
import type { ProfileColumn } from "$lib/types";

export class ProfileColumnActions extends DataModelerActions {
    @DataModelerActions.DerivedAction()
    public async collectProfileColumns({stateService}: DataProfileStateActionArg,
                                       entityType: EntityType, entityId: string): Promise<void> {
        const persistentEntity = this.dataModelerStateService
            .getEntityById(entityType, StateType.Persistent, entityId);
        const entity = stateService.getById(entityId);
        await Promise.all(entity.profile.map(column =>
            this.collectColumnInfo(entityType, entityId, persistentEntity.tableName, column)));
    }

    private async collectColumnInfo(entityType: EntityType, entityId: string,
                                    tableName: string, column: ProfileColumn): Promise<void> {
        const promises = [];
        if (CATEGORICALS.has(column.type)) {
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
            await this.databaseService.dispatch("getTopKAndCardinality", [tableName, column.name]),
        ]);
    }

    private async collectNumericHistogram(entityType: EntityType, entityId: string,
                                          tableName: string, column: ProfileColumn): Promise<void> {
        this.dataModelerStateService.dispatch("updateColumnSummary", [
            entityType, entityId, column.name,
            await this.databaseService.dispatch("getNumericHistogram", [tableName, column.name, column.type]),
        ]);
    }

    private async collectTimeRange(entityType: EntityType, entityId: string,
                                   tableName: string, column: ProfileColumn): Promise<void> {
        this.dataModelerStateService.dispatch("updateColumnSummary", [
            entityType, entityId, column.name,
            await this.databaseService.dispatch("getTimeRange", [tableName, column.name]),
        ]);
    }

    private async collectDescriptiveStatistics(entityType: EntityType, entityId: string,
                                               tableName: string, column: ProfileColumn): Promise<void> {
        this.dataModelerStateService.dispatch("updateColumnSummary", [
            entityType, entityId, column.name,
            await this.databaseService.dispatch("getDescriptiveStatistics", [tableName, column.name]),
        ]);
    }

    private async collectNullCount(entityType: EntityType, entityId: string,
                                   tableName: string, column: ProfileColumn): Promise<void> {
        this.dataModelerStateService.dispatch("updateNullCount", [
            entityType, entityId, column.name,
            await this.databaseService.dispatch("getNullCount", [tableName, column.name]),
        ]);
    }
}
