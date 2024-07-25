import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";

export function getUserDomain() {
  return createAdminServiceGetCurrentUser({
    query: {
      select: (data) => {
        if (!data.user?.email) return "";
        const domainParts = data.user.email.split("@");
        return domainParts.length ? domainParts[domainParts.length - 1] : "";
      },
    },
  });
}
