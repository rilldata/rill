import type { Writable } from "svelte/store";
import { getContext, setContext } from "svelte";
import type { TimeDimensionDetailsStore } from "./time-dimension-details-store";
import type { THeaderData } from "./mock-data";

const CONTEXT_KEY = "tdt-store";

type TDTContext = {
  store: Writable<TimeDimensionDetailsStore>;
  headers: THeaderData[];
};
export function createTDTContext(value: TDTContext) {
  setContext(CONTEXT_KEY, value);
}
export function useTDTContext() {
  return getContext<TDTContext>(CONTEXT_KEY);
}
