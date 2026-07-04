<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceDeleteReport,
    createAdminServiceTriggerReport,
  } from "@rilldata/web-admin/client";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import DeleteConfirmDialog from "@rilldata/web-common/features/resources/DeleteConfirmDialog.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { Pencil, PlayIcon, Trash2Icon } from "lucide-svelte";

  export let organization: string;
  export let project: string;
  export let id: string;
  export let title: string;
  export let isCreatedByCode: boolean;

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  const deleteReport = createAdminServiceDeleteReport();
  const triggerReport = createAdminServiceTriggerReport();

  let isDropdownOpen = false;
  let isDeleteConfirmOpen = false;

  async function handleRun() {
    try {
      await $triggerReport.mutateAsync({
        org: organization,
        project,
        name: id,
        data: {},
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

  function handleEdit() {
    void goto(`/${organization}/${project}/-/reports/${id}`);
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

{#if !isCreatedByCode}
  <div class="flex justify-end" data-no-row-click>
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
      <DropdownMenu.Content side="bottom" align="start" class="min-w-[95px]">
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
  </div>

  <DeleteConfirmDialog
    bind:open={isDeleteConfirmOpen}
    title="Delete this report?"
    onDelete={handleDelete}
  >
    The report "<strong>{title}</strong>" will be permanently deleted and will
    no longer send scheduled emails.
  </DeleteConfirmDialog>
{/if}
