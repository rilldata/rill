import {
  createAdminServiceCreateReport,
  createAdminServiceEditReport,
} from "@rilldata/web-admin/client";
import type {
  FormMetadataProvider,
  MutationFactory,
} from "@rilldata/web-common/features/scheduled-reports/FormMetadataProvider.svelte.ts";
import {
  createAdminServiceCreateReportUsingToken,
  createAdminServiceEditReportUsingToken,
} from "@rilldata/web-common/features/scheduled-reports/report-client-using-token.ts";

type ReportMutation = ReturnType<
  typeof createAdminServiceCreateReport | typeof createAdminServiceEditReport
>;

export type ReportFormMetadataProvider = FormMetadataProvider<ReportMutation>;

export function getReportMutationFactory(
  isEdit: boolean,
): MutationFactory<ReportMutation> {
  return isEdit
    ? {
        admin: () => createAdminServiceEditReport(),
        embed: (adminToken) =>
          createAdminServiceEditReportUsingToken(adminToken),
      }
    : {
        admin: () => createAdminServiceCreateReport(),
        embed: (adminToken) =>
          createAdminServiceCreateReportUsingToken(adminToken),
      };
}
