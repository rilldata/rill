import type { Writable } from "svelte/store";
import { getContext, setContext } from "svelte";
import type { TimeDimensionDetailsStore } from "./time-dimension-details-store";
import type { THeaderData } from "./mock-data";

const CONTEXT_KEY = "tdd-store";

type TDDContext = {
  store: Writable<TimeDimensionDetailsStore>;
  headers: THeaderData[];
};
export function createTDDContext(value: TDDContext) {
  setContext(CONTEXT_KEY, value);
}
export function useTDDContext() {
  return getContext<TDDContext>(CONTEXT_KEY);
}
