import type {ColumnarItem, DataModelerState, Table, Item, Model, ProfileColumn} from "$lib/types";

export class StateActions {
    protected static getByID<I extends Item>(items: (I[]), id: string): I | null {
        return items.find(item => item.id === id);
    }

    protected static getTable(state: DataModelerState, id: string): Table | null {
        return this.getByID(state.tables, id);
    }

    protected static getModel(state: DataModelerState, id: string): Model | null {
        return this.getByID(state.models, id);
    }

    protected static getProfile(items: ColumnarItem[], modelId: string, name: string): ProfileColumn | null {
        const model = this.getByID(items, modelId);
        return model.profile.find(profile => profile.name === name);
    }

    protected static shallowCopy(source: Record<string, any>, target: Record<string, any>): void {
        Object.keys(source).forEach((k) => {
            target[k] = source[k];
        });
    }
}
