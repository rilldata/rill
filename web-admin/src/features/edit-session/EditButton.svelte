<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import {
    branchPathPrefix,
    requestSkipBranchInjection,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    isActiveDeployment,
    useDevDeployments,
    useCreateDevDeployment,
    invalidateDevDeployments,
  } from "./use-edit-session";

  export let organization: string;
  export let project: string;

  const user = createAdminServiceGetCurrentUser();
  const devDeployments = useDevDeployments(organization, project);
  const createMutation = useCreateDevDeployment();

  $: currentUserId = $user.data?.user?.id;
  $: deployments = $devDeployments.data?.deployments ?? [];
  $: isLoading = $devDeployments.isLoading;

  // User's own active deployment (if any)
  $: ownDeployment =
    deployments.find(
      (d) => d.ownerUserId === currentUserId && isActiveDeployment(d),
    ) ?? null;

  // Another user's active deployment (single-editor lock for Phase 1)
  $: otherActiveDeployment =
    deployments.find(
      (d) => d.ownerUserId !== currentUserId && isActiveDeployment(d),
    ) ?? null;

  $: label = ownDeployment
    ? "Resume editing"
    : otherActiveDeployment
      ? "Editing locked"
      : "Edit";

  $: tooltipText = otherActiveDeployment
    ? "Another user is currently editing this project"
    : ownDeployment
      ? "Return to your editing session"
      : "Start an editing session";

  $: isOtherSession = !!otherActiveDeployment && !ownDeployment;
  $: isStarting = $createMutation.isPending;

  function editUrl(branch: string | undefined): string {
    return `/${organization}/${project}${branchPathPrefix(branch)}/-/edit`;
  }

  async function handleClick() {
    if (isOtherSession) return;

    if (ownDeployment) {
      // Skip branch injection; we're building the full URL ourselves
      requestSkipBranchInjection();
      await goto(editUrl(ownDeployment.branch));
      return;
    }

    // No active session: create one and navigate
    try {
      const resp = await $createMutation.mutateAsync({
        org: organization,
        project,
        data: {
          environment: "dev",
          editable: true,
        },
      });
      void invalidateDevDeployments(organization, project);
      requestSkipBranchInjection();
      await goto(editUrl(resp.deployment?.branch));
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
