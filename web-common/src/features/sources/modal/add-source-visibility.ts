import { resetConnectorStep } from "./connectorStepStore";

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
  };
})();
