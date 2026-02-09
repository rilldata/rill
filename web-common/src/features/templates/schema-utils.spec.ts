import { describe, expect, it } from "vitest";
import {
  isDisabledForValues,
  isVisibleForValues,
  getRequiredFieldsForValues,
  isEnumWithDisplay,
  isRadioEnum,
  isTabsEnum,
  isSelectEnum,
  isRichSelectEnum,
  buildEnumOptions,
  radioOptions,
  tabOptions,
  selectOptions,
} from "./schema-utils";
import type { MultiStepFormSchema, JSONSchemaField } from "./schemas/types";

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

  describe("enum display type checks", () => {
    const radioField: JSONSchemaField = {
      type: "string",
      enum: ["a", "b"],
      "x-display": "radio",
    };

    const tabsField: JSONSchemaField = {
      type: "string",
      enum: ["x", "y"],
      "x-display": "tabs",
    };

    const selectField: JSONSchemaField = {
      type: "string",
      enum: ["1", "2"],
      "x-display": "select",
    };

    const plainField: JSONSchemaField = {
      type: "string",
    };

    const enumNoDisplay: JSONSchemaField = {
      type: "string",
      enum: ["a", "b"],
    };

    describe("isEnumWithDisplay", () => {
      it("returns true when enum and display type match", () => {
        expect(isEnumWithDisplay(radioField, "radio")).toBe(true);
        expect(isEnumWithDisplay(tabsField, "tabs")).toBe(true);
        expect(isEnumWithDisplay(selectField, "select")).toBe(true);
      });

      it("returns false when display type does not match", () => {
        expect(isEnumWithDisplay(radioField, "tabs")).toBe(false);
        expect(isEnumWithDisplay(tabsField, "select")).toBe(false);
      });

      it("returns false when field has no enum", () => {
        expect(isEnumWithDisplay(plainField, "radio")).toBe(false);
      });

      it("returns false when field has enum but no x-display", () => {
        expect(isEnumWithDisplay(enumNoDisplay, "radio")).toBe(false);
      });
    });

    describe("isRadioEnum", () => {
      it("returns true for radio display", () => {
        expect(isRadioEnum(radioField)).toBe(true);
      });

      it("returns false for other display types", () => {
        expect(isRadioEnum(tabsField)).toBe(false);
        expect(isRadioEnum(selectField)).toBe(false);
        expect(isRadioEnum(plainField)).toBe(false);
      });
    });

    describe("isTabsEnum", () => {
      it("returns true for tabs display", () => {
        expect(isTabsEnum(tabsField)).toBe(true);
      });

      it("returns false for other display types", () => {
        expect(isTabsEnum(radioField)).toBe(false);
        expect(isTabsEnum(selectField)).toBe(false);
      });
    });

    describe("isSelectEnum", () => {
      it("returns true for select display", () => {
        expect(isSelectEnum(selectField)).toBe(true);
      });

      it("returns false for other display types", () => {
        expect(isSelectEnum(radioField)).toBe(false);
        expect(isSelectEnum(tabsField)).toBe(false);
      });
    });

    describe("isRichSelectEnum", () => {
      const richSelectField: JSONSchemaField = {
        type: "string",
        enum: ["a", "b"],
        "x-display": "select",
        "x-select-style": "rich",
      };

      const standardSelectField: JSONSchemaField = {
        type: "string",
        enum: ["a", "b"],
        "x-display": "select",
      };

      it("returns true for select with rich style", () => {
        expect(isRichSelectEnum(richSelectField)).toBe(true);
      });

      it("returns false for standard select without rich style", () => {
        expect(isRichSelectEnum(standardSelectField)).toBe(false);
      });

      it("returns false for other display types", () => {
        expect(isRichSelectEnum(radioField)).toBe(false);
        expect(isRichSelectEnum(tabsField)).toBe(false);
      });

      it("returns false for field with no enum", () => {
        expect(isRichSelectEnum(plainField)).toBe(false);
      });
    });
  });

  describe("buildEnumOptions", () => {
    const fieldWithLabels: JSONSchemaField = {
      type: "string",
      enum: ["cloud", "self-managed", "rill-managed"],
      "x-enum-labels": ["Cloud", "Self Managed", "Rill Managed"],
      "x-enum-descriptions": ["Cloud desc", "Self desc", "Rill desc"],
      "x-enum-icons": ["cloud-icon", "server-icon", "sparkles-icon"],
    };

    const fieldWithoutLabels: JSONSchemaField = {
      type: "string",
      enum: ["a", "b", "c"],
    };

    it("builds options with labels from x-enum-labels", () => {
      const options = buildEnumOptions(fieldWithLabels);
      expect(options).toHaveLength(3);
      expect(options[0]).toEqual({ value: "cloud", label: "Cloud" });
      expect(options[1]).toEqual({
        value: "self-managed",
        label: "Self Managed",
      });
    });

    it("falls back to enum value when no label provided", () => {
      const options = buildEnumOptions(fieldWithoutLabels);
      expect(options[0]).toEqual({ value: "a", label: "a" });
    });

    it("includes descriptions when includeDescription is true", () => {
      const options = buildEnumOptions(fieldWithLabels, {
        includeDescription: true,
      });
      expect(options[0].description).toBe("Cloud desc");
      expect(options[1].description).toBe("Self desc");
    });

    it("excludes descriptions by default", () => {
      const options = buildEnumOptions(fieldWithLabels);
      expect(options[0].description).toBeUndefined();
    });

    it("includes icons when includeIcons is true and iconMap provided", () => {
      const mockIcon = {} as any;
      const iconMap = { "cloud-icon": mockIcon };
      const options = buildEnumOptions(fieldWithLabels, {
        includeIcons: true,
        iconMap,
      });
      expect(options[0].icon).toBe(mockIcon);
      expect(options[1].icon).toBeUndefined(); // server-icon not in map
    });

    it("returns empty array when field has no enum", () => {
      const options = buildEnumOptions({ type: "string" });
      expect(options).toEqual([]);
    });
  });

  describe("radioOptions", () => {
    it("includes descriptions", () => {
      const field: JSONSchemaField = {
        type: "string",
        enum: ["a", "b"],
        "x-enum-descriptions": ["Desc A", "Desc B"],
      };
      const options = radioOptions(field);
      expect(options[0].description).toBe("Desc A");
    });
  });

  describe("tabOptions", () => {
    it("excludes descriptions", () => {
      const field: JSONSchemaField = {
        type: "string",
        enum: ["a", "b"],
        "x-enum-descriptions": ["Desc A", "Desc B"],
      };
      const options = tabOptions(field);
      expect(options[0].description).toBeUndefined();
    });
  });

  describe("selectOptions", () => {
    it("includes descriptions and icons", () => {
      const mockIcon = {} as any;
      const field: JSONSchemaField = {
        type: "string",
        enum: ["a"],
        "x-enum-descriptions": ["Desc A"],
        "x-enum-icons": ["icon-a"],
      };
      const options = selectOptions(field, { "icon-a": mockIcon });
      expect(options[0].description).toBe("Desc A");
      expect(options[0].icon).toBe(mockIcon);
    });
  });
});
