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
