import { writable } from "svelte/store";

export type ConnectorStep = "connector" | "source";

export const connectorStepStore = writable<{
  step: ConnectorStep;
  connectorConfig: Record<string, unknown> | null;
}>({
  step: "connector",
  connectorConfig: null,
});

export function setStep(step: ConnectorStep) {
  connectorStepStore.update((state) => ({ ...state, step }));
}

export function setConnectorConfig(config: Record<string, unknown>) {
  connectorStepStore.update((state) => ({ ...state, connectorConfig: config }));
}

export function resetConnectorStep() {
  connectorStepStore.set({
    step: "connector",
    connectorConfig: null,
  });
}
