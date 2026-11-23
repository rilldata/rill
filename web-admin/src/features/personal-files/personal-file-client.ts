/**
 * Personal file client helpers that work both in Rill Cloud and in embeds.
 * Rill Cloud authenticates with cookies; embeds authenticate with the embed's admin token.
 */

import {
  adminServiceGetPersonalFile,
  createAdminServiceCreatePersonalFile,
  createAdminServiceDeletePersonalFile,
  createAdminServiceListPersonalFiles,
  getAdminServiceGetPersonalFileQueryKey,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import type { MutationFactory } from "@rilldata/web-common/features/scheduled-reports/FormMetadataProvider.svelte.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  adminServiceGetPersonalFileUsingToken,
  createAdminServiceCreatePersonalFileUsingToken,
  createAdminServiceDeletePersonalFileUsingToken,
  createAdminServiceListPersonalFilesUsingToken,
} from "./personal-file-client-using-token.ts";

export const personalFileCreateMutationFactory: MutationFactory<
  ReturnType<typeof createAdminServiceCreatePersonalFile<RpcStatus>>
> = {
  admin: () => createAdminServiceCreatePersonalFile(),
  embed: (adminToken) =>
    createAdminServiceCreatePersonalFileUsingToken(adminToken),
};

export const personalFileDeleteMutationFactory: MutationFactory<
  ReturnType<typeof createAdminServiceDeletePersonalFile<RpcStatus>>
> = {
  admin: () => createAdminServiceDeletePersonalFile(),
  embed: (adminToken) =>
    createAdminServiceDeletePersonalFileUsingToken(adminToken),
};

/**
 * Creates a query listing the current user's personal files.
 * It must be called during component initialization, like any query creator.
 */
export function createListPersonalFilesQuery(org: string, project: string) {
  const adminToken = EmbedStore.getInstance()?.adminToken;
  return adminToken
    ? createAdminServiceListPersonalFilesUsingToken(org, project, adminToken)
    : createAdminServiceListPersonalFiles(org, project);
}

/**
 * Fetches a personal file through the query cache.
 * Pass `adminToken` to authenticate with an embed's admin token instead of cookies.
 */
export function fetchPersonalFile(
  org: string,
  project: string,
  name: string,
  { adminToken, staleTime }: { adminToken?: string; staleTime?: number } = {},
) {
  return queryClient.fetchQuery({
    queryKey: getAdminServiceGetPersonalFileQueryKey(org, project, name),
    queryFn: ({ signal }) =>
      adminToken
        ? adminServiceGetPersonalFileUsingToken(
            org,
            project,
            name,
            adminToken,
            signal,
          )
        : adminServiceGetPersonalFile(org, project, name, signal),
    staleTime,
  });
}

/**
 * Returns the link to a personal canvas, which depends on whether the app is embedded.
 */
export function personalCanvasHref(org: string, project: string, name: string) {
  return EmbedStore.isEmbedded()
    ? `/-/embed/personal/${name}`
    : `/${org}/${project}/-/personal/${name}`;
}

/**
 * Returns the link to the page listing personal canvases, which depends on whether the app is embedded.
 */
export function personalCanvasesHomeHref(org: string, project: string) {
  return EmbedStore.isEmbedded() ? "/-/embed" : `/${org}/${project}`;
}
