export const ORG_ROLES = ["admin", "editor", "viewer"];
export const PROJECT_ROLES = ["admin", "editor", "viewer"];

export const NAME_PATTERN = /^[a-zA-Z_][a-zA-Z0-9_-]*$/;

export function validateServiceName(value: string): string {
  if (!value.trim()) return "Name is required";
  if (!NAME_PATTERN.test(value.trim()))
    return "Must start with a letter or underscore, and contain only letters, digits, underscores, or hyphens";
  return "";
}

export function formatServiceDate(value: string | undefined): string {
  if (!value) return "-";
  return new Date(value).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function formatServiceDateTime(value: string | undefined): string {
  if (!value) return "-";
  return new Date(value).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "numeric",
  });
}
