<script lang="ts">
  import Lock from "@rilldata/web-common/components/icons/Lock.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { getPersonalCanvases } from "@rilldata/web-admin/features/virtual-file-editor/canvas/selectors.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import CreatePersonalCanvasDialog from "@rilldata/web-admin/features/virtual-file-editor/canvas/CreatePersonalCanvasDialog.svelte";

  let {
    org,
    project,
  }: {
    org: string;
    project: string;
  } = $props();

  const runtimeClient = useRuntimeClient();
  let personalCanvasesQuery = $derived(
    getPersonalCanvases(runtimeClient, org, project),
  );
  let personalCanvases = $derived(
    ($personalCanvasesQuery.data ?? []).map((r) => {
      const name = r.meta?.name?.name ?? "";
      const displayName = r.canvas?.state?.validSpec?.displayName ?? name;
      const lastUpdated = r.canvas?.state?.dataRefreshedOn;
      return {
        name,
        displayName,
        lastUpdated,
      };
    }),
  );

  let createOpen = $state(false);
</script>

<section class="flex flex-col gap-3">
  <header class="flex items-center justify-between">
    <div class="flex items-center gap-2">
      <Lock size="16px" />
      <h2 class="text-lg font-medium">My canvases</h2>
      <span class="text-sm text-fg-secondary">Only visible to you</span>
    </div>
    <Button type="primary" onClick={() => (createOpen = true)}>
      Create canvas
    </Button>
  </header>

  {#if $personalCanvasesQuery.isLoading}
    <div class="text-fg-secondary">Loading...</div>
  {:else}
    <ul class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
      {#each personalCanvases as item (item.name)}
        <li class="border rounded p-3 hover:bg-gray-50">
          <a
            href={`/${org}/${project}/-/personal/${item.name}`}
            class="flex items-center gap-2"
          >
            <Lock size="12px" />
            <span class="font-medium">
              {item.displayName || item.name}
            </span>
          </a>
          {#if item.lastUpdated}
            <p class="text-xs text-fg-secondary">
              Updated {new Date(item.lastUpdated).toLocaleString()}
            </p>
          {/if}
        </li>
      {:else}
        <div>
          <p>You don't have any personal canvases yet.</p>
          <p>Create one to explore the project's data your way.</p>
        </div>
      {/each}
    </ul>
  {/if}
</section>

<CreatePersonalCanvasDialog bind:open={createOpen} {org} {project} />
