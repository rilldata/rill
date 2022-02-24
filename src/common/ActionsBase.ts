import {
    EntityStateActionArg,
    EntityType,
    StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
    EntityStateActionArgMapType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateServicesMap";

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
        StateTypeArg extends StateType,
    >(entityType: EntityTypeArg, stateType: StateTypeArg) {
        return (target: ActionsBase, propertyKey: string,
                // make sure the decorator and the state action arg match using this
                descriptor: TypedPropertyDescriptor<
                    (stateArg: EntityStateActionArgMapType[EntityTypeArg][StateTypeArg], ...args: any[]) => any
                >) => {
            this.addStateAction(target.constructor as typeof ActionsBase, propertyKey, entityType, stateType);
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

    /**
     * Marks a method as a generic action.
     * Takes just a {@link StateType}. The method has to take {@link EntityType} as the 2nd argument.
     * @param stateType
     */
    public static GenericAction<StateTypeArg extends StateType>(stateType: StateTypeArg) {
        return (target: ActionsBase, propertyKey: string,
                // make sure the decorator and the state action arg match using this
                descriptor: TypedPropertyDescriptor<
                    (stateArg: EntityStateActionArg<any>, entityType: EntityType, ...args: any[]) => any
                >) => {
            this.addStateAction(target.constructor as typeof ActionsBase, propertyKey,
                undefined, stateType);
        };
    }

    public static PersistentAction() {
        return this.GenericAction(StateType.Persistent);
    }
    public static DerivedAction() {
        return this.GenericAction(StateType.Derived);
    }

    private static addStateAction(clazz: typeof ActionsBase, propertyKey: string,
                                  entityType: EntityType, stateType: StateType) {
        if (!Object.prototype.hasOwnProperty.call(clazz, "actionToStateTypesMap")) {
            clazz.actionToStateTypesMap = {...clazz.actionToStateTypesMap};
        }
        clazz.actionToStateTypesMap[propertyKey] = [entityType, stateType];
    }
}
