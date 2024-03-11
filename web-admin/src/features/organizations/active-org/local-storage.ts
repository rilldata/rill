const ACTIVE_ORG_LOCAL_STORAGE_KEY_PREFIX = "activeOrg";

export function getActiveOrgLocalStorageKey(userId: string) {
  return `${ACTIVE_ORG_LOCAL_STORAGE_KEY_PREFIX}_${userId}`;
}
