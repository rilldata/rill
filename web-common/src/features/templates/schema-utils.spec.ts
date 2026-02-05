import { describe, expect, it } from "vitest";
import {
  isDisabledForValues,
  isVisibleForValues,
  getRequiredFieldsForValues,
} from "./schema-utils";
import type { MultiStepFormSchema } from "./schemas/types";

describe("schema-utils", () => {
  describe("isDisabledForValues", () => {
    const schema: MultiStepFormSchema = {
      type: "object",
      properties: {
        mode: {
          type: "string",
          enum: ["auto", "manual"],
        },
        auto_field: {
          type: "string",
          "x-disabled-if": { mode: "manual" },
        },
        manual_field: {
          type: "string",
          "x-disabled-if": { mode: "auto" },
        },
        multi_condition_field: {
          type: "string",
          "x-disabled-if": { mode: ["auto", "disabled"] },
        },
        no_condition_field: {
          type: "string",
        },
      },
    };

    it("returns false when field has no x-disabled-if condition", () => {
      expect(
        isDisabledForValues(schema, "no_condition_field", { mode: "auto" }),
      ).toBe(false);
      expect(isDisabledForValues(schema, "mode", { mode: "auto" })).toBe(false);
    });

    it("returns false when field does not exist", () => {
      expect(isDisabledForValues(schema, "nonexistent", { mode: "auto" })).toBe(
        false,
      );
    });

    it("returns true when single condition matches", () => {
      expect(
        isDisabledForValues(schema, "auto_field", { mode: "manual" }),
      ).toBe(true);
      expect(
        isDisabledForValues(schema, "manual_field", { mode: "auto" }),
      ).toBe(true);
    });

    it("returns false when single condition does not match", () => {
      expect(isDisabledForValues(schema, "auto_field", { mode: "auto" })).toBe(
        false,
      );
      expect(
        isDisabledForValues(schema, "manual_field", { mode: "manual" }),
      ).toBe(false);
    });

    it("returns true when value matches any in array condition", () => {
      expect(
        isDisabledForValues(schema, "multi_condition_field", { mode: "auto" }),
      ).toBe(true);
      expect(
        isDisabledForValues(schema, "multi_condition_field", {
          mode: "disabled",
        }),
      ).toBe(true);
    });

    it("returns false when value does not match any in array condition", () => {
      expect(
        isDisabledForValues(schema, "multi_condition_field", {
          mode: "manual",
        }),
      ).toBe(false);
    });

    it("handles undefined values gracefully", () => {
      expect(isDisabledForValues(schema, "auto_field", {})).toBe(false);
      expect(
        isDisabledForValues(schema, "auto_field", { mode: undefined }),
      ).toBe(false);
    });

    it("converts values to strings for comparison", () => {
      const numericSchema: MultiStepFormSchema = {
        type: "object",
        properties: {
          count: { type: "number" },
          disabled_at_zero: {
            type: "string",
            "x-disabled-if": { count: "0" },
          },
        },
      };
      expect(
        isDisabledForValues(numericSchema, "disabled_at_zero", { count: 0 }),
      ).toBe(true);
      expect(
        isDisabledForValues(numericSchema, "disabled_at_zero", { count: 1 }),
      ).toBe(false);
    });
  });

  describe("isVisibleForValues", () => {
    const schema: MultiStepFormSchema = {
      type: "object",
      properties: {
        type: {
          type: "string",
          enum: ["cloud", "self-hosted"],
        },
        cloud_field: {
          type: "string",
          "x-visible-if": { type: "cloud" },
        },
        multi_type_field: {
          type: "string",
          "x-visible-if": { type: ["cloud", "hybrid"] },
        },
        always_visible: {
          type: "string",
        },
      },
    };

    it("returns true when field has no visibility condition", () => {
      expect(
        isVisibleForValues(schema, "always_visible", { type: "cloud" }),
      ).toBe(true);
    });

    it("returns true when single condition matches", () => {
      expect(isVisibleForValues(schema, "cloud_field", { type: "cloud" })).toBe(
        true,
      );
    });

    it("returns false when single condition does not match", () => {
      expect(
        isVisibleForValues(schema, "cloud_field", { type: "self-hosted" }),
      ).toBe(false);
    });

    it("returns true when value matches any in array condition", () => {
      expect(
        isVisibleForValues(schema, "multi_type_field", { type: "cloud" }),
      ).toBe(true);
      expect(
        isVisibleForValues(schema, "multi_type_field", { type: "hybrid" }),
      ).toBe(true);
    });

    it("returns false when value does not match array condition", () => {
      expect(
        isVisibleForValues(schema, "multi_type_field", { type: "self-hosted" }),
      ).toBe(false);
    });
  });

  describe("getRequiredFieldsForValues", () => {
    const schema: MultiStepFormSchema = {
      type: "object",
      properties: {
        mode: {
          type: "string",
          enum: ["cloud", "manual"],
          "x-step": "connector",
        },
        host: {
          type: "string",
          "x-step": "connector",
        },
        dsn: {
          type: "string",
          "x-step": "connector",
        },
        name: {
          type: "string",
          "x-step": "source",
        },
      },
      required: ["mode"],
      allOf: [
        {
          if: { properties: { mode: { const: "cloud" } } },
          then: { required: ["host"] },
        },
        {
          if: { properties: { mode: { const: "manual" } } },
          then: { required: ["dsn"] },
        },
      ],
    };

    it("includes base required fields", () => {
      const required = getRequiredFieldsForValues(schema, { mode: "cloud" });
      expect(required.has("mode")).toBe(true);
    });

    it("includes conditional required fields when condition matches", () => {
      const cloudRequired = getRequiredFieldsForValues(schema, {
        mode: "cloud",
      });
      expect(cloudRequired.has("host")).toBe(true);
      expect(cloudRequired.has("dsn")).toBe(false);

      const manualRequired = getRequiredFieldsForValues(schema, {
        mode: "manual",
      });
      expect(manualRequired.has("dsn")).toBe(true);
      expect(manualRequired.has("host")).toBe(false);
    });

    it("filters by step when provided", () => {
      const connectorRequired = getRequiredFieldsForValues(
        schema,
        { mode: "cloud" },
        "connector",
      );
      expect(connectorRequired.has("mode")).toBe(true);
      expect(connectorRequired.has("host")).toBe(true);

      // 'name' has x-step: "source", so shouldn't be included in connector step
      const sourceSchema: MultiStepFormSchema = {
        ...schema,
        required: ["mode", "name"],
      };
      const sourceRequired = getRequiredFieldsForValues(
        sourceSchema,
        { mode: "cloud" },
        "connector",
      );
      expect(sourceRequired.has("name")).toBe(false);
    });
  });
});
