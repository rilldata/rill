import { resetConnectorStep } from "./connectorStepStore";

export const addSourceModal = (() => {
  return {
    open: (connectorName?: string) => {
      const state = {
        step: connectorName ? 2 : 1, // Skip to step 2 if connector is pre-selected
        connector: connectorName ?? null,
        requestConnector: false,
      };
      window.history.pushState(state, "", "");
      dispatchEvent(new PopStateEvent("popstate", { state: state }));
    },
    close: () => {
      const state = { step: 0, connector: null, requestConnector: false };
      window.history.pushState(state, "", "");
      dispatchEvent(new PopStateEvent("popstate", { state: state }));
      resetConnectorStep();
    },
  };
})();
