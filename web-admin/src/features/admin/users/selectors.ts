// web-admin/src/features/admin/users/selectors.ts
import {
  createAdminServiceSearchUsers,
  createAdminServiceIssueRepresentativeAuthToken,
  createAdminServiceRevokeRepresentativeAuthTokens,
  createAdminServiceDeleteUser,
} from "@rilldata/web-admin/client";

export function searchUsers(emailQuery: string) {
  return createAdminServiceSearchUsers(
    { emailQuery },
    { query: { enabled: emailQuery.length >= 2 } },
  );
}

export function createAssumeUserMutation() {
  return createAdminServiceIssueRepresentativeAuthToken();
}

export function createUnassumeUserMutation() {
  return createAdminServiceRevokeRepresentativeAuthTokens();
}

export function createDeleteUserMutation() {
  return createAdminServiceDeleteUser();
}
