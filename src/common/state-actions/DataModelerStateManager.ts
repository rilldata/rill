import type {DataModelerState} from "$lib/types";
import type {DatasetStateActions} from "./DatasetStateActions";
import type {ModelStateActions} from "./ModelStateActions";
import type {ProfileColumnStateActions} from "./ProfileColumnStateActions";
import type {ExtractActionTypeDefinitions} from "$common/ActionDispatcher";
import type {StateActions} from "$common/state-actions/StateActions";
import { writable, get, Writable } from "svelte/store";
import produce, {enablePatches, applyPatches, Patch} from "immer";
import {initialState} from "../data-factory";
import {getActionMethods, PickActionFunctions} from "$common/ActionDispatcher";

enablePatches();

type StateActionsClasses = PickActionFunctions<DatasetStateActions & ModelStateActions & ProfileColumnStateActions>;
type StateActionsDefinition = ExtractActionTypeDefinitions<StateActionsClasses>;

export type PatchesSubscriber = (patches: Array<Patch>, inversePatches: Array<Patch>) => void;

export class DataModelerStateManager {
    private store: Writable<DataModelerState>;

    private readonly actionsMap: {
        [Action in keyof StateActionsDefinition]?: StateActionsClasses
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

    public subscribe(subscriber: (dataModelerState: DataModelerState) => void): void {
        this.store.subscribe(subscriber);
    }

    public subscribePatches(subscriber: PatchesSubscriber): void {
        this.patchesSubscribers.push(subscriber);
    }

    public dispatch<Action extends keyof StateActionsDefinition>(
        action: Action, args: StateActionsDefinition[Action],
    ): void {
        if (!this.actionsMap[action]?.[action]) {
            console.log(`${action} not found`);
            return;
        }
        const currentState = this.getCurrentState();
        this.store.set(produce(currentState, (draft) => {
            const actionsInstance = this.actionsMap[action];
            actionsInstance[action].call(actionsInstance, draft, ...args);
        }, (patches, inversePatches) => {
            this.patchesSubscribers.forEach(subscriber => subscriber(patches, inversePatches));
            // we can later add a subscriber to store patches and inversePatches into some store
        }));
    }

    public applyPatches(patches: Array<Patch>): void {
        this.store.set(applyPatches(this.getCurrentState(), patches));
    }
}
