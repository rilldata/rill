import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { DataProfileStateActionArg } from "$common/data-modeler-state-service/entity-state-service/DataProfileEntity";
import { DataModelerActions } from "$common/data-modeler-service/DataModelerActions";
import { CATEGORICALS, NUMERICS, TIMESTAMPS } from "$lib/duckdb-data-types";
import type { ProfileColumn } from "$lib/types";
import {
  DatabaseActionQueuePriority,
  DatabaseProfilesFieldPriority,
  MetadataPriority,
  getProfilePriority,
  ProfileMetadataPriorityMap,
} from "$common/priority-action-queue/DatabaseActionQueuePriority";
import { COLUMN_PROFILE_CONFIG } from "$lib/application-config";
import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type { PersistentTableEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";

const ProfileEntityPriorityMap = {
  [EntityType.Table]: DatabaseActionQueuePriority.TableProfile,
  [EntityType.Model]: DatabaseActionQueuePriority.ActiveModelProfile,
};

export class ProfileColumnActions extends DataModelerActions {
  @DataModelerActions.DerivedAction()
  public async collectProfileColumns(
    { stateService }: DataProfileStateActionArg,
    entityType: EntityType,
    entityId: string
  ): Promise<void> {
    const persistentEntity = this.dataModelerStateService.getEntityById(
      entityType,
      StateType.Persistent,
      entityId
    );
    const entity = stateService.getById(entityId);
    if (!entity) {
      console.error(
        `Entity not found. entityType=${entityType} entityId=${entityId}`
      );
      return;
    }
    try {
      await Promise.all(
        entity.profile.map((column) =>
          this.collectColumnInfo(
            entityType,
            entityId,
            (persistentEntity as PersistentModelEntity | PersistentTableEntity)
              .tableName,
            column
          )
        )
      );
    } catch (err) {
      // continue regardless of error
    }
  }

  private async collectColumnInfo(
    entityType: EntityType,
    entityId: string,
    tableName: string,
    column: ProfileColumn
  ): Promise<void> {
    const promises = [];
    if (CATEGORICALS.has(column.type)) {
      promises.push(
        this.collectCardinality(entityType, entityId, tableName, column)
      );
      promises.push(this.collectTopK(entityType, entityId, tableName, column));
    } else {
      if (NUMERICS.has(column.type)) {
        promises.push(
          this.collectNumericHistogram(entityType, entityId, tableName, column)
        );
        promises.push(
          this.collectRugHistogram(entityType, entityId, tableName, column)
        );
      }
      if (TIMESTAMPS.has(column.type)) {
        promises.push(
          this.collectTimeRange(entityType, entityId, tableName, column)
        );
        promises.push(
          this.collectSmallestTimegrainEstimate(
            entityType,
            entityId,
            tableName,
            column
          )
        );
        promises.push(
          this.collectTimestampRollup(
            entityType,
            entityId,
            tableName,
            column,
            // use the medium width for the spark line
            COLUMN_PROFILE_CONFIG.summaryVizWidth.medium,
            undefined
          )
        );
      } else {
        promises.push(
          this.collectDescriptiveStatistics(
            entityType,
            entityId,
            tableName,
            column
          )
        );
      }
    }
    promises.push(
      this.collectNullCount(entityType, entityId, tableName, column)
    );
    await Promise.all(promises);
  }

  private async collectTopK(
    entityType: EntityType,
    entityId: string,
    tableName: string,
    column: ProfileColumn
  ): Promise<void> {
    this.dataModelerStateService.dispatch("updateColumnSummary", [
      entityType,
      entityId,
      column.name,
      await this.databaseActionQueue.enqueue(
        {
          id: entityId + column.name + MetadataPriority.Essential,
          priority: getProfilePriority(
            ProfileEntityPriorityMap[entityType],
            DatabaseProfilesFieldPriority.NonFocused,
            ProfileMetadataPriorityMap[MetadataPriority.Essential]
          ),
        },
        "getTopK",
        [tableName, column.name]
      ),
    ]);
  }

  private async collectCardinality(
    entityType: EntityType,
    entityId: string,
    tableName: string,
    column: ProfileColumn
  ): Promise<void> {
    this.dataModelerStateService.dispatch("updateColumnSummary", [
      entityType,
      entityId,
      column.name,
      await this.databaseActionQueue.enqueue(
        {
          id: entityId + column.name + MetadataPriority.Summary,
          priority: getProfilePriority(
            ProfileEntityPriorityMap[entityType],
            DatabaseProfilesFieldPriority.NonFocused,
            ProfileMetadataPriorityMap[MetadataPriority.Summary]
          ),
        },
        "getCardinalityOfColumn",
        [tableName, column.name]
      ),
    ]);
  }

  private async collectSmallestTimegrainEstimate(
    entityType: EntityType,
    entityId: string,
    tableName: string,
    column: ProfileColumn
  ): Promise<void> {
    this.dataModelerStateService.dispatch("updateColumnSummary", [
      entityType,
      entityId,
      column.name,
      await this.databaseActionQueue.enqueue(
        {
          id: entityId + column.name + MetadataPriority.Essential,
          priority: getProfilePriority(
            ProfileEntityPriorityMap[entityType],
            DatabaseProfilesFieldPriority.NonFocused,
            ProfileMetadataPriorityMap[MetadataPriority.Essential]
          ),
        },
        "estimateSmallestTimeGrain",
        [tableName, column.name]
      ),
    ]);
  }

  private async collectTimestampRollup(
    entityType: EntityType,
    entityId: string,
    tableName: string,
    column: ProfileColumn,
    pixels: number = undefined,
    sampleSize: number = undefined
  ): Promise<void> {
    this.dataModelerStateService.dispatch("updateColumnSummary", [
      entityType,
      entityId,
      column.name,
      await this.databaseActionQueue.enqueue(
        {
          id: entityId + column.name + MetadataPriority.Summary,
          priority: getProfilePriority(
            ProfileEntityPriorityMap[entityType],
            DatabaseProfilesFieldPriority.NonFocused,
            ProfileMetadataPriorityMap[MetadataPriority.Summary]
          ),
        },
        "generateTimeSeries",
        [
          {
            tableName,
            timestampColumn: column.name,
            pixels,
            sampleSize,
          },
        ]
      ),
    ]);
  }

  private async collectNumericHistogram(
    entityType: EntityType,
    entityId: string,
    tableName: string,
    column: ProfileColumn
  ): Promise<void> {
    this.dataModelerStateService.dispatch("updateColumnSummary", [
      entityType,
      entityId,
      column.name,
      await this.databaseActionQueue.enqueue(
        {
          id: entityId + column.name + MetadataPriority.Summary,
          priority: getProfilePriority(
            ProfileEntityPriorityMap[entityType],
            DatabaseProfilesFieldPriority.NonFocused,
            ProfileMetadataPriorityMap[MetadataPriority.Summary]
          ),
        },
        "getNumericHistogram",
        [tableName, column.name, column.type]
      ),
    ]);
  }

  private async collectRugHistogram(
    entityType: EntityType,
    entityId: string,
    tableName: string,
    column: ProfileColumn
  ): Promise<void> {
    this.dataModelerStateService.dispatch("updateColumnSummary", [
      entityType,
      entityId,
      column.name,
      await this.databaseActionQueue.enqueue(
        {
          id: entityId + column.name + MetadataPriority.Deeper,
          priority: getProfilePriority(
            ProfileEntityPriorityMap[entityType],
            DatabaseProfilesFieldPriority.NonFocused,
            ProfileMetadataPriorityMap[MetadataPriority.Deeper]
          ),
        },
        "getRugHistogram",
        [tableName, column.name, column.type]
      ),
    ]);
  }

  private async collectTimeRange(
    entityType: EntityType,
    entityId: string,
    tableName: string,
    column: ProfileColumn
  ): Promise<void> {
    this.dataModelerStateService.dispatch("updateColumnSummary", [
      entityType,
      entityId,
      column.name,
      await this.databaseActionQueue.enqueue(
        {
          id: entityId + column.name + MetadataPriority.Essential,
          priority: getProfilePriority(
            ProfileEntityPriorityMap[entityType],
            DatabaseProfilesFieldPriority.NonFocused,
            ProfileMetadataPriorityMap[MetadataPriority.Essential]
          ),
        },
        "getTimeRange",
        [tableName, column.name]
      ),
    ]);
  }

  private async collectDescriptiveStatistics(
    entityType: EntityType,
    entityId: string,
    tableName: string,
    column: ProfileColumn
  ): Promise<void> {
    this.dataModelerStateService.dispatch("updateColumnSummary", [
      entityType,
      entityId,
      column.name,
      await this.databaseActionQueue.enqueue(
        {
          id: entityId + column.name + MetadataPriority.Essential,
          priority: getProfilePriority(
            ProfileEntityPriorityMap[entityType],
            DatabaseProfilesFieldPriority.NonFocused,
            ProfileMetadataPriorityMap[MetadataPriority.Essential]
          ),
        },
        "getDescriptiveStatistics",
        [tableName, column.name]
      ),
    ]);
  }

  private async collectNullCount(
    entityType: EntityType,
    entityId: string,
    tableName: string,
    column: ProfileColumn
  ): Promise<void> {
    this.dataModelerStateService.dispatch("updateNullCount", [
      entityType,
      entityId,
      column.name,
      await this.databaseActionQueue.enqueue(
        {
          id: entityId + column.name + MetadataPriority.Summary,
          priority: getProfilePriority(
            ProfileEntityPriorityMap[entityType],
            DatabaseProfilesFieldPriority.NonFocused,
            ProfileMetadataPriorityMap[MetadataPriority.Summary]
          ),
        },
        "getNullCount",
        [tableName, column.name]
      ),
    ]);
  }
}
