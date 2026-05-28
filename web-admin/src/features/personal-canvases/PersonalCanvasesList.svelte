<script lang="ts">
  import { page } from "$app/stores";
  import Lock from "@rilldata/web-common/components/icons/Lock.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import CreatePersonalCanvasDialog from "./CreatePersonalCanvasDialog.svelte";
  import { usePersonalCanvases } from "./selectors";

  export let copyableCanvases: { name: string; displayName: string }[] = [];

  let createOpen = false;

  $: ({
    params: { organization, project },
  } = $page);

  $: query = usePersonalCanvases(organization, project);
  $: data = $query.data;
  $: items = data?.files ?? [];
</script>

<section class="flex flex-col gap-3">
  <header class="flex items-center justify-between">
    <div class="flex items-center gap-2">
      <Lock size="14px" />
      <h2 class="text-lg font-medium">My canvases</h2>
      <span class="text-sm text-gray-500">Only visible to you</span>
    </div>
    <Button type="primary" onClick={() => (createOpen = true)}>
      Create canvas
    </Button>
  </header>

  {#if $query.isLoading}
    <p class="text-sm text-gray-500">Loading...</p>
  {:else if items.length === 0}
    <div
      class="border border-dashed rounded p-6 text-center text-sm text-gray-600"
    >
      <p>You don't have any personal canvases yet.</p>
      <p>Create one to explore the project's data your way.</p>
    </div>
  {:else}
    <ul class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
      {#each items as item (item.name)}
        <li class="border rounded p-3 hover:bg-gray-50">
          <a
            href={`/${organization}/${project}/-/my-canvases/${encodeURIComponent(item.name ?? "")}`}
            class="flex items-center gap-2"
          >
            <Lock size="12px" />
            <span class="font-medium">
              {item.displayName || item.name}
            </span>
          </a>
          <p class="text-xs text-gray-500">
            Updated {item.updatedOn
              ? new Date(item.updatedOn).toLocaleString()
              : ""}
          </p>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<CreatePersonalCanvasDialog
  bind:open={createOpen}
  org={organization}
  project={project}
  {copyableCanvases}
/>
