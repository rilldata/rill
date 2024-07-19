import { V1DashboardVariable } from "@rilldata/web-common/runtime-client";
import { Readable, derived, writable } from "svelte/store";

export interface DashboardStoreType {
  dashboards: Record<string, V1DashboardVariable[]>;
}
const { update, subscribe } = writable({
  dashboards: {},
} as DashboardStoreType);

export const updateDashboardByName = (
  name: string,
  callback: (dashboards: V1DashboardVariable[]) => void,
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
  init(name: string, variables: V1DashboardVariable[]) {
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

export function useVariableStore(
  name: string,
): Readable<V1DashboardVariable[]> {
  return derived(dashboardVariablesStore, ($store) => {
    return $store.dashboards[name];
  });
}

export function useVariableInputParams(
  name: string,
  inputParams: Record<string, any>[] | undefined,
): Readable<Record<string, any>> {
  return derived(dashboardVariablesStore, ($store) => {
    const variables = $store.dashboards[name];
    if (!inputParams?.length) return {};

    // const params = inputParams;

    // return variables.reduce((acc, variable) => {
    //   if (inputParams.includes(variable.name)) {
    //     acc[variable.name] = variable.value;
    //   }
    //   return acc;
    // }, {});
  });
}
