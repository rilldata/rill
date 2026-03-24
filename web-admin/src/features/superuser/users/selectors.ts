import {
  createAdminServiceSearchUsers,
  createAdminServiceDeleteUser,
} from "@rilldata/web-admin/client";

export function searchUsers(emailPattern: string) {
  return createAdminServiceSearchUsers(
    { emailPattern: `%${emailPattern}%` },
    { query: { enabled: emailPattern.length >= 3 } },
  );
}

export function createDeleteUserMutation() {
  return createAdminServiceDeleteUser();
}
