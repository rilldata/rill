import { getLocale } from "@rilldata/web-common/lib/i18n/gen/runtime";

export function formatDate(value: string) {
  return new Date(value).toLocaleDateString(getLocale(), {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "numeric",
  });
}

export function capitalize(str: string) {
  return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase();
}
