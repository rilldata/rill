import { describe, it, expect, vi, beforeEach } from "vitest";
import { mergeEnvVars } from "./generate-template";

// Mock the runtime store
vi.mock("../../../runtime-client/runtime-store", () => ({
  runtime: {
    subscribe: vi.fn((cb) => {
      cb({ instanceId: "test-instance" });
      return () => {};
    }),
  },
}));

// Mock the runtime client
const mockGetFile = vi.fn();
vi.mock("../../../runtime-client", () => ({
  getRuntimeServiceGetFileQueryKey: vi.fn(
    (instanceId: string, params: { path: string }) => [
      "runtimeServiceGetFile",
      instanceId,
      params,
    ],
  ),
  runtimeServiceGetFile: (...args: unknown[]) => mockGetFile(...args),
}));

// Mock replaceOrAddEnvVariable (use actual implementation)
vi.mock("../../connectors/code-utils", async () => {
  return {
    replaceOrAddEnvVariable: (
      existingEnvBlob: string,
      key: string,
      newValue: string,
    ): string => {
      const lines = existingEnvBlob.split("\n");
      let keyFound = false;

      const updatedLines = lines.map((line) => {
        if (line.startsWith(`${key}=`)) {
          keyFound = true;
          return `${key}=${newValue}`;
        }
        return line;
      });

      if (!keyFound) {
        updatedLines.push(`${key}=${newValue}`);
      }

      const newBlob = updatedLines
        .filter((line, index) => !(line === "" && index === 0))
        .join("\n")
        .trim();

      return newBlob;
    },
  };
});

describe("mergeEnvVars", () => {
  let queryClient: any;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = {
      invalidateQueries: vi.fn().mockResolvedValue(undefined),
      fetchQuery: vi.fn(),
    };
  });

  it("should merge env vars into existing .env content", async () => {
    queryClient.fetchQuery.mockResolvedValue({
      blob: "EXISTING_VAR=existing_value",
    });

    const result = await mergeEnvVars(queryClient, {
      CLICKHOUSE_PASSWORD: "secret123",
      CLICKHOUSE_HOST: "ch.example.com",
    });

    expect(result.originalBlob).toBe("EXISTING_VAR=existing_value");
    expect(result.newBlob).toContain("EXISTING_VAR=existing_value");
    expect(result.newBlob).toContain("CLICKHOUSE_PASSWORD=secret123");
    expect(result.newBlob).toContain("CLICKHOUSE_HOST=ch.example.com");
  });

  it("should handle empty .env file", async () => {
    queryClient.fetchQuery.mockResolvedValue({ blob: "" });

    const result = await mergeEnvVars(queryClient, {
      S3_ACCESS_KEY: "AKID123",
    });

    expect(result.originalBlob).toBe("");
    expect(result.newBlob).toContain("S3_ACCESS_KEY=AKID123");
  });

  it("should handle .env file not found", async () => {
    queryClient.fetchQuery.mockRejectedValue({
      response: { data: { message: "no such file" } },
    });

    const result = await mergeEnvVars(queryClient, {
      NEW_VAR: "new_value",
    });

    expect(result.originalBlob).toBe("");
    expect(result.newBlob).toContain("NEW_VAR=new_value");
  });

  it("should update existing env var values", async () => {
    queryClient.fetchQuery.mockResolvedValue({
      blob: "CLICKHOUSE_PASSWORD=old_secret",
    });

    const result = await mergeEnvVars(queryClient, {
      CLICKHOUSE_PASSWORD: "new_secret",
    });

    expect(result.newBlob).toContain("CLICKHOUSE_PASSWORD=new_secret");
    expect(result.newBlob).not.toContain("old_secret");
  });

  it("should handle empty envVars map", async () => {
    queryClient.fetchQuery.mockResolvedValue({
      blob: "EXISTING=value",
    });

    const result = await mergeEnvVars(queryClient, {});

    expect(result.originalBlob).toBe("EXISTING=value");
    expect(result.newBlob).toBe("EXISTING=value");
  });

  it("should skip entries with empty keys or values", async () => {
    queryClient.fetchQuery.mockResolvedValue({ blob: "" });

    const result = await mergeEnvVars(queryClient, {
      "": "no_key",
      VALID_KEY: "",
      REAL_KEY: "real_value",
    });

    expect(result.newBlob).toContain("REAL_KEY=real_value");
    expect(result.newBlob).not.toContain("no_key");
    expect(result.newBlob).not.toContain("VALID_KEY");
  });

  it("should invalidate query cache before fetching", async () => {
    queryClient.fetchQuery.mockResolvedValue({ blob: "" });

    await mergeEnvVars(queryClient, { KEY: "value" });

    expect(queryClient.invalidateQueries).toHaveBeenCalledBefore(
      queryClient.fetchQuery,
    );
  });

  it("should re-throw non-file-not-found errors", async () => {
    const error = new Error("network error");
    queryClient.fetchQuery.mockRejectedValue(error);

    await expect(mergeEnvVars(queryClient, { KEY: "value" })).rejects.toThrow(
      "network error",
    );
  });

  it("should handle suffixed env var names from backend", async () => {
    queryClient.fetchQuery.mockResolvedValue({
      blob: "CLICKHOUSE_PASSWORD=first_secret",
    });

    // Backend already resolved the conflict and returned _1 suffix
    const result = await mergeEnvVars(queryClient, {
      CLICKHOUSE_PASSWORD_1: "second_secret",
    });

    expect(result.newBlob).toContain("CLICKHOUSE_PASSWORD=first_secret");
    expect(result.newBlob).toContain("CLICKHOUSE_PASSWORD_1=second_secret");
  });
});
