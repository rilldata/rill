import {
  createAdminServiceCreateAlert,
  createAdminServiceEditAlert,
} from "@rilldata/web-admin/client";
import type {
  FormMetadataProvider,
  MutationFactory,
} from "@rilldata/web-common/features/scheduled-reports/FormMetadataProvider.svelte.ts";
import {
  createAdminServiceCreateAlertUsingToken,
  createAdminServiceEditAlertUsingToken,
} from "@rilldata/web-common/features/alerts/alert-client-using-token.ts";

type AlertMutation = ReturnType<
  typeof createAdminServiceCreateAlert | typeof createAdminServiceEditAlert
>;

export type AlertFormMetadataProvider = FormMetadataProvider<AlertMutation>;

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
