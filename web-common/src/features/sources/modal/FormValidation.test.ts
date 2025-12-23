import { describe, it, expect } from "vitest";

import { getValidationSchemaForConnector } from "./FormValidation";

describe("getValidationSchemaForConnector (multi-step auth)", () => {
  it("enforces required fields for access key auth", async () => {
    const schema = getValidationSchemaForConnector("s3", "connector", {
      isMultiStepConnector: true,
      authMethodGetter: () => "access_keys",
    });

    await expect(
      schema.validate({}, { abortEarly: false }),
    ).rejects.toMatchObject({
      inner: expect.arrayContaining([
        expect.objectContaining({ path: "aws_access_key_id" }),
        expect.objectContaining({ path: "aws_secret_access_key" }),
      ]),
    });
  });

  it("allows public auth without credentials", async () => {
    const schema = getValidationSchemaForConnector("s3", "connector", {
      isMultiStepConnector: true,
      authMethodGetter: () => "public",
    });

    await expect(
      schema.validate({ auth_method: "public" }, { abortEarly: false }),
    ).resolves.toMatchObject({});
  });

  it("requires source fields from JSON schema for multi-step connectors", async () => {
    const schema = getValidationSchemaForConnector("s3", "source", {
      isMultiStepConnector: true,
    });

    await expect(
      schema.validate({}, { abortEarly: false }),
    ).rejects.toMatchObject({
      inner: expect.arrayContaining([
        expect.objectContaining({ path: "path" }),
        expect.objectContaining({ path: "name" }),
      ]),
    });
  });
});
