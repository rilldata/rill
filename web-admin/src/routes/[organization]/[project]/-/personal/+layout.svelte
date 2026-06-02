<script lang="ts">
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
  import type { PageData } from "./$types";
  import type { Snippet } from "svelte";

  let { data, children }: { data: PageData; children: Snippet } = $props();

  let loaded = $state(false);
  const client = useRuntimeClient();
  $effect(() => {
    fileArtifacts.setClient(client, data.fileIo);
    loaded = true;
  });
</script>

{#if loaded}
  {@render children()}
{/if}
