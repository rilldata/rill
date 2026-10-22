import {
  createAdminServiceCreateReport,
  createAdminServiceDeleteReport,
  createAdminServiceEditReport,
} from "@rilldata/web-admin/client";
import type {
  FormMetadataProvider,
  MutationFactory,
} from "@rilldata/web-common/features/scheduled-reports/FormMetadataProvider.svelte.ts";
import {
  createAdminServiceCreateReportUsingToken,
  createAdminServiceDeleteReportUsingToken,
  createAdminServiceEditReportUsingToken,
} from "@rilldata/web-common/features/scheduled-reports/report-client-using-token.ts";

type ReportMutation = ReturnType<
  typeof createAdminServiceCreateReport | typeof createAdminServiceEditReport
>;
type ReportDeleteMutation = ReturnType<typeof createAdminServiceDeleteReport>;

export type ReportFormMetadataProvider = FormMetadataProvider<
  ReportMutation,
  ReportDeleteMutation
>;

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

export const reportDeleteMutationFactory: MutationFactory<ReportDeleteMutation> =
  {
    admin: () => createAdminServiceDeleteReport(),
    embed: (adminToken) => createAdminServiceDeleteReportUsingToken(adminToken),
  };
