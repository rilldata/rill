import { describe, it, expect, beforeEach } from "vitest";
import { addSourceModal } from "./add-source-visibility";
import { resetConnectorStep, connectorStepStore, get } from "./connectorStepStore";

describe("addSourceModal", () => {
  beforeEach(() => {
    resetConnectorStep();
  });

  describe("open", () => {
    it("should set step to explorer for connectors with explorer step (snowflake)", () => {
      addSourceModal.open("snowflake", "snowflake_1");

      expect(get(connectorStepStore).step).toBe("explorer");
      expect(get(connectorStepStore).connectorInstanceName).toBe("snowflake_1");
    });

    it("should set step to source for multi-step connectors (gcs)", () => {
      addSourceModal.open("gcs", "gcs_1");

      expect(get(connectorStepStore).step).toBe("source");
      expect(get(connectorStepStore).connectorInstanceName).toBe("gcs_1");
    });

    it("should not set connectorInstanceName if not provided", () => {
      addSourceModal.open("gcs");

      expect(get(connectorStepStore).step).toBe("source");
      expect(get(connectorStepStore).connectorInstanceName).toBeNull();
    });

    it("should reset step store when opening without connector", () => {
      // Set some state first
      addSourceModal.open("gcs", "gcs_1");
      expect(get(connectorStepStore).step).toBe("source");

      // Open without connector - should reset
      addSourceModal.open();

      // Verify store is reset to connector step
      expect(get(connectorStepStore).step).toBe("connector");
    });
  });

  describe("close", () => {
    it("should reset connector step store", () => {
      // Set some state first
      addSourceModal.open("gcs", "gcs_1");
      expect(get(connectorStepStore).step).toBe("source");
      expect(get(connectorStepStore).connectorInstanceName).toBe("gcs_1");

      // Close the modal
      addSourceModal.close();

      // Verify store is reset
      expect(get(connectorStepStore).step).toBe("connector");
      expect(get(connectorStepStore).connectorInstanceName).toBeNull();
    });
  });
});
