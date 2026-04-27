// All mutations in this file bake in `superuserForceAccess: true`. The
// superuser console routinely operates on users the caller isn't a member of,
// so every mutation needs the flag. Wrapping `mutateAsync` here means call
// sites just pass the business args (e.g. `{ email }`) and cannot forget.
import {
  createAdminServiceDeleteUser,
  createAdminServiceSearchUsers,
} from "@rilldata/web-admin/client";
import { derived } from "svelte/store";

export function searchUsers(emailPattern: string) {
  return createAdminServiceSearchUsers(
    { emailPattern: `%${emailPattern}%` },
    { query: { enabled: emailPattern.length >= 3 } },
  );
}

export function createDeleteUserMutation() {
  const mutation = createAdminServiceDeleteUser();
  return derived(mutation, ($m) => ({
    ...$m,
    mutateAsync: (vars: { email: string }) =>
      $m.mutateAsync({
        email: vars.email,
        params: { superuserForceAccess: true },
      }),
  }));
}
