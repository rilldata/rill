<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import {
    V1DeploymentStatus,
    createAdminServiceListDeployments,
    type V1Deployment,
  } from "../../client";

  export let organization: string;
  export let project: string;
  export let activeBranch: string | undefined;
  export let primaryBranch: string | undefined;
  // The live deployment status from GetProject; overrides ListDeployments for the active branch
  export let activeDeploymentStatus: V1DeploymentStatus | undefined = undefined;

  let active = false;

  $: deploymentsQuery = createAdminServiceListDeployments(
    organization,
    project,
    {},
    {
      query: {
        enabled: !!organization && !!project,
      },
    },
  );

  $: rawDeployments = $deploymentsQuery.data?.deployments ?? [];

  // Deduplicate: keep only the most recently updated deployment per branch.
  // Override the active branch's status with the live value from GetProject.
  $: deployments = (() => {
    const byBranch = new Map<string, V1Deployment>();
    for (const d of rawDeployments) {
      const branch = d.branch ?? "";
      const existing = byBranch.get(branch);
      if (!existing || (d.updatedOn ?? "") > (existing.updatedOn ?? "")) {
        byBranch.set(branch, d);
      }
    }
    // Patch the active branch with the live status from GetProject
    const activeDep = activeBranch
      ? byBranch.get(activeBranch)
      : byBranch.get(primaryBranch ?? "");
    if (activeDep && activeDeploymentStatus) {
      byBranch.set(activeDep.branch ?? "", {
        ...activeDep,
        status: activeDeploymentStatus,
      });
    }
    return [...byBranch.values()];
  })();

  // Only show the selector if there are non-production branch deployments
  $: hasBranchDeployments = deployments.some(
    (d) => d.branch && d.branch !== primaryBranch,
  );

  $: isOnBranch = !!activeBranch && activeBranch !== primaryBranch;

  $: displayLabel = isOnBranch ? activeBranch : "Branches";

  // Build the "back to production" href (strip ?branch from current URL)
  $: productionHref = (() => {
    const url = new URL($page.url);
    url.searchParams.delete("branch");
    return url.pathname + url.search;
  })();

  // Sort: production first, then alphabetically by branch name
  $: sortedDeployments = [...deployments].sort((a, b) => {
    const aIsProd = a.branch === primaryBranch;
    const bIsProd = b.branch === primaryBranch;
    if (aIsProd && !bIsProd) return -1;
    if (!aIsProd && bIsProd) return 1;
    return (a.branch ?? "").localeCompare(b.branch ?? "");
  });

  function branchHref(deployment: V1Deployment): string {
    const url = new URL($page.url);
    if (deployment.branch === primaryBranch || !deployment.branch) {
      url.searchParams.delete("branch");
    } else {
      url.searchParams.set("branch", deployment.branch);
    }
    return url.pathname + url.search;
  }

  function getStatusLabel(
    status: V1DeploymentStatus | undefined,
  ): string | undefined {
    switch (status) {
      case V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING:
        return undefined; // Don't show a label for running
      case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
      case V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING:
        return "pending";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED:
        return "error";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED:
      case V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING:
        return "stopped";
      default:
        return undefined;
    }
  }

  function getStatusColor(status: V1DeploymentStatus | undefined): string {
    switch (status) {
      case V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING:
        return "text-green-600";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
      case V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING:
        return "text-yellow-600";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED:
        return "text-red-600";
      default:
        return "text-gray-400";
    }
  }
</script>

{#if hasBranchDeployments || isOnBranch}
  <DropdownMenu.Root bind:open={active}>
    <DropdownMenu.Trigger asChild let:builder>
      <Chip
        removable={isOnBranch}
        {active}
        builders={[builder]}
        type={isOnBranch ? "amber" : "time"}
        removeTooltipText="Back to production"
        onRemove={() => {
          void goto(productionHref);
        }}
      >
        <div slot="body" class="flex items-center gap-x-1.5">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 16 16"
            fill="currentColor"
            class="size-3.5"
          >
            <path
              fill-rule="evenodd"
              d="M4.75 2a.75.75 0 0 1 .75.75v2.906c.495-.09 1.09-.016 1.658.242.6.272 1.09.74 1.388 1.062l.035.038c.343.378.655.666.962.836.303.168.618.236 1.057.178a.75.75 0 1 1 .196 1.487c-.766.101-1.384-.05-1.907-.339-.516-.286-.938-.69-1.27-1.057l-.034-.036c-.312-.34-.641-.639-.982-.793-.335-.152-.687-.177-1.103-.07v3.544a.75.75 0 0 1-1.5 0V2.75A.75.75 0 0 1 4.75 2ZM3 13.75a1.75 1.75 0 1 1 3.5 0 1.75 1.75 0 0 1-3.5 0Zm1.75-.25a.25.25 0 1 0 0 .5.25.25 0 0 0 0-.5ZM10.25 13.75a1.75 1.75 0 1 1 3.5 0 1.75 1.75 0 0 1-3.5 0Zm1.75-.25a.25.25 0 1 0 0 .5.25.25 0 0 0 0-.5Z"
              clip-rule="evenodd"
            />
          </svg>
          <span class="truncate max-w-[120px]">{displayLabel}</span>
        </div>
      </Chip>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content
      align="start"
      class="flex flex-col min-w-[200px] max-w-[300px]"
    >
      <DropdownMenu.Label>Branch deployments</DropdownMenu.Label>
      <DropdownMenu.Separator />
      {#each sortedDeployments as deployment (deployment.id)}
        {@const isProd = deployment.branch === primaryBranch}
        {@const isSelected = isProd
          ? !isOnBranch
          : activeBranch === deployment.branch}
        {@const statusLabel = getStatusLabel(deployment.status)}
        {@const statusColor = getStatusColor(deployment.status)}
        <DropdownMenu.Item
          href={branchHref(deployment)}
          class="flex items-center justify-between gap-x-2"
        >
          <div class="flex items-center gap-x-2 truncate">
            <span
              class="inline-block size-1.5 rounded-full flex-none {statusColor}"
              style="background-color: currentColor;"
            />
            <span class="truncate" class:font-medium={isSelected}>
              {isProd ? `${deployment.branch} (production)` : deployment.branch}
            </span>
          </div>
          {#if statusLabel}
            <span class="text-[10px] {statusColor} flex-none"
              >{statusLabel}</span
            >
          {/if}
        </DropdownMenu.Item>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
