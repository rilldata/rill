import { derived, writable } from "svelte/store";
import { themeControl } from "./theme-control";
import type { V1Organization } from "@rilldata/web-admin/client";

export const organizationData = writable<V1Organization | undefined>(undefined);

export const organizationLogoUrl = derived(
  [themeControl, organizationData],
  ([$theme, $org]) =>
    $theme === "dark" && $org?.logoDarkUrl ? $org.logoDarkUrl : $org?.logoUrl,
);
