<script lang="ts">
  import { page } from "$app/stores";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { optimisticallySetStatus } from "./branch-actions";
  import {
    extractBranchFromPath,
    injectBranchIntoPath,
    removeBranchFromPath,
    requestSkipBranchInjection,
  } from "./branch-utils";
  import { isProdDeployment } from "./deployment-utils";
  import {
    getStatusDotClass,
    isTransitoryStatus,
  } from "../projects/status/display-utils";
  import {
    V1DeploymentStatus,
    createAdminServiceListDeployments,
    createAdminServiceStartDeployment,
    type V1Deployment,
  } from "../../client";

  export let organization: string;
  export let project: string;
  export let primaryBranch: string | undefined = undefined;

  let open = false;
  let resumingId: string | null = null;

  const startMutation = createAdminServiceStartDeployment();

  $: activeBranch = extractBranchFromPath($page.url.pathname);

  // Poll at 2s only while the dropdown is open (so the user sees live status
  // transitions). When closed, the cached data is sufficient; freshness is
  // maintained by invalidateDeployments() calls after create/delete mutations.
  $: deploymentsQuery = createAdminServiceListDeployments(
    organization,
    project,
    {},
    {
      query: {
        enabled: !!organization && !!project,
        refetchInterval: open ? 2000 : false,
      },
    },
  );

  $: deployments = $deploymentsQuery.data?.deployments ?? [];

  $: hasBranchDeployments = deployments.some(
    (d) => d.branch && d.branch !== primaryBranch,
  );

  $: isOnBranch = !!activeBranch && activeBranch !== primaryBranch;

  // Sort: production first, then alphabetically by branch name
  $: sortedDeployments = [...deployments].sort((a, b) => {
    const aIsProd = isProdDeployment(a);
    const bIsProd = isProdDeployment(b);
    if (aIsProd && !bIsProd) return -1;
    if (!aIsProd && bIsProd) return 1;
    return (a.branch ?? "").localeCompare(b.branch ?? "");
  });

  // Current branch label for the trigger
  $: currentDeployment = isOnBranch
    ? deployments.find((d) => d.branch === activeBranch)
    : deployments.find(isProdDeployment);
  $: triggerLabel = isOnBranch
    ? truncateBranch(activeBranch ?? "")
    : truncateBranch(primaryBranch ?? "");

  function truncateBranch(branch: string): string {
    if (branch.length <= 20) return branch;
    return branch.slice(0, 19) + "…";
  }

  function getDeploymentHref(deployment: V1Deployment): string {
    const basePath = removeBranchFromPath($page.url.pathname);
    if (isProdDeployment(deployment)) return basePath + $page.url.search;
    return (
      injectBranchIntoPath(basePath, deployment.branch!) + $page.url.search
    );
  }

  function handleClick(deployment: V1Deployment) {
    if (isProdDeployment(deployment)) {
      requestSkipBranchInjection();
    }
    open = false;
  }

  function isHibernated(status: V1DeploymentStatus | undefined): boolean {
    return (
      status === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED ||
      status === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING
    );
  }

  async function handleResume(deployment: V1Deployment) {
    const id = deployment.id;
    if (!id || resumingId) return;
    resumingId = id;
    try {
      await $startMutation.mutateAsync({ deploymentId: id, data: {} });
      optimisticallySetStatus(
        organization,
        project,
        id,
        deployment.branch,
        V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING,
      );
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to resume branch: ${getRpcErrorMessage(err)}`,
      });
    } finally {
      resumingId = null;
    }
  }

  function statusDot(status: V1DeploymentStatus | undefined): string {
    return getStatusDotClass(
      status ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED,
    );
  }
</script>

{#if hasBranchDeployments || isOnBranch}
  <li class="branch-selector">
    <DropdownMenu.Root bind:open>
      <DropdownMenu.Trigger>
        {#snippet child({ props })}
          <button {...props} class="chip">
            <span class="status-dot {statusDot(currentDeployment?.status)}"
            ></span>
            <span>{triggerLabel}</span>
            <span class="caret" class:open>
              <CaretDownIcon size="10px" />
            </span>
          </button>
        {/snippet}
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="start" class="min-w-[200px] max-w-[300px]">
        <DropdownMenu.Group>
          <DropdownMenu.Label>All branches</DropdownMenu.Label>
        </DropdownMenu.Group>
        {#each sortedDeployments as deployment (deployment.id)}
          {@const prod = isProdDeployment(deployment)}
          {@const isSelected = prod
            ? !isOnBranch
            : activeBranch === deployment.branch}
          {@const hibernated = isHibernated(deployment.status)}
          {@const isResuming =
            resumingId === deployment.id ||
            (deployment.id != null &&
              isTransitoryStatus(
                deployment.status ??
                  V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED,
              ))}
          <DropdownMenu.CheckboxItem
            checked={isSelected}
            href={hibernated ? undefined : getDeploymentHref(deployment)}
            onclick={(e: MouseEvent) => {
              if (hibernated) {
                // Block opening a hibernated branch; offer Resume in place.
                e.preventDefault();
                void handleResume(deployment);
                return;
              }
              handleClick(deployment);
            }}
            class="flex items-center gap-x-2"
          >
            <div class="flex items-center gap-x-2 truncate flex-1">
              <span
                class="inline-block size-1.5 rounded-full flex-none {statusDot(
                  deployment.status,
                )}"
              ></span>
              <span class="truncate" class:text-fg-muted={hibernated}>
                {deployment.branch || primaryBranch || "main"}
              </span>
              {#if prod}
                <span class="text-[10px] text-fg-muted flex-none">
                  production
                </span>
              {/if}
            </div>
            {#if hibernated}
              <span class="text-[10px] text-primary-700 font-medium flex-none">
                {isResuming ? "Resuming…" : "Resume"}
              </span>
            {/if}
          </DropdownMenu.CheckboxItem>
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  </li>
{/if}

<style lang="postcss">
  .branch-selector {
    @apply flex items-center mr-2;
  }

  /* Styled to match the dimension chip used elsewhere in the header */
  .chip {
    @apply flex items-center gap-x-1;
    @apply px-2 py-0 rounded-2xl border;
    @apply bg-primary-50 border-primary-200 text-primary-800;
    @apply transition-colors;
  }

  .chip:hover {
    @apply bg-primary-100;
  }

  .status-dot {
    @apply size-1.5 rounded-full flex-none;
  }

  .caret {
    @apply flex-none transition-transform;
  }

  .caret.open {
    @apply rotate-180;
  }
</style>
