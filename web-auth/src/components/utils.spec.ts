import { describe, it, expect } from "vitest";
import { getConnectionFromEmail, validateEmail } from "./utils";

describe("getConnectionFromEmail", () => {
  const connectionMapping = {
    "connection-a": ["domain1.com", "domain2.com"],
    "connection-b": ["domain3.com", "domain4.com"],
    "connection-c": ["domain5.com"],
  };

  it("should return the connection name for a matching email domain", () => {
    const email = "example@domain2.com";
    const connectionName = getConnectionFromEmail(email, connectionMapping);

    expect(connectionName).toBe("connection-a");
  });

  it("should return undefined for a non-matching email domain", () => {
    const email = "example@domain6.com";
    const connectionName = getConnectionFromEmail(email, connectionMapping);

    expect(connectionName).toBeUndefined();
  });

  it("should return undefined for an empty email address", () => {
    const email = "";
    const connectionName = getConnectionFromEmail(email, connectionMapping);

    expect(connectionName).toBeUndefined();
  });
});

describe("validateEmail", () => {
  it("should return true for a valid email address", () => {
    const validEmail = "example@example.com";
    const isValid = validateEmail(validEmail);

    expect(isValid).toBe(true);
  });

  it("should return false for an invalid email address", () => {
    const invalidEmail = "example.com";
    const isValid = validateEmail(invalidEmail);

    expect(isValid).toBe(false);
  });

  it("should return false for an empty email address", () => {
    const emptyEmail = "";
    const isValid = validateEmail(emptyEmail);

    expect(isValid).toBe(false);
  });
});
