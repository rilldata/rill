import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";

export function getUserDomain() {
  return createAdminServiceGetCurrentUser({
    query: {
      select: (data) => {
        if (!data.user?.email) return "";
        return getDomain(data.user.email);
      },
    },
  });
}

export function userDomainIsPublic() {
  return createAdminServiceGetCurrentUser({
    query: {
      select: (data) => {
        if (!data.user?.email) return false;
        const domain = getDomain(data.user.email);
        return ((window as any).RillPublicEmailDomains as string[]).includes(
          domain,
        );
      },
    },
  });
}

function getDomain(email: string) {
  const domainParts = email.split("@");
  return domainParts.length ? domainParts[domainParts.length - 1] : "";
}
