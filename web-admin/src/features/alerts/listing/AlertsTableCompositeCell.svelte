<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceDeleteAlert } from "@rilldata/web-admin/client";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import AlertIcon from "@rilldata/web-common/components/icons/AlertIcon.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import ResourceListRow from "@rilldata/web-common/features/resources/ResourceListRow.svelte";
  import { timeAgo } from "@rilldata/web-common/lib/time/relative-time";
  import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { Pencil, Trash2Icon } from "lucide-svelte";
  import ProjectAccessControls from "../../projects/ProjectAccessControls.svelte";
  import AlertOwnerBullet from "./AlertOwnerBullet.svelte";
  import DeleteAlertConfirmDialog from "./DeleteAlertConfirmDialog.svelte";

  export let organization: string;
  export let project: string;
  export let id: string;
  export let title: string;
  export let lastTrigger: string | undefined;
  export let ownerId: string;
  export let lastTriggerErrorMessage: string | undefined;

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  const deleteAlert = createAdminServiceDeleteAlert();

  $: isCreatedByCode = !ownerId;

  let isDropdownOpen = false;
  let isDeleteConfirmOpen = false;

  // TODO: Consider adding ?edit=true query param to auto-open the edit dialog on the resource page
  function handleEdit() {
    goto(`/${organization}/${project}/-/alerts/${id}`);
  }

  async function handleDelete() {
    await $deleteAlert.mutateAsync({
      org: organization,
      project,
      name: id,
    });
    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(
        runtimeClient.instanceId,
      ),
    });
  }
</script>

<ResourceListRow
  href={`alerts/${id}`}
  {title}
  icon={AlertIcon}
  errorMessage={lastTriggerErrorMessage}
>
  <svelte:fragment slot="subtitle">
    {#if !lastTrigger}
      <span class="shrink-0">Hasn't been checked yet</span>
    {:else}
      <span class="shrink-0">Last checked {timeAgo(new Date(lastTrigger))}</span
      >
    {/if}
    <ProjectAccessControls {organization} {project}>
      <svelte:fragment slot="manage-project">
        <span class="shrink-0">•</span>
        <AlertOwnerBullet {organization} {project} {ownerId} />
      </svelte:fragment>
    </ProjectAccessControls>
  </svelte:fragment>

  <svelte:fragment slot="actions">
    {#if !isCreatedByCode}
      <DropdownMenu.Root bind:open={isDropdownOpen}>
        <DropdownMenu.Trigger class="flex-none">
          <IconButton rounded active={isDropdownOpen}>
            <ThreeDot size="16px" />
          </IconButton>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content align="start" class="min-w-[95px]">
          <DropdownMenu.Item
            class="font-normal flex items-center"
            onclick={handleEdit}
          >
            <Pencil size="12px" />
            <span class="ml-2">Edit</span>
          </DropdownMenu.Item>
          <DropdownMenu.Item
            class="font-normal flex items-center"
            type="destructive"
            onclick={() => {
              isDeleteConfirmOpen = true;
            }}
          >
            <Trash2Icon size="12px" />
            <span class="ml-2">Delete</span>
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    {/if}
  </svelte:fragment>
</ResourceListRow>

<DeleteAlertConfirmDialog
  bind:open={isDeleteConfirmOpen}
  {title}
  onDelete={handleDelete}
/>
