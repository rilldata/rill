import { writable } from "svelte/store";

export type ConnectorStep = "connector" | "source";

export type ConnectorStepState = {
  step: ConnectorStep;
  connectorConfig: Record<string, unknown> | null;
  connectorInstanceName: string | null;
  selectedAuthMethod: string | null;
};

export const connectorStepStore = writable<ConnectorStepState>({
  step: "connector",
  connectorConfig: null,
  connectorInstanceName: null,
  selectedAuthMethod: null,
});

export function setStep(step: ConnectorStep) {
  connectorStepStore.update((state) => ({ ...state, step }));
}

export function setConnectorConfig(config: Record<string, unknown>) {
  connectorStepStore.update((state) => ({ ...state, connectorConfig: config }));
}

export function setConnectorInstanceName(name: string | null) {
  connectorStepStore.update((state) => ({
    ...state,
    connectorInstanceName: name,
  }));
}

export function setAuthMethod(method: string | null) {
  connectorStepStore.update((state) => ({
    ...state,
    selectedAuthMethod: method,
  }));
}

export function resetConnectorStep() {
  connectorStepStore.set({
    step: "connector",
    connectorConfig: null,
    connectorInstanceName: null,
    selectedAuthMethod: null,
  });
}
