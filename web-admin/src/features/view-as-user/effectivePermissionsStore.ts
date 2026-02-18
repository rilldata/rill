import { writable } from "svelte/store";
import type { V1ProjectPermissions } from "../../client";

/**
 * Store for effective project permissions when "View As" is active.
 * When null, the actual user's permissions should be used.
 * When set, these are the impersonated user's permissions (from server).
 */
export const effectiveProjectPermissionsStore =
  writable<V1ProjectPermissions | null>(null);
