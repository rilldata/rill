import type { V1Organization } from "@rilldata/web-admin/client";

export function getThemedLogoUrl(
  theme: "light" | "dark",
  org: V1Organization | undefined,
): string | undefined {
  return theme === "dark" && org?.logoDarkUrl ? org.logoDarkUrl : org?.logoUrl;
}
