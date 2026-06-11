import { page } from "$app/stores";
import { derived, writable, type Readable } from "svelte/store";
import { extractBranchFromPath } from "./branch-utils";

/**
 * Current project's primary branch. Populated by the project layout from the
 * GetProject query response. Read by `isBranchPreview`.
 */
export const primaryBranchStore = writable<string | undefined>(undefined);

/**
 * True when the URL's `@branch` segment is set and differs from the project's
 * primary branch — i.e. the user is viewing a non-prod branch deployment.
 *
 * Single source of truth for "are we in a branch preview?" across the cloud UI.
 */
export const isBranchPreview: Readable<boolean> = derived(
  [page, primaryBranchStore],
  ([$page, $primary]) => {
    const active = extractBranchFromPath($page.url.pathname);
    return !!active && active !== $primary;
  },
);
