import { getLocale } from "@rilldata/web-common/paraglide/runtime.js";
import { syncDocumentLocale } from "./document-locale";
import { syncLuxonLocale } from "./luxon-locale";

export function initializeI18n(): void {
  const locale = getLocale();
  syncDocumentLocale(locale);
  syncLuxonLocale(locale);
}

export { syncDocumentLocale } from "./document-locale";
export { syncLuxonLocale } from "./luxon-locale";
export { normalizeLocale } from "./normalize-locale";
export { escapeHtml } from "./escape-html";
