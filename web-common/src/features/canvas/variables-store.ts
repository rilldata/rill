import { type V1ComponentVariable } from "@rilldata/web-common/runtime-client";
import { type Readable, derived, writable } from "svelte/store";

interface ComponentVariable extends V1ComponentVariable {
  currentValue?: unknown;
}
export interface CanvasStoreType {
  canvases: Record<string, ComponentVariable[]>;
}
const { update, subscribe } = writable({
  canvases: {},
} as CanvasStoreType);

export const updateCanvasByName = (
  name: string,
  callback: (variables: ComponentVariable[]) => void,
) => {
  update((state) => {
    if (!state.canvases[name]) {
      return state;
    }

    callback(state.canvases[name]);
    return state;
  });
};

const canvasVariableReducers = {
  init(name: string, variables: ComponentVariable[]) {
    update((state) => {
      if (state.canvases[name]) return state;
      state.canvases[name] = variables;
      return state;
    });
  },
  remove(name: string) {
    update((state) => {
      delete state.canvases[name];
      return state;
    });
  },
  updateVariable(name: string, variableName: string, value: unknown) {
    updateCanvasByName(name, (variables) => {
      const variable = variables.find((v) => v.name === variableName);

      if (variable) {
        variable.currentValue = value;
      }
    });
  },
};

export const canvasVariablesStore: Readable<CanvasStoreType> &
  typeof canvasVariableReducers = {
  subscribe,
  ...canvasVariableReducers,
};

export function useVariableStore(name: string): Readable<ComponentVariable[]> {
  return derived(canvasVariablesStore, ($store) => {
    return $store.canvases[name];
  });
}

export function useVariable(
  name: string,
  variableName: string,
): Readable<unknown> {
  return derived(canvasVariablesStore, ($store) => {
    const variables = $store.canvases[name] || [];
    const variable = variables.find((v) => v.name === variableName);

    return variable?.currentValue || variable?.defaultValue;
  });
}

export function useVariableInputParams(
  name: string,
  inputParams: V1ComponentVariable[] | undefined,
): Readable<Record<string, unknown>> {
  return derived(canvasVariablesStore, ($store) => {
    if (!inputParams || !inputParams?.length) return {};

    const result: Record<string, unknown> = {};
    const variables: ComponentVariable[] = $store.canvases?.[name] || [];

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
