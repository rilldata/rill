import { describe, it, expect } from "vitest";
import { processFileContent, getFileAccept } from "./file-encoding";
import type { JSONSchemaField } from "./schemas/types";

describe("processFileContent", () => {
  describe("base64 encoding", () => {
    it("encodes ASCII content to base64", () => {
      const field: JSONSchemaField = {
        type: "string",
        "x-file-encoding": "base64",
      };
      const result = processFileContent("hello world", field);
      expect(result.encodedContent).toBe(btoa("hello world"));
      expect(result.extractedValues).toEqual({});
    });

    it("encodes PEM file content to base64", () => {
      const pem =
        "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBg==\n-----END PRIVATE KEY-----";
      const field: JSONSchemaField = {
        type: "string",
        "x-file-encoding": "base64",
      };
      const result = processFileContent(pem, field);
      expect(result.encodedContent).toBe(btoa(pem));
    });

    it("throws on non-Latin-1 characters", () => {
      const field: JSONSchemaField = {
        type: "string",
        "x-file-encoding": "base64",
      };
      expect(() => processFileContent("hello \u{1F600}", field)).toThrow(
        "Invalid file encoding: contains non-Latin-1 characters",
      );
    });
  });

  describe("json encoding", () => {
    it("parses and re-serializes valid JSON", () => {
      const field: JSONSchemaField = {
        type: "string",
        "x-file-encoding": "json",
      };
      const input = '{"project_id":"my-project","type":"service_account"}';
      const result = processFileContent(input, field);
      expect(result.encodedContent).toBe(input);
      expect(result.extractedValues).toEqual({});
    });

    it("normalizes JSON whitespace", () => {
      const field: JSONSchemaField = {
        type: "string",
        "x-file-encoding": "json",
      };
      const result = processFileContent('{ "a" : 1 }', field);
      expect(result.encodedContent).toBe('{"a":1}');
    });

    it("throws on invalid JSON", () => {
      const field: JSONSchemaField = {
        type: "string",
        "x-file-encoding": "json",
      };
      expect(() => processFileContent("not json", field)).toThrow(
        "Invalid JSON file",
      );
    });

    it("extracts values with x-file-extract", () => {
      const field: JSONSchemaField = {
        type: "string",
        "x-file-encoding": "json",
        "x-file-extract": { project_id: "project_id" },
      };
      const input = JSON.stringify({
        project_id: "my-project",
        type: "service_account",
      });
      const result = processFileContent(input, field);
      expect(result.extractedValues).toEqual({ project_id: "my-project" });
    });

    it("skips missing extract keys without error", () => {
      const field: JSONSchemaField = {
        type: "string",
        "x-file-encoding": "json",
        "x-file-extract": { project_id: "project_id" },
      };
      const input = JSON.stringify({ type: "service_account" });
      const result = processFileContent(input, field);
      expect(result.extractedValues).toEqual({});
    });

    it("maps extract keys to different form field names", () => {
      const field: JSONSchemaField = {
        type: "string",
        "x-file-encoding": "json",
        "x-file-extract": { myFormField: "source_key" },
      };
      const input = JSON.stringify({ source_key: "value123" });
      const result = processFileContent(input, field);
      expect(result.extractedValues).toEqual({ myFormField: "value123" });
    });
  });

  describe("raw encoding", () => {
    it("passes content through unchanged", () => {
      const field: JSONSchemaField = {
        type: "string",
        "x-file-encoding": "raw",
      };
      const result = processFileContent("raw content here", field);
      expect(result.encodedContent).toBe("raw content here");
      expect(result.extractedValues).toEqual({});
    });

    it("defaults to raw when no encoding specified", () => {
      const field: JSONSchemaField = { type: "string" };
      const result = processFileContent("some content", field);
      expect(result.encodedContent).toBe("some content");
      expect(result.extractedValues).toEqual({});
    });
  });

  describe("extract is ignored for non-json encodings", () => {
    it("ignores x-file-extract with base64 encoding", () => {
      const field: JSONSchemaField = {
        type: "string",
        "x-file-encoding": "base64",
        "x-file-extract": { project_id: "project_id" },
      };
      const result = processFileContent("hello", field);
      expect(result.extractedValues).toEqual({});
    });

    it("ignores x-file-extract with raw encoding", () => {
      const field: JSONSchemaField = {
        type: "string",
        "x-file-encoding": "raw",
        "x-file-extract": { project_id: "project_id" },
      };
      const result = processFileContent("hello", field);
      expect(result.extractedValues).toEqual({});
    });
  });
});

describe("getFileAccept", () => {
  it("returns x-file-accept when present", () => {
    const field: JSONSchemaField = {
      type: "string",
      "x-file-accept": ".pem,.p8",
    };
    expect(getFileAccept(field)).toBe(".pem,.p8");
  });

  it("falls back to x-accept", () => {
    const field: JSONSchemaField = {
      type: "string",
      "x-accept": ".json",
    };
    expect(getFileAccept(field)).toBe(".json");
  });

  it("prefers x-file-accept over x-accept", () => {
    const field: JSONSchemaField = {
      type: "string",
      "x-file-accept": ".pem",
      "x-accept": ".json",
    };
    expect(getFileAccept(field)).toBe(".pem");
  });

  it("returns undefined when neither is set", () => {
    const field: JSONSchemaField = { type: "string" };
    expect(getFileAccept(field)).toBeUndefined();
  });
});
