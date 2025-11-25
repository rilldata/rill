import { writable } from "svelte/store";

export type ConnectorStep = "connector" | "source";

export const connectorStepStore = writable<{
  step: ConnectorStep;
  connectorConfig: Record<string, unknown> | null;
  connectorInstanceName: string | null;
}>({
  step: "connector",
  connectorConfig: null,
  connectorInstanceName: null,
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

export function resetConnectorStep() {
  connectorStepStore.set({
    step: "connector",
    connectorConfig: null,
    connectorInstanceName: null,
  });
}
