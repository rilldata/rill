<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceDeleteReport,
    createAdminServiceTriggerReport,
  } from "@rilldata/web-admin/client";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import DeleteConfirmDialog from "@rilldata/web-common/features/resources/DeleteConfirmDialog.svelte";
  import ResourceListRow from "@rilldata/web-common/features/resources/ResourceListRow.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import cronstrue from "cronstrue";
  import { Pencil, PlayIcon, Trash2Icon } from "lucide-svelte";
  import ProjectAccessControls from "../../projects/ProjectAccessControls.svelte";
  import { formatRunDate } from "../tableUtils";
  import ReportOwnerBullet from "./ReportOwnerBullet.svelte";

  export let organization: string;
  export let project: string;
  export let id: string;
  export let title: string;
  export let lastRun: string | undefined;
  export let timeZone: string;
  export let frequency: string;
  export let ownerId: string;
  export let lastRunErrorMessage: string | undefined;

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  const deleteReport = createAdminServiceDeleteReport();
  const triggerReport = createAdminServiceTriggerReport();

  const humanReadableFrequency = cronstrue.toString(frequency);

  $: isCreatedByCode = !ownerId;

  let isDropdownOpen = false;
  let isDeleteConfirmOpen = false;

  async function handleRun() {
    try {
      await $triggerReport.mutateAsync({
        org: organization,
        project,
        name: id,
        data: undefined,
      });
      eventBus.emit("notification", {
        message: "Triggered an ad-hoc run of this report.",
        type: "success",
      });
      await queryClient.invalidateQueries({
        queryKey: getRuntimeServiceListResourcesQueryKey(
          runtimeClient.instanceId,
        ),
      });
    } catch {
      eventBus.emit("notification", {
        message: "Failed to trigger report",
        type: "error",
      });
    }
  }

  // TODO: Consider adding ?edit=true query param to auto-open the edit dialog on the resource page
  function handleEdit() {
    goto(`/${organization}/${project}/-/reports/${id}`);
  }

  async function handleDelete() {
    try {
      await $deleteReport.mutateAsync({
        org: organization,
        project,
        name: id,
      });
      await queryClient.invalidateQueries({
        queryKey: getRuntimeServiceListResourcesQueryKey(
          runtimeClient.instanceId,
        ),
      });
    } catch {
      eventBus.emit("notification", {
        message: "Failed to delete report",
        type: "error",
      });
    }
  }
</script>

<ResourceListRow
  href={`reports/${id}`}
  {title}
  icon={ReportIcon}
  errorMessage={lastRun ? lastRunErrorMessage : undefined}
>
  {#snippet subtitle()}
    {#if !lastRun}
      <span class="shrink-0">Hasn't run yet</span>
    {:else}
      <span class="shrink-0">Last run {formatRunDate(lastRun, timeZone)}</span>
    {/if}
    <span class="shrink-0">•</span>
    <span class="shrink-0 truncate">{humanReadableFrequency}</span>
    <ProjectAccessControls {organization} {project}>
      <svelte:fragment slot="manage-project">
        <span class="shrink-0">•</span>
        <ReportOwnerBullet {organization} {project} {ownerId} />
      </svelte:fragment>
    </ProjectAccessControls>
  {/snippet}

  {#snippet actions()}
    {#if !isCreatedByCode}
      <DropdownMenu.Root bind:open={isDropdownOpen}>
        <DropdownMenu.Trigger class="flex-none">
          <IconButton
            rounded
            active={isDropdownOpen}
            ariaLabel={`Actions for ${title}`}
          >
            <ThreeDot size="16px" />
          </IconButton>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content align="start" class="min-w-[95px]">
          <DropdownMenu.Item
            class="font-normal flex items-center"
            onclick={handleRun}
          >
            <PlayIcon size="12px" />
            <span class="ml-2">Run</span>
          </DropdownMenu.Item>
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
  {/snippet}
</ResourceListRow>

<DeleteConfirmDialog
  bind:open={isDeleteConfirmOpen}
  title="Delete this report?"
  onDelete={handleDelete}
>
  The report "<strong>{title}</strong>" will be permanently deleted and will no
  longer send scheduled emails.
</DeleteConfirmDialog>
