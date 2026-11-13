<script lang="ts">
  import { page } from "$app/state";
  import PublicURLPasswordGate from "@rilldata/web-admin/features/public-urls/PublicURLPasswordGate.svelte";
  import type { Snippet } from "svelte";
  import type { LayoutData } from "./$types";

  let { data, children }: { data: LayoutData; children: Snippet } = $props();
</script>

{#if data.publicURLLocked}
  <PublicURLPasswordGate token={page.params.token ?? ""} />
{:else}
  {@render children()}
{/if}
