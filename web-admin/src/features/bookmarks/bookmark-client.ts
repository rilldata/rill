/**
 * Bookmark client helpers that work both in Rill Cloud and in embeds.
 * Rill Cloud authenticates with cookies; embeds authenticate with the embed's admin token.
 */

import {
  createAdminServiceCreateBookmark,
  createAdminServiceRemoveBookmark,
  createAdminServiceUpdateBookmark,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import type { MutationFactory } from "@rilldata/web-common/features/scheduled-reports/FormMetadataProvider.svelte.ts";
import { derived, type Readable } from "svelte/store";
import {
  createAdminServiceCreateBookmarkUsingToken,
  createAdminServiceRemoveBookmarkUsingToken,
  createAdminServiceUpdateBookmarkUsingToken,
  getAdminServiceListBookmarksUsingTokenQueryOptions,
} from "./bookmark-client-using-token.ts";

export const bookmarkCreateMutationFactory: MutationFactory<
  ReturnType<typeof createAdminServiceCreateBookmark<RpcStatus>>
> = {
  admin: () => createAdminServiceCreateBookmark(),
  embed: (adminToken) => createAdminServiceCreateBookmarkUsingToken(adminToken),
};

export const bookmarkUpdateMutationFactory: MutationFactory<
  ReturnType<typeof createAdminServiceUpdateBookmark<RpcStatus>>
> = {
  admin: () => createAdminServiceUpdateBookmark(),
  embed: (adminToken) => createAdminServiceUpdateBookmarkUsingToken(adminToken),
};

export const bookmarkRemoveMutationFactory: MutationFactory<
  ReturnType<typeof createAdminServiceRemoveBookmark<RpcStatus>>
> = {
  admin: () => createAdminServiceRemoveBookmark(),
  embed: (adminToken) => createAdminServiceRemoveBookmarkUsingToken(adminToken),
};

/**
 * Query options for the embed user's bookmarks on a dashboard.
 * The project ID is omitted; the server resolves it from the embed's admin token.
 */
export function getEmbedBookmarksQueryOptions(
  resourceKind: ResourceKind,
  resourceNameStore: Readable<string>,
) {
  const adminToken = EmbedStore.getInstance()?.adminToken ?? "";
  return derived(resourceNameStore, (resourceName) =>
    getAdminServiceListBookmarksUsingTokenQueryOptions(
      { resourceKind, resourceName },
      adminToken,
    ),
  );
}
