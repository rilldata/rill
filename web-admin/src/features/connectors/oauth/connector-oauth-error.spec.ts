import { Code, ConnectError } from "@connectrpc/connect";
import { describe, expect, it } from "vitest";
import { connectorOAuthRequired } from "./connector-oauth-error";

describe("connectorOAuthRequired", () => {
  const message = 'sign in to connector "dbx" to view this data';

  it("returns the connector for a FailedPrecondition sign-in error", () => {
    expect(connectorOAuthRequired({ code: 9, message })).toBe("dbx");
  });

  it("matches a ConnectError, whose message is prefixed with the code", () => {
    expect(
      connectorOAuthRequired(
        new ConnectError(message, Code.FailedPrecondition),
      ),
    ).toBe("dbx");
  });

  it("returns null for other FailedPrecondition errors", () => {
    expect(connectorOAuthRequired({ code: 9, message: "other" })).toBeNull();
  });

  it("returns null for other codes", () => {
    expect(connectorOAuthRequired({ code: 5, message })).toBeNull();
  });

  it("returns null for undefined", () => {
    expect(connectorOAuthRequired(undefined)).toBeNull();
  });
});
