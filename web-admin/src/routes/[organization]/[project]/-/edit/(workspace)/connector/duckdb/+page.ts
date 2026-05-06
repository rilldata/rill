import { redirect } from "@sveltejs/kit";

export function load({ url }) {
  // We don't use `withEditorPrefix()` here because its store is set in +layout.svelte, which mounts after this load runs.
  throw redirect(307, url.pathname.split("/connector/")[0] + "/");
}
