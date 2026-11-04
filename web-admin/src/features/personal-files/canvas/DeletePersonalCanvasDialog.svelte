<script lang="ts">
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";
  import { getAdminServiceListPersonalFilesQueryKey } from "@rilldata/web-admin/client";
  import { personalFileDeleteMutationFactory } from "@rilldata/web-admin/features/personal-files/personal-file-client.ts";
  import { createMutationFromFactory } from "@rilldata/web-common/features/scheduled-reports/FormMetadataProvider.svelte.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  let {
    open = $bindable(false),
    org,
    project,
    name,
    displayName,
    onDeleted,
  }: {
    open?: boolean;
    org: string;
    project: string;
    name: string;
    displayName: string;
    onDeleted?: () => void | Promise<void>;
  } = $props();

  const runtimeClient = useRuntimeClient();
  const deleteMutation = createMutationFromFactory(
    personalFileDeleteMutationFactory,
  );

  async function deleteCanvas() {
    await $deleteMutation.mutateAsync({ org, project, name });

    // Invalidate resources and personal files queries
    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(
        runtimeClient.instanceId,
        {},
      ),
    });
    await queryClient.invalidateQueries({
      queryKey: getAdminServiceListPersonalFilesQueryKey(org, project),
      type: "all",
    });

    eventBus.emit("notification", {
      type: "success",
      message: m.personal_files_dashboard_deleted_notification({
        name: displayName,
      }),
    });
    await onDeleted?.();
  }
</script>

<AlertDialogGuardedConfirmation
  bind:open
  title={m.canvas_delete_dashboard_title()}
  description={m.canvas_delete_dashboard_desc({ name: displayName })}
  confirmText={`delete ${displayName}`}
  confirmButtonText={m.settings_delete_button()}
  confirmButtonType="destructive"
  loading={$deleteMutation.isPending}
  error={$deleteMutation.error?.message}
  onConfirm={deleteCanvas}
>
  <div class="hidden"></div>
</AlertDialogGuardedConfirmation>
