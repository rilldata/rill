import { describe, it, expect } from "vitest";
import {
  capitalize,
  formatOrgRole,
  validateServiceName,
  formatServiceDate,
  formatServiceDateTime,
  NAME_PATTERN,
  NONE_ROLE,
  ORG_ROLES,
  PROJECT_ROLES,
  DEFAULT_PROJECT_ROLE,
} from "./utils";

describe("utils", () => {
  describe("constants", () => {
    it("NONE_ROLE is empty string", () => {
      expect(NONE_ROLE).toBe("");
    });

    it("ORG_ROLES includes expected roles", () => {
      expect(ORG_ROLES).toContain("admin");
      expect(ORG_ROLES).toContain("editor");
      expect(ORG_ROLES).toContain("viewer");
      expect(ORG_ROLES).toContain("guest");
      expect(ORG_ROLES).toContain(NONE_ROLE);
    });

    it("PROJECT_ROLES includes expected roles", () => {
      expect(PROJECT_ROLES).toContain("admin");
      expect(PROJECT_ROLES).toContain("editor");
      expect(PROJECT_ROLES).toContain("viewer");
    });

    it("DEFAULT_PROJECT_ROLE is viewer", () => {
      expect(DEFAULT_PROJECT_ROLE).toBe("viewer");
      expect(PROJECT_ROLES).toContain(DEFAULT_PROJECT_ROLE);
    });
  });

  describe("capitalize", () => {
    it("capitalizes the first letter", () => {
      expect(capitalize("viewer")).toBe("Viewer");
      expect(capitalize("admin")).toBe("Admin");
    });

    it("handles empty string", () => {
      expect(capitalize("")).toBe("");
    });

    it("handles already capitalized string", () => {
      expect(capitalize("Admin")).toBe("Admin");
    });

    it("handles single character", () => {
      expect(capitalize("a")).toBe("A");
    });
  });

  describe("formatOrgRole", () => {
    it("returns 'None' for undefined", () => {
      expect(formatOrgRole(undefined)).toBe("None");
    });

    it("returns 'None' for empty string", () => {
      expect(formatOrgRole("")).toBe("None");
    });

    it("capitalizes known roles", () => {
      expect(formatOrgRole("admin")).toBe("Admin");
      expect(formatOrgRole("editor")).toBe("Editor");
      expect(formatOrgRole("viewer")).toBe("Viewer");
      expect(formatOrgRole("guest")).toBe("Guest");
    });
  });

  describe("NAME_PATTERN", () => {
    it("allows letters", () => {
      expect(NAME_PATTERN.test("myservice")).toBe(true);
    });

    it("allows underscores at start", () => {
      expect(NAME_PATTERN.test("_myservice")).toBe(true);
    });

    it("allows hyphens", () => {
      expect(NAME_PATTERN.test("my-service")).toBe(true);
    });

    it("allows digits after first character", () => {
      expect(NAME_PATTERN.test("service123")).toBe(true);
    });

    it("rejects names starting with digits", () => {
      expect(NAME_PATTERN.test("123service")).toBe(false);
    });

    it("rejects names starting with hyphens", () => {
      expect(NAME_PATTERN.test("-service")).toBe(false);
    });

    it("rejects names with spaces", () => {
      expect(NAME_PATTERN.test("my service")).toBe(false);
    });

    it("rejects names with special characters", () => {
      expect(NAME_PATTERN.test("my.service")).toBe(false);
      expect(NAME_PATTERN.test("my@service")).toBe(false);
    });
  });

  describe("validateServiceName", () => {
    it("returns empty string for valid names", () => {
      expect(validateServiceName("my-service")).toBe("");
      expect(validateServiceName("_internal")).toBe("");
      expect(validateServiceName("Service_123")).toBe("");
    });

    it("returns error for empty/whitespace-only names", () => {
      expect(validateServiceName("")).toBe("Name is required");
      expect(validateServiceName("   ")).toBe("Name is required");
    });

    it("returns error for invalid characters", () => {
      const error = validateServiceName("my service");
      expect(error).toContain("Must start with a letter or underscore");
    });

    it("returns error for names starting with digit", () => {
      expect(validateServiceName("1service")).toContain(
        "Must start with a letter",
      );
    });

    it("returns error for names starting with hyphen", () => {
      expect(validateServiceName("-service")).toContain(
        "Must start with a letter",
      );
    });

    it("trims whitespace before validating", () => {
      expect(validateServiceName("  valid-name  ")).toBe("");
    });
  });

  describe("formatServiceDate", () => {
    it("returns dash for undefined", () => {
      expect(formatServiceDate(undefined)).toBe("-");
    });

    it("returns dash for empty string", () => {
      expect(formatServiceDate("")).toBe("-");
    });

    it("returns dash for invalid date", () => {
      expect(formatServiceDate("not-a-date")).toBe("-");
    });

    it("returns dash for dates before 1970", () => {
      expect(formatServiceDate("1969-01-01T00:00:00Z")).toBe("-");
    });

    it("formats a valid ISO date string", () => {
      const result = formatServiceDate("2024-03-15T10:30:00Z");
      expect(result).not.toBe("-");
      expect(result).toContain("2024");
    });
  });

  describe("formatServiceDateTime", () => {
    it("returns dash for undefined", () => {
      expect(formatServiceDateTime(undefined)).toBe("-");
    });

    it("returns dash for empty string", () => {
      expect(formatServiceDateTime("")).toBe("-");
    });

    it("returns dash for invalid date", () => {
      expect(formatServiceDateTime("invalid")).toBe("-");
    });

    it("returns dash for dates before 1970", () => {
      expect(formatServiceDateTime("1969-06-15T00:00:00Z")).toBe("-");
    });

    it("formats a valid ISO datetime string", () => {
      const result = formatServiceDateTime("2024-03-15T10:30:00Z");
      expect(result).not.toBe("-");
      expect(result).toContain("2024");
    });
  });
});
