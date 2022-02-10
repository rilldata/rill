// pick only action functions from handler
export type PickActionFunctions<FirstArg, Handler> = Pick<Handler, {
    [Action in keyof Handler]: Handler[Action] extends
        (firstArg: FirstArg, ...args: any[]) => any ? Action : never
}[keyof Handler]>;
// converts handler to Map of "Action Type" to "Array of args to the action"
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
