<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import CreatePersonalCanvasDialog from "@rilldata/web-admin/features/personal-files/canvas/CreatePersonalCanvasDialog.svelte";
  import DeletePersonalCanvasDialog from "@rilldata/web-admin/features/personal-files/canvas/DeletePersonalCanvasDialog.svelte";
  import { getPersonalFilteredResources } from "@rilldata/web-admin/features/personal-files/selectors.ts";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { getOrgAndProjectForEmbed } from "@rilldata/web-common/features/embeds/selectors.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ChevronDown, Lock, Plus, Trash } from "lucide-svelte";

  const runtimeClient = useRuntimeClient();

  const orgAndProjectQuery = getOrgAndProjectForEmbed(runtimeClient);
  let { organization, project } = $derived($orgAndProjectQuery);

  let personalCanvasesQuery = $derived(
    getPersonalFilteredResources(
      runtimeClient,
      organization,
      project,
      ResourceKind.Canvas,
    ),
  );
  let personalCanvases = $derived(
    ($personalCanvasesQuery.data ?? []).map((r) => {
      const name = r.meta?.name?.name ?? "";
      const displayName =
        r.canvas?.state?.validSpec?.displayName ||
        r.canvas?.spec?.displayName ||
        name;
      return { name, displayName };
    }),
  );

  let open = $state(false);
  let createOpen = $state(false);
  let deleteOpen = $state(false);
  let canvasToDelete = $state<{ name: string; displayName: string }>();

  function onDeleteClick(
    e: MouseEvent,
    canvas: { name: string; displayName: string },
  ) {
    // Don't navigate to the canvas through the parent item.
    e.preventDefault();
    e.stopPropagation();
    canvasToDelete = canvas;
    open = false;
    deleteOpen = true;
  }

  async function onDeleted() {
    // Leave the deleted canvas' page if it is open.
    if (
      page.route.id?.includes("/personal/") &&
      page.params.name === canvasToDelete?.name
    ) {
      await goto("/-/embed");
    }
  }
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <Button
        {...props}
        type="secondary"
        noStroke
        label={m.embed_personal_canvases_button()}
      >
        <Lock size="14px" class="text-icon-muted" />
        {m.embed_personal_canvases_button()}
        <ChevronDown size="14px" />
      </Button>
    {/snippet}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="end" class="min-w-56 max-h-96 overflow-y-auto">
    {#each personalCanvases as canvas (canvas.name)}
      <DropdownMenu.Item
        class="justify-between"
        href={`/-/embed/personal/${canvas.name}`}
        preloadData={false}
      >
        <span class="truncate">{canvas.displayName}</span>
        <button
          class="text-fg-muted hover:text-red-600 shrink-0"
          aria-label={m.embed_personal_canvases_delete({
            name: canvas.displayName,
          })}
          onclick={(e) => onDeleteClick(e, canvas)}
        >
          <Trash size="14px" />
        </button>
      </DropdownMenu.Item>
    {:else}
      {#if !$personalCanvasesQuery.isPending}
        <!-- Not a DropdownMenu.Label, which is a group heading and must be inside a DropdownMenu.Group. -->
        <div class="px-2 py-1.5 text-xm text-fg-muted">
          {m.embed_personal_canvases_empty()}
        </div>
      {/if}
    {/each}
    <DropdownMenu.Separator />
    <DropdownMenu.Item onclick={() => (createOpen = true)}>
      <Plus size="14px" />
      {m.embed_personal_canvases_new()}
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>

{#if organization && project}
  <CreatePersonalCanvasDialog
    org={organization}
    {project}
    bind:open={createOpen}
    showTrigger={false}
  />
  {#if canvasToDelete}
    <DeletePersonalCanvasDialog
      bind:open={deleteOpen}
      org={organization}
      {project}
      name={canvasToDelete.name}
      displayName={canvasToDelete.displayName}
      {onDeleted}
    />
  {/if}
{/if}
