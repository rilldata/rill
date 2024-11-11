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
        return emailIsPublic(data.user.email);
      },
    },
  });
}

export function emailIsPublic(email: string) {
  const domain = getDomain(email);
  return RillPublicEmailDomains.includes(domain);
}

function getDomain(email: string) {
  const domainParts = email.split("@");
  return domainParts.length ? domainParts[domainParts.length - 1] : "";
}
