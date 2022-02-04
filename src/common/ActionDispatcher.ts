import type {DataModelerState} from "$lib/types";

// pick only action functions from handler
export type PickActionFunctions<Handler> = Pick<Handler, {
    [Action in keyof Handler]: Handler[Action] extends
        (draftState: DataModelerState, ...args: any[]) => void | Promise<void> ? Action : never
}[keyof Handler]>;
// converts handler to Map of "Action Type" to "Array of args to the action"
export type ExtractActionTypeDefinitions<Handler> = {
    [Action in keyof Handler]: Handler[Action] extends
        (draftState: DataModelerState, ...args: infer Args) => void | Promise<void> ? Args : never
};

export function getActionMethods(instance: any): Array<string> {
    return Object.getOwnPropertyNames(instance.constructor.prototype).filter((prototypeMember) => {
        const descriptor = Object.getOwnPropertyDescriptor(instance.constructor.prototype, prototypeMember);
        return prototypeMember !== "constructor" && typeof descriptor.value === "function";
    });
}
