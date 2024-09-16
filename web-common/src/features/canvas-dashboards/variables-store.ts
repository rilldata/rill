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
  callback: (variables: ComponentVariable[]) => void,
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
  updateVariable(name: string, variableName: string, value: unknown) {
    updateDashboardByName(name, (variables) => {
      const variable = variables.find((v) => v.name === variableName);

      if (variable) {
        variable.currentValue = value;
      }
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

export function useVariable(
  name: string,
  variableName: string,
): Readable<unknown> {
  return derived(dashboardVariablesStore, ($store) => {
    const variables = $store.dashboards[name] || [];
    const variable = variables.find((v) => v.name === variableName);

    return variable?.currentValue || variable?.defaultValue;
  });
}

export function useVariableInputParams(
  name: string,
  inputParams: V1ComponentVariable[] | undefined,
): Readable<Record<string, unknown>> {
  return derived(dashboardVariablesStore, ($store) => {
    if (!inputParams || !inputParams?.length) return {};

    const result: Record<string, unknown> = {};
    const variables: ComponentVariable[] = $store.dashboards?.[name] || [];

    inputParams.forEach((param) => {
      if (!param?.name) return;

      const variable = variables?.find((v) => v.name === param.name);
      if (variable) {
        result[param.name] = variable?.currentValue || variable?.defaultValue;
      } else {
        result[param.name] = param.defaultValue;
      }
    });

    return result;
  });
}
