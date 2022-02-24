/**
 * Picks only action functions from handler.
 * Action function is identified based on FirstArg type.
 */
export type PickActionFunctions<FixedArgs, Handler> = Pick<Handler, {
    [Action in keyof Handler]: Handler[Action] extends
        (firstArg: FixedArgs, ...args: any[]) => any ? Action : never
}[keyof Handler]>;
/**
 * Converts handler to Map of "Action Type" to "Array of args to the action"
 * Handler is identified based on FirstArg type.
 */
export type ExtractActionTypeDefinitions<FirstArg, Handler> = {
    [Action in keyof Handler]: Handler[Action] extends
        (firstArg: FirstArg, ...args: infer Args) => any ? Args : never
};

export function getActionMethods(instance: any): Array<string> {
    return Object.getOwnPropertyNames(instance.constructor.prototype).filter((prototypeMember) => {
        const descriptor = Object.getOwnPropertyDescriptor(instance.constructor.prototype, prototypeMember);
        return prototypeMember !== "constructor" && typeof descriptor.value === "function";
    });
}
