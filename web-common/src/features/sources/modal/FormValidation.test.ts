import { describe, it, expect } from "vitest";

import { getValidationSchemaForConnector } from "./FormValidation";

describe("getValidationSchemaForConnector (multi-step auth)", () => {
  it("enforces required fields for access key auth", async () => {
    const schema = getValidationSchemaForConnector("s3", "connector", {
      isMultiStepConnector: true,
    });

    const result = await schema.validate({});
    expect(result.success).toBe(false);
    if (result.success) throw new Error("expected validation to fail");
    expect(result.issues).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ path: ["aws_access_key_id"] }),
        expect.objectContaining({ path: ["aws_secret_access_key"] }),
      ]),
    );
  });

  it("allows public auth without credentials", async () => {
    const schema = getValidationSchemaForConnector("s3", "connector", {
      isMultiStepConnector: true,
    });

    const result = await schema.validate({ auth_method: "public" });
    expect(result.success).toBe(true);
  });

  it("requires source fields from JSON schema for multi-step connectors", async () => {
    const schema = getValidationSchemaForConnector("s3", "source", {
      isMultiStepConnector: true,
    });

    const result = await schema.validate({});
    expect(result.success).toBe(false);
    if (result.success) throw new Error("expected validation to fail");
    expect(result.issues).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ path: ["path"] }),
        expect.objectContaining({ path: ["name"] }),
      ]),
    );
  });

  it("rejects invalid s3 path on source step", async () => {
    const schema = getValidationSchemaForConnector("s3", "source", {
      isMultiStepConnector: true,
    });

    const result = await schema.validate({
      path: "s3:/bucket",
      name: "valid_name",
    });
    expect(result.success).toBe(false);
    if (result.success) throw new Error("expected validation to fail");
    expect(result.issues).toEqual(
      expect.arrayContaining([expect.objectContaining({ path: ["path"] })]),
    );
  });

  it("accepts valid s3 path on source step", async () => {
    const schema = getValidationSchemaForConnector("s3", "source", {
      isMultiStepConnector: true,
    });

    const result = await schema.validate({
      path: "s3://bucket/prefix",
      name: "valid_name",
    });
    expect(result.success).toBe(true);
  });
});
