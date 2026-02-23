import { sanitizeSlug } from "@rilldata/web-common/lib/string-utils";

/**
 * Sanitizes an organization name for use in URLs.
 * @deprecated Use sanitizeSlug from @rilldata/web-common/lib/string-utils instead
 */
export function sanitizeOrgName(name: string): string {
  return sanitizeSlug(name);
}
