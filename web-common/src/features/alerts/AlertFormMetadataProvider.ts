import {
  createAdminServiceCreateAlert,
  createAdminServiceDeleteAlert,
  createAdminServiceEditAlert,
} from "@rilldata/web-admin/client";
import type {
  FormMetadataProvider,
  MutationFactory,
} from "@rilldata/web-common/features/scheduled-reports/FormMetadataProvider.svelte.ts";
import {
  createAdminServiceCreateAlertUsingToken,
  createAdminServiceDeleteAlertUsingToken,
  createAdminServiceEditAlertUsingToken,
} from "@rilldata/web-common/features/alerts/alert-client-using-token.ts";

type AlertMutation = ReturnType<
  typeof createAdminServiceCreateAlert | typeof createAdminServiceEditAlert
>;
type AlertDeleteMutation = ReturnType<typeof createAdminServiceDeleteAlert>;

export type AlertFormMetadataProvider = FormMetadataProvider<
  AlertMutation,
  AlertDeleteMutation
>;

export function getAlertMutationFactory(
  isEdit: boolean,
): MutationFactory<AlertMutation> {
  return isEdit
    ? {
        admin: () => createAdminServiceEditAlert(),
        embed: (adminToken) =>
          createAdminServiceEditAlertUsingToken(adminToken),
      }
    : {
        admin: () => createAdminServiceCreateAlert(),
        embed: (adminToken) =>
          createAdminServiceCreateAlertUsingToken(adminToken),
      };
}

export const alertDeleteMutationFactory: MutationFactory<AlertDeleteMutation> =
  {
    admin: () => createAdminServiceDeleteAlert(),
    embed: (adminToken) => createAdminServiceDeleteAlertUsingToken(adminToken),
  };
