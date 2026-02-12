import { resetConnectorStep } from "./connectorStepStore";
import { toConnectorDriver } from "./connector-schemas";

export const addSourceModal = (() => {
  return {
    open: () => {
      const state = { step: 1, connector: null, requestConnector: false };
      window.history.pushState(state, "", "");
      dispatchEvent(new PopStateEvent("popstate", { state: state }));
    },
    /**
     * Open the modal directly at step 2 for a specific connector schema.
     * Used for AI connectors in the "Add > More > AI Connector" menu.
     */
    openForConnector: (schemaName: string) => {
      resetConnectorStep();
      const selectedConnector = toConnectorDriver(schemaName);
      if (!selectedConnector) return;

      const state = {
        step: 2,
        selectedConnector,
        schemaName,
        requestConnector: false,
      };
      window.history.pushState(state, "", "");
      dispatchEvent(new PopStateEvent("popstate", { state }));
    },
    close: () => {
      const state = { step: 0, connector: null, requestConnector: false };
      window.history.replaceState(state, "", "");
      dispatchEvent(new PopStateEvent("popstate", { state: state }));
      resetConnectorStep();
    },
  };
})();
