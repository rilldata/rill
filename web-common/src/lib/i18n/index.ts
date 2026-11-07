import { getLocale } from "@rilldata/web-common/lib/i18n/gen/runtime";
import { syncDocumentLocale, syncLuxonLocale } from "./locale-utils";

export function initializeI18n(): void {
  const locale = getLocale();
  syncDocumentLocale(locale);
  syncLuxonLocale(locale);
}

export {
  escapeHtml,
  localeDirection,
  syncDocumentLocale,
  syncLuxonLocale,
} from "./locale-utils";
