import type { V1Project } from "@rilldata/web-admin/client";

// The limits match the admin service's (see `NormalizeProjectTags` in admin/database/validate.go).
export const MAX_PROJECT_TAGS = 10;
const MAX_PROJECT_TAG_LENGTH = 40;

// Letters and digits in any script, combining marks, hyphens and underscores, starting with a letter or digit.
const PROJECT_TAG_REGEX = /^[\p{L}\p{N}][\p{L}\p{M}\p{N}_-]*$/u;

/**
 * Turns text typed as a tag into the form the admin service stores:
 * trimmed, lowercase and NFC-normalized.
 * If the admin service would still reject it, runs of other characters, such as spaces, become a hyphen,
 * so "Data Science" becomes "data-science".
 * Tags it accepts are kept as they are, so saving doesn't rename a tag set from the CLI, such as "prod-".
 */
export function sanitizeProjectTag(input: string): string {
  const tag = input.normalize("NFC").trim().toLowerCase();
  if (isValidProjectTag(tag)) return tag;
  return tag
    .replace(/[^\p{L}\p{M}\p{N}_-]+/gu, "-")
    .replace(/^[-_]+|-+$/g, "");
}

/** Returns whether a sanitized tag is accepted by the admin service. */
export function isValidProjectTag(tag: string): boolean {
  return (
    [...tag].length <= MAX_PROJECT_TAG_LENGTH && PROJECT_TAG_REGEX.test(tag)
  );
}

/** Returns the tags to add and remove to turn `before` into `after`. */
export function diffProjectTags(before: string[], after: string[]) {
  return {
    addTags: after.filter((t) => !before.includes(t)),
    removeTags: before.filter((t) => !after.includes(t)),
  };
}

/** Returns the distinct tags of the projects, sorted. */
export function getProjectTags(projects: V1Project[]): string[] {
  const tags = new Set(projects.flatMap((p) => p.tags ?? []));
  return [...tags].sort((a, b) => a.localeCompare(b));
}

/** Returns the projects that have any of the tags, or all projects if there are no tags. */
export function filterProjectsByTags(
  projects: V1Project[],
  tags: string[],
): V1Project[] {
  if (tags.length === 0) return projects;
  return projects.filter((p) => p.tags?.some((t) => tags.includes(t)));
}
