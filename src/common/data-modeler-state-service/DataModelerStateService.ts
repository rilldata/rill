import type {DataModelerState} from "$lib/types";
import type {DatasetStateActions} from "./DatasetStateActions";
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

enablePatches();

type DataModelerStateActionsClasses = PickActionFunctions<DataModelerState, (
    DatasetStateActions &
    ModelStateActions &
    ProfileColumnStateActions
)>;
export type DataModelerStateActionsDefinition = ExtractActionTypeDefinitions<DataModelerState, DataModelerStateActionsClasses>;

export type PatchesSubscriber = (patches: Array<Patch>, inversePatches: Array<Patch>) => void;

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

    private patchesSubscribers: Array<PatchesSubscriber> = [];

    public constructor(private readonly stateActions: Array<StateActions>) {
        stateActions.forEach((actions) => {
            getActionMethods(actions).forEach(action => {
                this.actionsMap[action] = actions;
            });
        });
    }

    public init(): void {
        this.store = writable(initialState());
    }

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
}
