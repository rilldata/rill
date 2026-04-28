<script lang="ts">
  import {
    createAdminServiceDeleteProject,
    createAdminServiceHibernateProject,
    getAdminServiceGetProjectQueryKey,
    getAdminServiceListProjectsForOrganizationQueryKey,
    type RpcStatus,
    type V1ProjectPermissions,
  } from "@rilldata/web-admin/client";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog";
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { AxiosError } from "axios";
  import { Moon, Pencil, Share2, Trash2 } from "lucide-svelte";

  export let organization: string;
  export let project: string;
  export let permissions: V1ProjectPermissions | undefined;

  $: canEdit = !!permissions?.manageProject;
  $: canShare = !!permissions?.manageProjectMembers;
  $: canHibernate = !!permissions?.manageProject;
  $: canDelete = !!permissions?.admin;
  $: hasAnyAction = canEdit || canShare || canHibernate || canDelete;

  let menuOpen = false;
  let hibernateDialogOpen = false;
  let deleteDialogOpen = false;

  const hibernateMutation = createAdminServiceHibernateProject();
  const deleteMutation = createAdminServiceDeleteProject();

  $: hibernateResult = $hibernateMutation;
  $: deleteResult = $deleteMutation;

  async function hibernate() {
    try {
      await $hibernateMutation.mutateAsync({ org: organization, project });
      await queryClient.refetchQueries({
        queryKey: getAdminServiceGetProjectQueryKey(organization, project),
      });
      hibernateDialogOpen = false;
      eventBus.emit("notification", { message: "Project hibernated" });
    } catch (err) {
      const axiosError = err as AxiosError<RpcStatus>;
      eventBus.emit("notification", {
        message:
          axiosError.response?.data?.message ?? "Failed to hibernate project",
        type: "error",
      });
    }
  }

  async function deleteProject() {
    await $deleteMutation.mutateAsync({ org: organization, project });
    queryClient.removeQueries({
      queryKey: getAdminServiceGetProjectQueryKey(organization, project),
    });
    await queryClient.invalidateQueries({
      queryKey:
        getAdminServiceListProjectsForOrganizationQueryKey(organization),
    });
    eventBus.emit("notification", { message: "Deleted project" });
  }
</script>

{#if hasAnyAction}
  <DropdownMenu.Root bind:open={menuOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={menuOpen} ariaLabel="Project actions">
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="end" class="min-w-[160px]">
      {#if canEdit}
        <DropdownMenu.Item
          href={`/${organization}/${project}/-/settings`}
          class="flex items-center"
        >
          <Pencil size="12" />
          <span class="ml-2">Edit</span>
        </DropdownMenu.Item>
      {/if}
      {#if canShare}
        <DropdownMenu.Item
          href={`/${organization}/${project}?share=true`}
          class="flex items-center"
        >
          <Share2 size="12" />
          <span class="ml-2">Share</span>
        </DropdownMenu.Item>
      {/if}
      {#if canHibernate}
        <DropdownMenu.Item
          class="flex items-center"
          onclick={() => (hibernateDialogOpen = true)}
        >
          <Moon size="12" />
          <span class="ml-2">Hibernate</span>
        </DropdownMenu.Item>
      {/if}
      {#if canDelete}
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          class="flex items-center"
          type="destructive"
          onclick={() => (deleteDialogOpen = true)}
        >
          <Trash2 size="12" />
          <span class="ml-2">Delete</span>
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>

  <AlertDialog bind:open={hibernateDialogOpen}>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Hibernate project?</AlertDialogTitle>
        <AlertDialogDescription>
          The project "{project}" will be put into hibernation mode and will
          stop consuming resources. You can wake it up at any time.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <Button type="secondary" onClick={() => (hibernateDialogOpen = false)}>
          Cancel
        </Button>
        <Button
          type="primary"
          loading={hibernateResult.isPending}
          onClick={hibernate}
        >
          Hibernate
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>

  <AlertDialogGuardedConfirmation
    bind:open={deleteDialogOpen}
    title="Delete project?"
    description={`The project "${project}" will be permanently deleted along with all its dashboards, data, and settings. This action cannot be undone.`}
    confirmText={`delete ${project}`}
    confirmButtonText="Delete"
    confirmButtonType="destructive"
    loading={deleteResult.isPending}
    error={deleteResult.error?.message}
    onConfirm={deleteProject}
  />
{/if}
