import { describe, expect, it } from "vitest";
import {
  GENERATED_PASSWORD_LENGTH,
  generatePassword,
} from "./generate-password";

describe("generatePassword", () => {
  it("generates passwords of the requested length from the unambiguous alphabet", () => {
    expect(generatePassword()).toHaveLength(GENERATED_PASSWORD_LENGTH);
    expect(generatePassword(32)).toHaveLength(32);
    for (let i = 0; i < 50; i++) {
      expect(generatePassword()).toMatch(/^[A-HJ-NP-Za-km-z2-9]{16}$/);
    }
  });

  it("generates different passwords on each call", () => {
    const passwords = new Set(
      Array.from({ length: 20 }, () => generatePassword()),
    );
    expect(passwords.size).toBe(20);
  });
});
