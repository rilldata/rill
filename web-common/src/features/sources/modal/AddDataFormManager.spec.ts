import { describe, it, expect, beforeAll, beforeEach } from "vitest";
import { AddDataFormManager } from "./AddDataFormManager";
import {
  resetConnectorStep,
  setStep,
  connectorStepStore,
} from "./connectorStepStore";
import { populateSchemaCache } from "./connector-schemas";
import type { MultiStepFormSchema } from "../../templates/schemas/types";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import { get, writable } from "svelte/store";

const testSchemas: Record<string, MultiStepFormSchema> = {
  gcs: {
    type: "object",
    "x-category": "objectStore",
    properties: {
      google_application_credentials: {
        type: "string",
        "x-step": "connector",
      },
      path: { type: "string", "x-step": "source" },
    },
  } as unknown as MultiStepFormSchema,
  snowflake: {
    type: "object",
    "x-category": "warehouse",
    properties: {
      account: { type: "string", "x-step": "connector" },
      sql: { type: "string", "x-step": "explorer" },
      name: { type: "string", "x-step": "explorer" },
    },
  } as unknown as MultiStepFormSchema,
};

describe("AddDataFormManager", () => {
  beforeAll(() => {
    populateSchemaCache(testSchemas);
  });
  beforeEach(() => {
    resetConnectorStep();
  });

  describe("handleSkip", () => {
    it("should skip to source step for multi-step connectors", () => {
      const connector: V1ConnectorDriver = {
        name: "gcs",
        displayName: "GCS",
        implementsObjectStore: true,
        implementsOlap: false,
        implementsSqlStore: false,
        implementsWarehouse: false,
        implementsFileStore: false,
      };

      const formStore = writable({});
      const errorsStore = writable({});
      const manager = new AddDataFormManager({
        connector,
        formType: "connector",
        formStore: formStore as any,
        errorsStore: errorsStore as any,
        schemaName: "gcs",
      });

      // Set to connector step first
      setStep("connector");
      expect(get(connectorStepStore).step).toBe("connector");

      // Call handleSkip
      manager.handleSkip();

      // Should advance to source step
      expect(get(connectorStepStore).step).toBe("source");
      expect(manager.isMultiStepConnector).toBe(true);
    });

    it("should skip to explorer step for connectors with explorer step", () => {
      const connector: V1ConnectorDriver = {
        name: "snowflake",
        displayName: "Snowflake",
        implementsObjectStore: false,
        implementsOlap: false,
        implementsSqlStore: false,
        implementsWarehouse: true,
        implementsFileStore: false,
      };

      const formStore = writable({});
      const errorsStore = writable({});
      const manager = new AddDataFormManager({
        connector,
        formType: "connector",
        formStore: formStore as any,
        errorsStore: errorsStore as any,
        schemaName: "snowflake",
      });

      // Set to connector step first
      setStep("connector");
      expect(get(connectorStepStore).step).toBe("connector");

      // Call handleSkip
      manager.handleSkip();

      // Should advance to explorer step
      expect(get(connectorStepStore).step).toBe("explorer");
      expect(manager.hasExplorerStep).toBe(true);
    });

    it("should not skip if not on connector step", () => {
      const connector: V1ConnectorDriver = {
        name: "gcs",
        displayName: "GCS",
        implementsObjectStore: true,
        implementsOlap: false,
        implementsSqlStore: false,
        implementsWarehouse: false,
        implementsFileStore: false,
      };

      const formStore = writable({});
      const errorsStore = writable({});
      const manager = new AddDataFormManager({
        connector,
        formType: "connector",
        formStore: formStore as any,
        errorsStore: errorsStore as any,
        schemaName: "gcs",
      });

      // Set to source step
      setStep("source");
      expect(get(connectorStepStore).step).toBe("source");

      // Call handleSkip - should not change step
      manager.handleSkip();

      // Step should remain unchanged
      expect(get(connectorStepStore).step).toBe("source");
    });
  });
});
