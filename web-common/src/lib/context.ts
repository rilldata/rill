import { getContext, setContext } from "svelte";

export function createContext<T>(key: string = `rill:${crypto.randomUUID()}`) {
  return {
    get: () => getContext<T>(key),
    set: (ctx: T) => setContext(key, ctx),
  };
}
