import { V1ComponentVariable } from "@rilldata/web-common/runtime-client";
import { Readable, derived, writable } from "svelte/store";

interface ComponentVariable extends V1ComponentVariable {
  currentValue?: unknown;
}
export interface DashboardStoreType {
  dashboards: Record<string, ComponentVariable[]>;
}
const { update, subscribe } = writable({
  dashboards: {},
} as DashboardStoreType);

export const updateDashboardByName = (
  name: string,
  callback: (dashboards: ComponentVariable[]) => void,
) => {
  update((state) => {
    if (!state.dashboards[name]) {
      return state;
    }

    callback(state.dashboards[name]);
    return state;
  });
};

const dashboardVariableReducers = {
  init(name: string, variables: ComponentVariable[]) {
    update((state) => {
      if (state.dashboards[name]) return state;
      state.dashboards[name] = variables;
      return state;
    });
  },
  remove(name: string) {
    update((state) => {
      delete state.dashboards[name];
      return state;
    });
  },
};

export const dashboardVariablesStore: Readable<DashboardStoreType> &
  typeof dashboardVariableReducers = {
  subscribe,
  ...dashboardVariableReducers,
};

export function useVariableStore(name: string): Readable<ComponentVariable[]> {
  return derived(dashboardVariablesStore, ($store) => {
    return $store.dashboards[name];
  });
}

export function useVariableInputParams(
  name: string,
  inputParams: V1ComponentVariable[] | undefined,
): Readable<Record<string, unknown>> {
  return derived(dashboardVariablesStore, ($store) => {
    const variables = $store.dashboards[name];
    if (!inputParams || !inputParams?.length) return {};

    const result: Record<string, unknown> = {};
    inputParams.forEach((param) => {
      if (!param?.name) return;

      const variable = variables.find((v) => v.name === param.name);
      if (variable) {
        result[param.name] = variable?.currentValue;
      } else {
        result[param.name] = param.defaultValue;
      }
    });

    return result;
  });
}
