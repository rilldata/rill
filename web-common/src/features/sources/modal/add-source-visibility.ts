import { resetConnectorStep } from "./connectorStepStore";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";

export const addSourceModal = (() => {
  return {
    open: () => {
      const state = { step: 1, connector: null, requestConnector: false };
      window.history.pushState(state, "", "");
      dispatchEvent(new PopStateEvent("popstate", { state: state }));
    },
    openWithConnector: (
      connector: V1ConnectorDriver,
      schemaName: string,
    ) => {
      resetConnectorStep();
      const state = {
        step: 2,
        selectedConnector: connector,
        schemaName: schemaName,
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
