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
     *
     * Uses replaceState (not pushState) because the user never saw step 1
     * (connector picker), so Back should return to the page before the modal,
     * not to a step the user never visited.
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
      window.history.replaceState(state, "", "");
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
