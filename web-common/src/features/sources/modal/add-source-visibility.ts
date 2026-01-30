import { resetConnectorStep, setStep } from "./connectorStepStore";

export const addSourceModal = (() => {
  return {
    open: (connectorName?: string, connectorInstanceName?: string) => {
      // If connector is pre-selected, skip to the "source" step (import form)
      // This assumes the connector is already configured (e.g., gcs.yaml exists)
      if (connectorName) {
        setStep("source");
      }

      const state = {
        step: connectorName ? 2 : 1, // Skip to step 2 if connector is pre-selected
        connector: connectorName ?? null,
        connectorInstanceName: connectorInstanceName ?? null,
        requestConnector: false,
      };
      window.history.pushState(state, "", "");
      dispatchEvent(new PopStateEvent("popstate", { state: state }));
    },
    close: () => {
      const state = {
        step: 0,
        connector: null,
        connectorInstanceName: null,
        requestConnector: false,
      };
      window.history.pushState(state, "", "");
      dispatchEvent(new PopStateEvent("popstate", { state: state }));
      resetConnectorStep();
    },
  };
})();
