import type {DataModelerState} from "$lib/types";
import type {TableStateActions} from "./TableStateActions";
import type {ModelStateActions} from "./ModelStateActions";
import type {ProfileColumnStateActions} from "./ProfileColumnStateActions";
import type {ExtractActionTypeDefinitions} from "$common/ServiceBase";
import type {StateActions} from "$common/data-modeler-state-service/StateActions";
import { writable, get } from "svelte/store";
import type {Writable} from "svelte/store";
import produce, {enablePatches, applyPatches} from "immer";
import type {Patch} from "immer";
import {initialState} from "../stateInstancesFactory";
import {getActionMethods} from "$common/ServiceBase";
import type {PickActionFunctions} from "$common/ServiceBase";
import type { RootConfig } from "$common/config/RootConfig";
import type {
    EntityRecord,
    EntityStateActionArg,
    EntityStateService,
    EntityType,
    StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
    EntityRecordMapType,
    EntityStateServicesMapType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateServicesMap";

enablePatches();

type DataModelerStateActionsClasses = PickActionFunctions<EntityStateActionArg<any>, (
    TableStateActions &
    ModelStateActions &
    ProfileColumnStateActions
)>;
export type DataModelerStateActionsDefinition = ExtractActionTypeDefinitions<EntityStateActionArg<any>, DataModelerStateActionsClasses>;

export type PatchesSubscriber = (patches: Array<Patch>, inversePatches?: Array<Patch>) => void;

/**
 * Lower order actions that update the data modeler state directly and somewhat atomically.
 * Use dispatch for taking actions.
 *
 * Takes an array of {@link StateActions} instances.
 * Actions supported is dependent on these instances passed in the constructor.
 * One caveat to note, type definition and actual instances passed might not match.
 *
 * Emits immer patches. These patches are forwarded to client by {@link SocketServer}
 */
export class DataModelerStateService {
    public store: Writable<DataModelerState>;

    private readonly actionsMap: {
        [Action in keyof DataModelerStateActionsDefinition]?: DataModelerStateActionsClasses
    } = {};
    private readonly entityStateServicesMap: EntityStateServicesMapType = {};

    private patchesSubscribers: Array<PatchesSubscriber> = [];

    public constructor(private readonly stateActions: Array<StateActions>,
                       private readonly entityStateServices: Array<EntityStateService<any>>,
                       protected readonly config?: RootConfig) {
        stateActions.forEach((actions) => {
            getActionMethods(actions).forEach(action => {
                this.actionsMap[action] = actions;
            });
        });
        entityStateServices.forEach((entityStateService) => {
            this.entityStateServicesMap[entityStateService.entityType] ??= {};
            (this.entityStateServicesMap[entityStateService.entityType] as any)[entityStateService.stateType] = entityStateService;
        });
    }

    public init(): void {
        this.store = writable(initialState());
    }

    public destroy(): void {}

    public getCurrentState(): DataModelerState {
        return get(this.store);
    }

    /**
     * Subscribe to underlying store
     * @param subscriber
     */
    public subscribe(subscriber: (dataModelerState: DataModelerState) => void): void {
        this.store.subscribe(subscriber);
    }

    /**
     * Subscribe to patch emitted by immer.
     * @param subscriber
     */
    public subscribePatches(subscriber: PatchesSubscriber): void {
        this.patchesSubscribers.push(subscriber);
    }

    public updateState(dataModelerState: DataModelerState): void {
        this.store.set(dataModelerState);
    }

    /**
     * Forwards action to the appropriate class.
     * @param action
     * @param args
     */
    public dispatch<Action extends keyof DataModelerStateActionsDefinition>(
        action: Action, args: DataModelerStateActionsDefinition[Action],
    ): void {
        if (!this.actionsMap[action]?.[action]) {
            console.log(`${action} not found`);
            return;
        }
        const currentState = this.getCurrentState();
        this.updateState(produce(currentState, (draft) => {
            const actionsInstance = this.actionsMap[action];
            actionsInstance[action].call(actionsInstance, draft, ...args);
        }, (patches, inversePatches) => {
            this.patchesSubscribers.forEach(subscriber => subscriber(patches, inversePatches));
            // we can later add a subscriber to store patches and inversePatches into some store
        }));
    }

    public applyPatches(patches: Array<Patch>): void {
        this.updateState(applyPatches(this.getCurrentState(), patches));
    }

    public getEntityById<EntityTypeArg extends EntityType, StateTypeArg extends StateType>(
        entityType: EntityTypeArg, stateType: StateTypeArg, entityId: string,
    ): EntityRecordMapType[EntityTypeArg][StateTypeArg] {
        return this.entityStateServicesMap[entityType][stateType].getById(entityId) as any;
    }

    public addEntities(entityType: EntityType,
                       stateTypeEntities: Array<[StateType, EntityRecord]>,
                       atIndex?: number): void {
        for (const [stateType, entityRecord] of stateTypeEntities) {
            const service = this.entityStateServicesMap[entityType][stateType];
            this.updateStateAndEmitPatches(service, (draft) => {
                service.addEntity(draft, entityRecord as any, atIndex);
            });
        }
    }

    public deleteEntities(entityType: EntityType, stateTypes: Array<StateType>,
                          entityId: string): void {
        for (const stateType of stateTypes) {
            const service = this.entityStateServicesMap[entityType][stateType];
            this.updateStateAndEmitPatches(service, (draft) => {
                service.deleteEntity(draft, entityId);
            });
        }
    }

    public moveEntitiesUp(entityType: EntityType, stateTypes: Array<StateType>,
                          entityId: string): void {
        for (const stateType of stateTypes) {
            const service = this.entityStateServicesMap[entityType][stateType];
            this.updateStateAndEmitPatches(service, (draft) => {
                service.moveEntityUp(draft, entityId);
            });
        }
    }

    public moveEntitiesDown(entityType: EntityType, stateTypes: Array<StateType>,
                            entityId: string): void {
        for (const stateType of stateTypes) {
            const service = this.entityStateServicesMap[entityType][stateType];
            this.updateStateAndEmitPatches(service, (draft) => {
                service.moveEntityDown(draft, entityId);
            });
        }
    }

    private updateStateAndEmitPatches(service: EntityStateService<any>,
                                      callback: (draft) => void) {
        service.updateState(callback, (patches) => {
            this.patchesSubscribers.forEach(subscriber => subscriber(patches));
        });
    }
}
