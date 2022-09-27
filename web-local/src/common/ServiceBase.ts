/**
 * Picks only action functions from handler.
 * Action function is identified based on FirstArg type.
 */
export type PickActionFunctions<FirstArg, Handler> = Pick<
  Handler,
  {
    [Action in keyof Handler]: Handler[Action] extends (
      firstArg: FirstArg,
      ...args: unknown[]
    ) => unknown
      ? Action
      : never;
  }[keyof Handler]
>;
/**
 * Converts handler to Map of "Action Type" to "Array of args to the action"
 * Handler is identified based on FirstArg type.
 */
export type ExtractActionTypeDefinitions<FirstArg, Handler> = {
  [Action in keyof Handler]: Handler[Action] extends (
    firstArg: FirstArg,
    ...args: infer Args
  ) => unknown
    ? Args
    : never;
};

export function getActionMethods(instance: unknown): Array<string> {
  return Object.getOwnPropertyNames(instance.constructor.prototype).filter(
    (prototypeMember) => {
      const descriptor = Object.getOwnPropertyDescriptor(
        instance.constructor.prototype,
        prototypeMember
      );
      return (
        prototypeMember !== "constructor" &&
        typeof descriptor.value === "function"
      );
    }
  );
}

export interface ActionServiceBase<
  ActionsDefinition extends Record<string, Array<unknown>>
> {
  /**
   * Will be called by ActionQueueOrchestrator once the action has been scheduled
   * @param action
   * @param args
   */
  dispatch<Action extends keyof ActionsDefinition>(
    action: Action,
    args: ActionsDefinition[Action]
  ): Promise<unknown>;
}
