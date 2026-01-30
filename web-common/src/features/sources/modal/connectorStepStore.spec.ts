import { describe, it, expect, beforeEach } from "vitest";
import { get } from "svelte/store";
import {
  connectorStepStore,
  setStep,
  setConnectorConfig,
  setAuthMethod,
  setConnectorInstanceName,
  resetConnectorStep,
} from "./connectorStepStore";

describe("connectorStepStore", () => {
  beforeEach(() => {
    resetConnectorStep();
  });

  it("should have correct initial state", () => {
    const state = get(connectorStepStore);
    expect(state).toEqual({
      step: "connector",
      connectorConfig: null,
      selectedAuthMethod: null,
      connectorInstanceName: null,
    });
  });

  it("should update step correctly", () => {
    setStep("source");
    const state = get(connectorStepStore);
    expect(state.step).toBe("source");
  });

  it("should update connector config correctly", () => {
    const config = { key: "value", secret: "secret123" };
    setConnectorConfig(config);
    const state = get(connectorStepStore);
    expect(state.connectorConfig).toEqual(config);
  });

  it("should update auth method correctly", () => {
    setAuthMethod("public");
    const state = get(connectorStepStore);
    expect(state.selectedAuthMethod).toBe("public");
  });

  it("should update connector instance name correctly", () => {
    setConnectorInstanceName("gcs_1");
    const state = get(connectorStepStore);
    expect(state.connectorInstanceName).toBe("gcs_1");
  });

  it("should preserve other state when updating individual fields", () => {
    setStep("source");
    setConnectorInstanceName("s3_2");
    setAuthMethod("hmac");

    const state = get(connectorStepStore);
    expect(state).toEqual({
      step: "source",
      connectorConfig: null,
      selectedAuthMethod: "hmac",
      connectorInstanceName: "s3_2",
    });
  });

  it("should reset all state correctly", () => {
    // Set various state
    setStep("source");
    setConnectorConfig({ key: "value" });
    setAuthMethod("public");
    setConnectorInstanceName("gcs_3");

    // Reset
    resetConnectorStep();

    const state = get(connectorStepStore);
    expect(state).toEqual({
      step: "connector",
      connectorConfig: null,
      selectedAuthMethod: null,
      connectorInstanceName: null,
    });
  });

  it("should allow setting connector instance name to null", () => {
    setConnectorInstanceName("gcs_1");
    expect(get(connectorStepStore).connectorInstanceName).toBe("gcs_1");

    setConnectorInstanceName(null);
    expect(get(connectorStepStore).connectorInstanceName).toBeNull();
  });
});
