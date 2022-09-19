import {
  EntityRecord,
  EntityStateActionArg,
  EntityStatus,
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { EntityStateActionArgMapType } from "$common/data-modeler-state-service/entity-state-service/EntityStateServicesMap";

export abstract class ActionsBase {
  public static actionToStateTypesMap: Record<string, [EntityType, StateType]>;

  /**
   * Method decorator that marks a method as an action.
   * Takes an {@link EntityType} and {@link StateType} to denote what state the actions is for.
   * @param entityType
   * @param stateType
   */
  public static Action<
    EntityTypeArg extends EntityType,
    StateTypeArg extends StateType
  >(entityType: EntityTypeArg, stateType: StateTypeArg) {
    return (
      target: ActionsBase,
      propertyKey: string,
      // make sure the decorator and the state action arg match using this
      _descriptor: TypedPropertyDescriptor<
        (
          stateArg: EntityStateActionArgMapType[EntityTypeArg][StateTypeArg],
          ...args: unknown[]
        ) => any
      >
    ) => {
      this.addStateAction(
        target.constructor as typeof ActionsBase,
        propertyKey,
        entityType,
        stateType
      );
    };
  }

  // aliases for easy access
  public static PersistentTableAction() {
    return this.Action(EntityType.Table, StateType.Persistent);
  }
  public static DerivedTableAction() {
    return this.Action(EntityType.Table, StateType.Derived);
  }
  public static PersistentModelAction() {
    return this.Action(EntityType.Model, StateType.Persistent);
  }
  public static DerivedModelAction() {
    return this.Action(EntityType.Model, StateType.Derived);
  }
  public static ApplicationAction() {
    return this.Action(EntityType.Application, StateType.Derived);
  }
  public static MetricsDefinitionAction() {
    return this.Action(
      EntityType.MetricsDefinition,
      StateType.Persistent
    ) as any;
  }
  public static MeasureDefinitionAction() {
    return this.Action(
      EntityType.MeasureDefinition,
      StateType.Persistent
    ) as any;
  }
  public static DimensionDefinitionAction() {
    return this.Action(
      EntityType.DimensionDefinition,
      StateType.Persistent
    ) as any;
  }

  /**
   * Marks a method as a generic action on a state type.
   * Takes just a {@link StateType}. The method has to take {@link EntityType} as the 2nd argument.
   * @param stateType
   */
  public static GenericStateAction<StateTypeArg extends StateType>(
    stateType: StateTypeArg
  ) {
    return (
      target: ActionsBase,
      propertyKey: string,
      // make sure the decorator and the state action arg match using this
      _descriptor: TypedPropertyDescriptor<
        (
          stateArg: EntityStateActionArg<EntityRecord>,
          entityType: EntityType,
          ...args: unknown[]
        ) => any
      >
    ) => {
      this.addStateAction(
        target.constructor as typeof ActionsBase,
        propertyKey,
        undefined,
        stateType
      );
    };
  }

  public static PersistentAction() {
    return this.GenericStateAction(StateType.Persistent);
  }
  public static DerivedAction() {
    return this.GenericStateAction(StateType.Derived);
  }

  /**
   * Marks a method as a generic action.
   * EntityType and StateType are passed as 1st and 2nd arguments respectively.
   */
  public static GenericAction() {
    return (
      target: ActionsBase,
      propertyKey: string,
      // make sure the decorator and the state action arg match using this
      _descriptor: TypedPropertyDescriptor<
        (
          stateArg: EntityStateActionArg<EntityRecord>,
          entityType: EntityType,
          stateType: StateType,
          ...args: unknown[]
        ) => any
      >
    ) => {
      this.addStateAction(
        target.constructor as typeof ActionsBase,
        propertyKey,
        undefined,
        undefined
      );
    };
  }

  private static addStateAction(
    clazz: typeof ActionsBase,
    propertyKey: string,
    entityType: EntityType,
    stateType: StateType
  ) {
    if (!Object.prototype.hasOwnProperty.call(clazz, "actionToStateTypesMap")) {
      clazz.actionToStateTypesMap = { ...clazz.actionToStateTypesMap };
    }
    clazz.actionToStateTypesMap[propertyKey] = [entityType, stateType];
  }

  /**
   * Resets state to idle for the entity on exit
   */
  public static ResetStateToIdle(entityType: EntityType) {
    return (
      target: ActionsBase,
      propertyKey: string,
      // make sure the decorator and the state action arg match using this
      descriptor: TypedPropertyDescriptor<
        (
          stateArg: EntityStateActionArg<EntityRecord>,
          entityId: string,
          ...args: unknown[]
        ) => any
      >
    ) => {
      const previousMethod = descriptor.value;
      descriptor.value = async function (
        stateArg: EntityStateActionArg<EntityRecord>,
        entityId: string,
        ...args: unknown[]
      ): Promise<any> {
        try {
          const resp = await previousMethod.call(
            this,
            stateArg,
            entityId,
            ...args
          );
          this.dataModelerStateService.dispatch("setEntityStatus", [
            entityType,
            entityId,
            EntityStatus.Idle,
          ]);
          await this.dataModelerStateService.waitForAllUpdates(
            entityType,
            StateType.Derived
          );
          return resp;
        } catch (err) {
          this.dataModelerStateService.dispatch("setEntityStatus", [
            entityType,
            entityId,
            EntityStatus.Idle,
          ]);
          throw err;
        }
      };
      return descriptor;
    };
  }
}
