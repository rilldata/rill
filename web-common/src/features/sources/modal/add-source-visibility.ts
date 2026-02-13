import type { V1ConnectorDriver } from "../../../runtime-client";
import { resetConnectorStep } from "./connectorStepStore";

export const addSourceModal = (() => {
  return {
    open: () => {
      const state = { step: 1, connector: null, requestConnector: false };
      window.history.pushState(state, "", "");
      dispatchEvent(new PopStateEvent("popstate", { state: state }));
    },
    /**
     * Open the modal directly to a specific connector's form (step 2).
     * Used when clicking + on an OLAP connector in the Data Explorer.
     */
    openForConnector: (
      schemaName: string,
      connectorDriver: V1ConnectorDriver,
    ) => {
      resetConnectorStep();
      const state = {
        step: 2,
        selectedConnector: connectorDriver,
        schemaName,
        requestConnector: false,
      };
      window.history.pushState(state, "", "");
      dispatchEvent(new PopStateEvent("popstate", { state }));
    },
    close: () => {
      const state = { step: 0, connector: null, requestConnector: false };
      window.history.pushState(state, "", "");
      dispatchEvent(new PopStateEvent("popstate", { state: state }));
      resetConnectorStep();
    },
  };
})();
