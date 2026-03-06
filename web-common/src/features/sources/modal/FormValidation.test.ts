import { describe, it, expect, beforeAll } from "vitest";

import { getValidationSchemaForConnector } from "./FormValidation";
import { populateSchemaCache } from "./connector-schemas";
import type { MultiStepFormSchema } from "../../templates/schemas/types";

// Load the real S3 schema from the backend template definition.
// This ensures frontend validation stays in sync with the backend.
import s3Template from "../../../../../runtime/templates/definitions/duckdb-models/s3-duckdb.json";

beforeAll(() => {
  populateSchemaCache({
    s3: {
      ...s3Template.json_schema,
      title: s3Template.display_name,
    } as unknown as MultiStepFormSchema,
  });
});

describe("getValidationSchemaForConnector (multi-step auth)", () => {
  it("enforces required fields for access key auth", async () => {
    const schema = getValidationSchemaForConnector("s3", "connector");

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
    const schema = getValidationSchemaForConnector("s3", "connector");

    const result = await schema.validate({ auth_method: "public" });
    expect(result.success).toBe(true);
  });

  it("requires source fields from JSON schema for multi-step connectors", async () => {
    const schema = getValidationSchemaForConnector("s3", "source");

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
    const schema = getValidationSchemaForConnector("s3", "source");

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
    const schema = getValidationSchemaForConnector("s3", "source");

    const result = await schema.validate({
      path: "s3://bucket/prefix",
      name: "valid_name",
    });
    expect(result.success).toBe(true);
  });
});
