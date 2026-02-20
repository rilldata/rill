import { describe, it, expect } from "vitest";
import { normalizeConnectorError } from "./utils";

describe("normalizeConnectorError", () => {
  it("should handle standard Error instance", () => {
    const error = new Error("connection refused");
    const result = normalizeConnectorError("postgres", error);
    expect(result.message).toBe("connection refused");
    expect(result.details).toBeUndefined();
  });

  it("should handle Error with details (from waitForResourceReconciliation)", () => {
    const error = new Error("Resource configuration failed to reconcile");
    (error as any).details =
      'failed to connect to "host=localhost user=postgres database=postgres": connection refused';

    // submitAddConnectorForm re-throws as a plain object with { message, details }
    const rethrown = {
      message: error.message,
      details: (error as any).details,
    };
    const result = normalizeConnectorError("postgres", rethrown);

    expect(result.message).toBe("Resource configuration failed to reconcile");
    expect(result.details).toBe(
      'failed to connect to "host=localhost user=postgres database=postgres": connection refused',
    );
  });

  it("should handle error with HTTP response data", () => {
    const error = {
      response: {
        data: {
          message: 'unknown property "bad_prop" for driver',
          code: 3, // InvalidArgument
        },
      },
    };
    const result = normalizeConnectorError("postgres", error);

    // Postgres is not in connectorErrorMap, so it gets the generic template
    expect(result.message).toContain("postgres");
    expect(result.details).toBe('unknown property "bad_prop" for driver');
  });

  it("should handle plain object with message only", () => {
    const error = { message: "Unable to establish a connection" };
    const result = normalizeConnectorError("postgres", error);
    expect(result.message).toBe("Unable to establish a connection");
    expect(result.details).toBeUndefined();
  });

  it("should handle plain object with same message and details", () => {
    const error = {
      message: "connection refused",
      details: "connection refused",
    };
    const result = normalizeConnectorError("postgres", error);
    expect(result.message).toBe("connection refused");
    // details should be undefined when same as message
    expect(result.details).toBeUndefined();
  });

  it("should handle unknown error types", () => {
    const result = normalizeConnectorError("postgres", 42);
    expect(result.message).toBe("Unknown error");
    expect(result.details).toBeUndefined();
  });

  it("should handle null error", () => {
    const result = normalizeConnectorError("postgres", null);
    expect(result.message).toBe("Unknown error");
  });

  it("should handle string error", () => {
    // Strings don't match any of the type guards (not Error, not object with response, not object with message)
    const result = normalizeConnectorError("postgres", "some error");
    expect(result.message).toBe("Unknown error");
  });

  it("should use humanReadableErrorMessage for ClickHouse connection errors", () => {
    const error = {
      response: {
        data: {
          message: "connection refused to host:9000",
          code: 2, // Unknown gRPC code
        },
      },
    };
    const result = normalizeConnectorError("clickhouse", error);

    // ClickHouse has custom error mapping for "connection refused"
    expect(result.message).toContain("Could not connect to ClickHouse");
    // Details should have the original message since humanReadable differs
    expect(result.details).toBe("connection refused to host:9000");
  });

  it("should handle GenerateTemplate RPC validation error", () => {
    const error = {
      response: {
        data: {
          message: 'unknown property "connection_mode" for driver',
          code: 3, // InvalidArgument
        },
      },
    };
    const result = normalizeConnectorError("postgres", error);

    expect(result.message).toBeTruthy();
    // The details should include the original error
    expect(result.details).toBe(
      'unknown property "connection_mode" for driver',
    );
  });

  it("should handle reconciliation error that was auto-deleted", () => {
    const error = new Error(
      "Resource configuration failed to reconcile and was automatically deleted. This usually indicates a connection or configuration error.",
    );
    const result = normalizeConnectorError("postgres", error);

    expect(result.message).toBe(
      "Resource configuration failed to reconcile and was automatically deleted. This usually indicates a connection or configuration error.",
    );
    expect(result.details).toBeUndefined();
  });
});
