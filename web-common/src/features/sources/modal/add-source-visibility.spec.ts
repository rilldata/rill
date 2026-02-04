import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { addSourceModal } from "./add-source-visibility";

// Mock the connectorStepStore
vi.mock("./connectorStepStore", () => ({
  resetConnectorStep: vi.fn(),
  setStep: vi.fn(),
  setConnectorInstanceName: vi.fn(),
}));

// Mock the connector-schemas
vi.mock("./connector-schemas", () => ({
  getConnectorSchema: vi.fn((name: string) => {
    // Return mock schemas based on connector name
    if (name === "bigquery" || name === "snowflake") {
      return { "x-category": "warehouse" };
    }
    if (name === "s3" || name === "gcs") {
      return { "x-category": "objectStore" };
    }
    return null;
  }),
  hasExplorerStep: vi.fn((schema) => {
    const category = schema?.["x-category"];
    return category === "sqlStore" || category === "warehouse";
  }),
}));

describe("addSourceModal", () => {
  let pushStateSpy: ReturnType<typeof vi.spyOn>;
  let dispatchEventSpy: ReturnType<typeof vi.spyOn>;
  let capturedState: unknown;

  beforeEach(() => {
    // Spy on window.history.pushState
    pushStateSpy = vi.spyOn(window.history, "pushState").mockImplementation(
      (state) => {
        capturedState = state;
      },
    );

    // Spy on window.dispatchEvent
    dispatchEventSpy = vi
      .spyOn(window, "dispatchEvent")
      .mockImplementation(() => true);

    capturedState = null;
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("open()", () => {
    it("should open step 1 when called without parameters", () => {
      addSourceModal.open();

      expect(pushStateSpy).toHaveBeenCalledWith(
        {
          step: 1,
          connector: null,
          connectorInstanceName: null,
          requestConnector: false,
        },
        "",
        "",
      );
      expect(dispatchEventSpy).toHaveBeenCalled();
    });

    it("should open step 2 when called with connector name", () => {
      addSourceModal.open("bigquery");

      expect(pushStateSpy).toHaveBeenCalledWith(
        {
          step: 2,
          connector: "bigquery",
          connectorInstanceName: null,
          requestConnector: false,
        },
        "",
        "",
      );
    });

    it("should open step 2 with connector instance name when both provided", () => {
      addSourceModal.open("bigquery", "my-bq-connector");

      expect(pushStateSpy).toHaveBeenCalledWith(
        {
          step: 2,
          connector: "bigquery",
          connectorInstanceName: "my-bq-connector",
          requestConnector: false,
        },
        "",
        "",
      );
    });

    it("should set explorer step for warehouse connectors", async () => {
      const { setStep } = await import("./connectorStepStore");

      addSourceModal.open("bigquery");

      expect(setStep).toHaveBeenCalledWith("explorer");
    });

    it("should set source step for object store connectors", async () => {
      const { setStep } = await import("./connectorStepStore");
      const { hasExplorerStep } = await import("./connector-schemas");

      // Re-mock for object store
      vi.mocked(hasExplorerStep).mockReturnValueOnce(false);

      addSourceModal.open("s3");

      expect(setStep).toHaveBeenCalledWith("source");
    });

    it("should reset connector step when called without parameters", async () => {
      const { resetConnectorStep } = await import("./connectorStepStore");

      addSourceModal.open();

      expect(resetConnectorStep).toHaveBeenCalled();
    });

    it("should set connector instance name when provided", async () => {
      const { setConnectorInstanceName } = await import("./connectorStepStore");

      addSourceModal.open("bigquery", "my-bq-instance");

      expect(setConnectorInstanceName).toHaveBeenCalledWith("my-bq-instance");
    });
  });

  describe("openWithConnector()", () => {
    it("should open step 2 with full connector object", () => {
      const connector = {
        name: "bigquery",
        displayName: "BigQuery",
        implementsWarehouse: true,
      };

      addSourceModal.openWithConnector(connector, "bigquery");

      expect(pushStateSpy).toHaveBeenCalledWith(
        {
          step: 2,
          selectedConnector: connector,
          schemaName: "bigquery",
          requestConnector: false,
        },
        "",
        "",
      );
    });

    it("should reset connector step", async () => {
      const { resetConnectorStep } = await import("./connectorStepStore");

      addSourceModal.openWithConnector({ name: "test" }, "test");

      expect(resetConnectorStep).toHaveBeenCalled();
    });
  });

  describe("close()", () => {
    it("should reset to step 0", () => {
      addSourceModal.close();

      expect(pushStateSpy).toHaveBeenCalledWith(
        {
          step: 0,
          connector: null,
          connectorInstanceName: null,
          requestConnector: false,
        },
        "",
        "",
      );
    });

    it("should reset connector step", async () => {
      const { resetConnectorStep } = await import("./connectorStepStore");

      addSourceModal.close();

      expect(resetConnectorStep).toHaveBeenCalled();
    });

    it("should dispatch popstate event", () => {
      addSourceModal.close();

      expect(dispatchEventSpy).toHaveBeenCalled();
      const event = dispatchEventSpy.mock.calls[0][0] as PopStateEvent;
      expect(event.type).toBe("popstate");
    });
  });
});
