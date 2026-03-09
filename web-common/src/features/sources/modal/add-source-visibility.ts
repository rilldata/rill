import {
  resetConnectorStep,
  setStep,
  setConnectorInstanceName,
} from "./connectorStepStore";
import { getConnectorSchema, hasExplorerStep } from "./connector-schemas";

export const addSourceModal = (() => {
  return {
    open: (schemaName?: string, connectorInstanceName?: string) => {
      // If connector is pre-selected, skip to the appropriate step (source or explorer)
      // This assumes the connector is already configured (e.g., gcs.yaml exists)
      if (schemaName) {
        // Determine which step to skip to based on connector type
        const schema = getConnectorSchema(schemaName);
        const hasExplorer = hasExplorerStep(schema);
        // For connectors with explorer step (warehouses/databases), skip to explorer
        // For multi-step connectors (object stores), skip to source
        // This matches the skip link behavior
        const targetStep = hasExplorer ? "explorer" : "source";
        setStep(targetStep);

        if (connectorInstanceName) {
          setConnectorInstanceName(connectorInstanceName);
        }
      } else {
        // Reset to connector step when opening without a pre-selected connector
        resetConnectorStep();
      }

      const state = {
        step: schemaName ? 2 : 1, // Always skip to step 2 if connector is pre-selected
        connector: schemaName ?? null,
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
