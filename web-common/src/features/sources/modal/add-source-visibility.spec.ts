import { afterEach, describe, expect, it, vi } from "vitest";
import { addSourceModal } from "@rilldata/web-common/features/sources/modal/add-source-visibility";

describe("addSourceModal.openForConnector", () => {
  const pushStateSpy = vi.spyOn(window.history, "pushState");
  const dispatchSpy = vi.spyOn(window, "dispatchEvent");

  afterEach(() => {
    pushStateSpy.mockClear();
    dispatchSpy.mockClear();
  });

  it("sets step 2 and implementsAi for claude", () => {
    addSourceModal.openForConnector("claude");

    const state = pushStateSpy.mock.calls[0][0] as Record<string, unknown>;
    expect(state.step).toBe(2);
    expect(state.schemaName).toBe("claude");
    expect(state.requestConnector).toBe(false);

    const connector = state.selectedConnector as Record<string, unknown>;
    expect(connector.name).toBe("claude");
    expect(connector.displayName).toBe("Claude");
    expect(connector.implementsAi).toBe(true);
    expect(connector.implementsWarehouse).toBe(false);
    expect(connector.implementsObjectStore).toBe(false);
  });

  it("sets implementsWarehouse for bigquery", () => {
    addSourceModal.openForConnector("bigquery");

    const state = pushStateSpy.mock.calls[0][0] as Record<string, unknown>;
    const connector = state.selectedConnector as Record<string, unknown>;
    expect(connector.implementsWarehouse).toBe(true);
    expect(connector.implementsAi).toBe(false);
    expect(connector.name).toBe("bigquery");
    expect(connector.displayName).toBe("BigQuery");
  });

  it("dispatches a popstate event with the same state", () => {
    addSourceModal.openForConnector("claude");

    expect(dispatchSpy).toHaveBeenCalledOnce();
    const event = dispatchSpy.mock.calls[0][0] as PopStateEvent;
    expect(event.type).toBe("popstate");
    expect(event.state).toEqual(pushStateSpy.mock.calls[0][0]);
  });
});
