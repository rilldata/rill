import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { string } from "yup";

// Returns the validation schema for the row limit of a data export file (see ExportFileOptions.svelte).
// An empty value means no limit.
export function getExportLimitSchema() {
  return string().matches(/^\d*$/, m.report_form_row_limit_invalid());
}
