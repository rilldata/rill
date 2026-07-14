<script lang="ts">
  import { getEnvFileStore } from "@rilldata/web-common/features/env-management/env-file-store.ts";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import type { Snippet } from "svelte";

  // Gates its children on the first `.env` pull. Forms that allocate env var
  // names must not construct an EnvEditSession against an empty store, otherwise
  // the collision set is empty and a commit can overwrite an existing secret.
  const { children }: { children: Snippet } = $props();

  const envStore = getEnvFileStore();
</script>

{#await envStore.whenReady()}
  <div class="flex items-center justify-center p-8">
    <Spinner status={EntityStatus.Running} size="2rem" />
  </div>
{:then}
  {@render children()}
{/await}
