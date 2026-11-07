import { describe, expect, it } from "vitest";
import { isValidPublicURLPassword } from "./password-policy";

describe("public URL password policy", () => {
  it.each([
    ["", false],
    ["1234567", false],
    ["12345678", true],
    ["a".repeat(72), true],
    ["a".repeat(73), false],
    ["é".repeat(36), true],
    ["é".repeat(37), false],
    ["😀".repeat(4), false],
    ["😀".repeat(18), true],
  ])("validates %j as %s", (password, valid) => {
    expect(isValidPublicURLPassword(password)).toBe(valid);
  });
});
