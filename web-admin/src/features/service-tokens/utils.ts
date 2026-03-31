import { capitalize } from "@rilldata/web-common/components/table/utils";

export { capitalize };

export const NONE_ROLE = "";
export const ORG_ROLES = ["admin", "editor", "viewer", "guest", NONE_ROLE];
export const PROJECT_ROLES = ["admin", "editor", "viewer"];

export function formatOrgRole(role: string | undefined): string {
  if (!role) return "None";
  return capitalize(role);
}

export const NAME_PATTERN = /^[a-zA-Z_][a-zA-Z0-9_-]*$/;

export function validateServiceName(value: string): string {
  if (!value.trim()) return "Name is required";
  if (!NAME_PATTERN.test(value.trim()))
    return "Must start with a letter or underscore, and contain only letters, digits, underscores, or hyphens";
  return "";
}

function isValidDate(value: string): boolean {
  const date = new Date(value);
  return !isNaN(date.getTime()) && date.getFullYear() > 1970;
}

export function formatServiceDate(value: string | undefined): string {
  if (!value || !isValidDate(value)) return "-";
  return new Date(value).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function formatServiceDateTime(value: string | undefined): string {
  if (!value || !isValidDate(value)) return "-";
  return new Date(value).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "numeric",
  });
}
