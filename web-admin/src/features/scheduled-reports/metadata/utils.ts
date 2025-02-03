import type { V1ReportSpec } from "@rilldata/web-common/runtime-client";
import cronstrue from "cronstrue";
import { DateTime } from "luxon";
import { V1ExportFormat } from "../../../client";

export function exportFormatToPrettyString(format: V1ExportFormat): string {
  switch (format) {
    case V1ExportFormat.EXPORT_FORMAT_UNSPECIFIED:
      return "Unspecified Format";
    case V1ExportFormat.EXPORT_FORMAT_CSV:
      return "CSV";
    case V1ExportFormat.EXPORT_FORMAT_XLSX:
      return "Excel (XLSX)";
    case V1ExportFormat.EXPORT_FORMAT_PARQUET:
      return "Parquet";
    default:
      return "Unknown";
  }
}

export function formatNextRunOn(nextRunOn: string, timeZone: string): string {
  // If the timezone is empty, interpret it as UTC
  if (timeZone === "") {
    timeZone = "UTC";
  }
  return DateTime.fromISO(nextRunOn)
    .setZone(timeZone)
    .toLocaleString(DateTime.DATETIME_FULL);
}

const suffixMap = {
  1: "st",
  2: "nd",
  3: "rd",
};
export function formatRefreshSchedule(reportSpec: V1ReportSpec) {
  if (!reportSpec.refreshSchedule?.cron) return "";
  let formattedRefreshSchedule = cronstrue.toString(
    reportSpec.refreshSchedule.cron,
    {
      verbose: true,
    },
  );
  formattedRefreshSchedule = formattedRefreshSchedule.replace(
    /on day (\d*) of the month/,
    (_, day: number) => {
      const suffix = suffixMap[day % 10] ?? "th";
      return `on the ${day}${suffix} of each month`;
    },
  );

  return formattedRefreshSchedule;
}
