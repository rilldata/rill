import { describe, expect, it } from "vitest";
import { updateDotEnvBlobWithNewSecret } from "./code-utils";

describe("updateEnvVariables", () => {
  it("should create a new env file", () => {
    const updatedEnvBlob = updateDotEnvBlobWithNewSecret("", "KEY1", "VALUE1");
    expect(updatedEnvBlob).toBe("KEY1=VALUE1");
  });

  const existingEnvBlob = `# This is a comment
# This is another comment
KEY1=VALUE1
KEY2=VALUE2`;

  it("should update the env file", () => {
    const updatedEnvBlob = updateDotEnvBlobWithNewSecret(
      existingEnvBlob,
      "KEY1",
      "NEW_VALUE1",
    );
    expect(updatedEnvBlob).toBe(`# This is a comment
# This is another comment
KEY1=NEW_VALUE1
KEY2=VALUE2`);
  });

  it("should add a new key to the env file", () => {
    const updatedEnvBlob = updateDotEnvBlobWithNewSecret(
      existingEnvBlob,
      "KEY3",
      "VALUE3",
    );
    expect(updatedEnvBlob).toBe(`# This is a comment
# This is another comment
KEY1=VALUE1
KEY2=VALUE2
KEY3=VALUE3`);
  });
});
