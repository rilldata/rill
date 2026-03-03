import {
  resetConnectorStep,
  setStep,
  setConnectorInstanceName,
} from "./connectorStepStore";
import { getConnectorSchema, hasExplorerStep } from "./connector-schemas";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";

export const addModelModal = (() => {
  return {
    open: (connectorName?: string, connectorInstanceName?: string) => {
      // If connector is pre-selected, skip to the appropriate step (source or explorer)
      if (connectorName) {
        const schema = getConnectorSchema(connectorName);
        const hasExplorer = hasExplorerStep(schema);
        const targetStep = hasExplorer ? "explorer" : "source";
        setStep(targetStep);

        if (connectorInstanceName) {
          setConnectorInstanceName(connectorInstanceName);
        }
      } else {
        resetConnectorStep();
      }

      const state = {
        modal: "model" as const,
        step: connectorName ? 2 : 1,
        connector: connectorName ?? null,
        connectorInstanceName: connectorInstanceName ?? null,
      };
      window.history.pushState(state, "", "");
      window.dispatchEvent(new PopStateEvent("popstate", { state }));
    },
    openWithConnector: (connector: V1ConnectorDriver, schemaName: string) => {
      resetConnectorStep();
      const state = {
        modal: "model" as const,
        step: 2,
        selectedConnector: connector,
        schemaName: schemaName,
      };
      window.history.pushState(state, "", "");
      window.dispatchEvent(new PopStateEvent("popstate", { state }));
    },
    close: () => {
      const state = {
        modal: "model" as const,
        step: 0,
        connector: null,
        connectorInstanceName: null,
      };
      window.history.pushState(state, "", "");
      window.dispatchEvent(new PopStateEvent("popstate", { state }));
      resetConnectorStep();
    },
  };
})();
