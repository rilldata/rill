import { describe, it, expect, beforeEach, vi, afterEach } from "vitest";
import { addSourceModal } from "./add-source-visibility";
import * as connectorStepStore from "./connectorStepStore";

describe("addSourceModal", () => {
  let pushStateSpy: ReturnType<typeof vi.spyOn>;
  let dispatchEventSpy: ReturnType<typeof vi.spyOn>;
  let setStepSpy: ReturnType<typeof vi.spyOn>;
  let resetConnectorStepSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    pushStateSpy = vi.spyOn(window.history, "pushState").mockImplementation(() => {});
    dispatchEventSpy = vi.spyOn(window, "dispatchEvent").mockImplementation(() => true);
    setStepSpy = vi.spyOn(connectorStepStore, "setStep").mockImplementation(() => {});
    resetConnectorStepSpy = vi.spyOn(connectorStepStore, "resetConnectorStep").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("open", () => {
    it("should open modal with step 1 when no connector provided", () => {
      addSourceModal.open();

      expect(pushStateSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          step: 1,
          connector: null,
          connectorInstanceName: null,
          requestConnector: false,
        }),
        "",
        ""
      );
    });

    it("should open modal with step 2 and connector when connector name provided", () => {
      addSourceModal.open("gcs");

      expect(setStepSpy).toHaveBeenCalledWith("source");
      expect(pushStateSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          step: 2,
          connector: "gcs",
          connectorInstanceName: null,
          requestConnector: false,
        }),
        "",
        ""
      );
    });

    it("should pass connector instance name when both connector and instance name provided", () => {
      addSourceModal.open("gcs", "gcs_1");

      expect(setStepSpy).toHaveBeenCalledWith("source");
      expect(pushStateSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          step: 2,
          connector: "gcs",
          connectorInstanceName: "gcs_1",
          requestConnector: false,
        }),
        "",
        ""
      );
    });

    it("should dispatch popstate event with correct state", () => {
      addSourceModal.open("s3", "s3_custom");

      expect(dispatchEventSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          type: "popstate",
          state: expect.objectContaining({
            connector: "s3",
            connectorInstanceName: "s3_custom",
          }),
        })
      );
    });
  });

  describe("close", () => {
    it("should reset state and call resetConnectorStep", () => {
      addSourceModal.close();

      expect(pushStateSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          step: 0,
          connector: null,
          connectorInstanceName: null,
          requestConnector: false,
        }),
        "",
        ""
      );
      expect(resetConnectorStepSpy).toHaveBeenCalled();
    });
  });
});
