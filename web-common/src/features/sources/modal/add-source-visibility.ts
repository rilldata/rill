import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import { resetConnectorStep, setStep } from "./connectorStepStore";

export const addSourceModal = (() => {
  return {
    open: () => {
      const state = { step: 1, connector: null, requestConnector: false };
      window.history.pushState(state, "", "");
      dispatchEvent(new PopStateEvent("popstate", { state: state }));
    },
    close: () => {
      const state = { step: 0, connector: null, requestConnector: false };
      window.history.pushState(state, "", "");
      dispatchEvent(new PopStateEvent("popstate", { state: state }));
      resetConnectorStep();
    },
    /**
     * Open the Data Explorer modal for a specific OLAP connector.
     * This directly opens the explorer step without requiring form submission.
     */
    openExplorerForConnector: (
      connector: V1ConnectorDriver,
      connectorInstanceName?: string,
    ) => {
      // Reset any previous state
      resetConnectorStep();
      // Set the step to explorer before opening the modal
      setStep("explorer");
      // Open the modal at step 2 with the selected connector
      // schemaName is the driver name for form schema lookup
      // connectorInstanceName is the specific instance to pre-select in explorer
      const state = {
        step: 2,
        selectedConnector: connector,
        schemaName: connector.name, // Driver name for schema lookup
        connectorInstanceName: connectorInstanceName, // Instance name to pre-select
        requestConnector: false,
      };
      window.history.pushState(state, "", "");
      dispatchEvent(new PopStateEvent("popstate", { state: state }));
    },
  };
})();
