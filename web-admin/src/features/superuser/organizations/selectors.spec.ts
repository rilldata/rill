import { describe, expect, it } from "vitest";
import { pickAssumableMember } from "./selectors";

describe("pickAssumableMember", () => {
  it("returns undefined for an empty list", () => {
    expect(pickAssumableMember(undefined)).toBeUndefined();
    expect(pickAssumableMember([])).toBeUndefined();
  });

  it("prefers a member with the admin role", () => {
    const result = pickAssumableMember([
      { userEmail: "viewer@example.com", roleName: "viewer" },
      { userEmail: "admin@example.com", roleName: "admin" },
      { userEmail: "editor@example.com", roleName: "editor" },
    ]);
    expect(result).toEqual({ userEmail: "admin@example.com" });
  });

  it("falls back to the first member when no admin exists", () => {
    const result = pickAssumableMember([
      { userEmail: "viewer@example.com", roleName: "viewer" },
      { userEmail: "editor@example.com", roleName: "editor" },
    ]);
    expect(result).toEqual({ userEmail: "viewer@example.com" });
  });

  it("skips members without a userEmail", () => {
    const result = pickAssumableMember([{ roleName: "admin" }]);
    expect(result).toBeUndefined();
  });
});
