<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import type { V1Bookmark } from "@rilldata/web-admin/client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ParsedBookmark } from "@rilldata/web-admin/features/project-wide-bookmarks/ParsedBookmark.svelte.ts";

  let { bookmark }: { bookmark: V1Bookmark } = $props();

  const runtimeClient = useRuntimeClient();
  let parsedBookmark = $derived(new ParsedBookmark(runtimeClient, bookmark));
</script>

<Dropdown.Item href={parsedBookmark.fullUrl}>
  <div>{bookmark.displayName}</div>
  <div>{bookmark.description}</div>
</Dropdown.Item>
