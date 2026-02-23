const slugSanitiserRegex = /[^\w-]/g;

/**
 * Sanitizes a string to be used as a URL slug.
 * Replaces any non-word characters (except hyphens) with hyphens.
 */
export function sanitizeSlug(name: string): string {
  return name.replace(slugSanitiserRegex, "-");
}
