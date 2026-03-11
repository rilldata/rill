<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    useActiveDevDeployment,
    useCreateDevDeployment,
    invalidateDevDeployments,
  } from "./use-edit-session";

  export let organization: string;
  export let project: string;

  const user = createAdminServiceGetCurrentUser();
  const activeDevDeployment = useActiveDevDeployment(organization, project);
  const createMutation = useCreateDevDeployment();

  $: currentUserId = $user.data?.user?.id;
  $: deployment = $activeDevDeployment.data;
  $: isLoading = $activeDevDeployment.isLoading;

  // Determine session state
  $: hasActiveSession = !!deployment;
  $: isOwnSession =
    hasActiveSession && deployment?.ownerUserId === currentUserId;
  $: isOtherSession = hasActiveSession && !isOwnSession;

  $: label = isOwnSession
    ? "Resume editing"
    : isOtherSession
      ? "Editing locked"
      : "Edit";

  $: tooltipText = isOtherSession
    ? "Another user is currently editing this project"
    : isOwnSession
      ? "Return to your editing session"
      : "Start an editing session";

  $: isStarting = $createMutation.isPending;

  async function handleClick() {
    if (isOtherSession) return;

    if (isOwnSession) {
      await goto(`/${organization}/${project}/-/edit`);
      return;
    }

    // No active session: create one and navigate
    try {
      await $createMutation.mutateAsync({
        org: organization,
        project,
        data: {
          environment: "dev",
          editable: true,
        },
      });
      void invalidateDevDeployments(organization, project);
      await goto(`/${organization}/${project}/-/edit`);
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to start edit session: ${getRpcErrorMessage(err as any)}`,
      });
    }
  }
</script>

<Tooltip distance={8}>
  <Button
    type="secondary"
    disabled={isOtherSession || isStarting || isLoading}
    loading={isStarting}
    loadingCopy="Starting..."
    onClick={handleClick}
  >
    {label}
  </Button>
  <TooltipContent slot="tooltip-content">{tooltipText}</TooltipContent>
</Tooltip>
