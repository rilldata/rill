import type { DataModelerService } from "$web-local/common/data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "$web-local/common/data-modeler-state-service/DataModelerStateService";
import { CSVFileTestData, TestDataColumns } from "../data/DataLoader.data";
import { DATA_FOLDER } from "../data/generator/data-constants";
import {
  EntityRecord,
  EntityStateService,
  EntityStatus,
  EntityType,
  StateType,
} from "$web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { asyncWait, waitUntil } from "$web-local/common/utils/waitUtils";
import type { PersistentTableEntity } from "$web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import type { DerivedTableEntity } from "$web-local/common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import type { PersistentModelEntity } from "$web-local/common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type { DerivedModelEntity } from "$web-local/common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import type { ProfileColumn } from "$web-local/lib/types";
import type { ActiveEntity } from "$web-local/common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type { MetricsDefinitionEntity } from "$web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { DatabaseService } from "$web-local/common/database-service/DatabaseService";

export class TestServer {
  constructor(
    public readonly dataModelerService: DataModelerService,
    public readonly dataModelerStateService: DataModelerStateService,
    public readonly databaseService: DatabaseService
  ) {}

  public async loadTestTables(): Promise<void> {
    await Promise.all(
      CSVFileTestData.subData.map(async (parquetFileData) => {
        await this.dataModelerService.dispatch("addOrUpdateTableFromFile", [
          `${DATA_FOLDER}/${parquetFileData.title}`,
        ]);
      })
    );
    await this.waitForTables();
  }

  public async waitForTables(): Promise<void> {
    await this.waitForEntity(EntityType.Table);
    await asyncWait(100);
  }
  public async waitForModels(): Promise<void> {
    await this.waitForEntity(EntityType.Model);
    await asyncWait(100);
  }

  public getTables<Field extends keyof PersistentTableEntity>(
    field: Field,
    value: PersistentTableEntity[Field]
  ): [PersistentTableEntity, DerivedTableEntity] {
    return this.getStatesForEntityType(EntityType.Table, field, value) as [
      PersistentTableEntity,
      DerivedTableEntity
    ];
  }
  public getModels<Field extends keyof PersistentModelEntity>(
    field: Field,
    value: PersistentModelEntity[Field]
  ): [PersistentModelEntity, DerivedModelEntity] {
    return this.getStatesForEntityType(EntityType.Model, field, value) as [
      PersistentModelEntity,
      DerivedModelEntity
    ];
  }
  public getMetricsDefinition<Field extends keyof MetricsDefinitionEntity>(
    field: Field,
    value: MetricsDefinitionEntity[Field]
  ): MetricsDefinitionEntity {
    return this.dataModelerStateService
      .getMetricsDefinitionService()
      .getByField(field, value);
  }

  public assertColumns(
    profileColumns: ProfileColumn[],
    columns: TestDataColumns
  ): void {
    profileColumns.forEach((profileColumn, idx) => {
      expect(profileColumn.name).toBe(columns[idx].name);
      expect(profileColumn.type).toBe(columns[idx].type);
      expect(profileColumn.nullCount > 0).toBe(columns[idx].isNull);
      // TODO: assert summary
      // console.log(profileColumn.name, profileColumn.summary);
    });
  }

  public getActiveEntity(): ActiveEntity {
    return this.dataModelerStateService.getApplicationState().activeEntity;
  }

  private async waitForEntity(entityType: EntityType): Promise<void> {
    await asyncWait(200);
    await waitUntil(() => {
      const currentState = this.dataModelerStateService
        .getEntityStateService(entityType, StateType.Derived)
        .getCurrentState();
      return (currentState.entities as any[]).every(
        (item) => item.status === EntityStatus.Idle
      );
    });
  }

  private getEntityByField(
    entityType: EntityType,
    stateType: StateType,
    field: string,
    value: unknown
  ): EntityRecord {
    return (
      this.dataModelerStateService.getEntityStateService(
        entityType,
        stateType
      ) as EntityStateService<any>
    ).getByField(field, value);
  }

  private getStatesForEntityType(
    entityType: EntityType,
    field: string,
    value: unknown
  ): [EntityRecord, EntityRecord] {
    const persistent = this.getEntityByField(
      entityType,
      StateType.Persistent,
      field,
      value
    );
    return [
      persistent,
      persistent
        ? this.getEntityByField(
            entityType,
            StateType.Derived,
            "id",
            persistent.id
          )
        : undefined,
    ];
  }
}
