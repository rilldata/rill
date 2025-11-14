import { humanReadableErrorMessage } from "../errors/errors";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import type { ClickHouseConnectorType } from "./constants";

/**
 * Returns true for undefined, null, empty string, or whitespace-only string.
 * Useful for validating optional text inputs.
 */
export function isEmpty(val: any) {
  return (
    val === undefined ||
    val === null ||
    val === "" ||
    (typeof val === "string" && val.trim() === "")
  );
}

/**
 * Normalizes a variety of error shapes into a string, string[], or undefined.
 * - If input is an array, returns it as-is.
 * - If input is a string, returns it.
 * - If input resembles a Zod `_errors` array, returns that.
 * - Otherwise returns undefined.
 */
export function normalizeErrors(
  err: any,
): string | string[] | null | undefined {
  if (!err) return undefined;
  if (Array.isArray(err)) return err;
  if (typeof err === "string") return err;
  if (err._errors && Array.isArray(err._errors)) return err._errors;
  return undefined;
}

/**
 * Converts unknown error inputs into a unified connector error shape.
 * - Prefers native Error.message when present
 * - Maps server error responses to human-readable messages via `humanReadableErrorMessage`
 * - Returns `details` with original message when it differs from the human-readable message
 */
export function normalizeConnectorError(
  connectorName: string,
  err: any,
): { message: string; details?: string } {
  let message: string;
  let details: string | undefined = undefined;

  if (err instanceof Error) {
    message = err.message;
  } else if (err?.message && err?.details) {
    message = err.message;
    details = err.details !== err.message ? err.details : undefined;
  } else if (err?.response?.data) {
    const originalMessage = err.response.data.message;
    const humanReadable = humanReadableErrorMessage(
      connectorName,
      err.response.data.code,
      originalMessage,
    );
    message = humanReadable;
    details = humanReadable !== originalMessage ? originalMessage : undefined;
  } else if (err?.message) {
    message = err.message;
  } else {
    message = "Unknown error";
  }

  return { message, details };
}

/**
 * Indicates whether a connector in "connector" mode exposes only a DSN field
 * (i.e., DSN exists and no other config properties are present).
 */
export function hasOnlyDsn(
  connector: V1ConnectorDriver | undefined,
  isConnectorForm: boolean,
): boolean {
  if (!isConnectorForm) return false;
  const props = connector?.configProperties ?? [];
  const hasDsn = props.some((p) => p.key === "dsn");
  const hasOthers = props.some((p) => p.key !== "dsn");
  return hasDsn && !hasOthers;
}

/**
 * Applies ClickHouse Cloud-specific default requirements for connector values.
 * - For ClickHouse Cloud: enforces `ssl: true` and `port: "8443"`
 * - Otherwise returns values unchanged
 */
export function applyClickHouseCloudRequirements(
  connectorName: string | undefined,
  connectorType: ClickHouseConnectorType,
  values: Record<string, unknown>,
): Record<string, unknown> {
  if (connectorName === "clickhouse" && connectorType === "clickhouse-cloud") {
    return { ...values, ssl: true, port: "8443" } as Record<string, unknown>;
  }
  return values;
}
