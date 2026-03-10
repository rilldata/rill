<script lang="ts">
  import { page } from "$app/stores";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import {
    extractBranchFromPath,
    injectBranchIntoPath,
    removeBranchFromPath,
    requestSkipBranchInjection,
  } from "@rilldata/web-admin/lib/branch-utils";
  import {
    V1DeploymentStatus,
    createAdminServiceGetProject,
    createAdminServiceListDeployments,
    type V1Deployment,
  } from "../../client";

  export let organization: string;
  export let project: string;
  export let onSelect: () => void = () => {};

  let subMenuOpen = false;

  $: activeBranch = extractBranchFromPath($page.url.pathname);

  $: projectQuery = createAdminServiceGetProject(organization, project);
  $: primaryBranch = $projectQuery.data?.project?.primaryBranch;

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
  $: deployments = (() => {
    const byBranch = new Map<string, V1Deployment>();
    for (const d of rawDeployments) {
      const branch = d.branch ?? "";
      const existing = byBranch.get(branch);
      if (!existing || (d.updatedOn ?? "") > (existing.updatedOn ?? "")) {
        byBranch.set(branch, d);
      }
    }
    return [...byBranch.values()];
  })();

  $: hasBranchDeployments = deployments.some(
    (d) => d.branch && d.branch !== primaryBranch,
  );

  $: isOnBranch = !!activeBranch && activeBranch !== primaryBranch;

  // Sort: production first, then alphabetically by branch name
  $: sortedDeployments = [...deployments].sort((a, b) => {
    const aIsProd = a.branch === primaryBranch;
    const bIsProd = b.branch === primaryBranch;
    if (aIsProd && !bIsProd) return -1;
    if (!aIsProd && bIsProd) return 1;
    return (a.branch ?? "").localeCompare(b.branch ?? "");
  });

  function getDeploymentHref(deployment: V1Deployment): string {
    const basePath = removeBranchFromPath($page.url.pathname);
    const isProd = deployment.branch === primaryBranch || !deployment.branch;
    const newPath = isProd
      ? basePath
      : injectBranchIntoPath(basePath, deployment.branch);
    return newPath + $page.url.search;
  }

  function handleClick(deployment: V1Deployment) {
    const isProd = deployment.branch === primaryBranch || !deployment.branch;
    if (isProd) {
      // Navigating to production; tell beforeNavigate to not re-inject @branch
      requestSkipBranchInjection();
    }
    onSelect();
  }

  function getStatusLabel(
    status: V1DeploymentStatus | undefined,
  ): string | undefined {
    switch (status) {
      case V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING:
        return undefined;
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
  <DropdownMenu.Sub bind:open={subMenuOpen}>
    <DropdownMenu.SubTrigger
      on:click={() => {
        subMenuOpen = !subMenuOpen;
      }}
    >
      Branch
    </DropdownMenu.SubTrigger>
    <DropdownMenu.SubContent class="flex flex-col min-w-[200px] max-w-[300px]">
      {#each sortedDeployments as deployment (deployment.id)}
        {@const isProd = deployment.branch === primaryBranch}
        {@const isSelected = isProd
          ? !isOnBranch
          : activeBranch === deployment.branch}
        {@const statusLabel = getStatusLabel(deployment.status)}
        {@const statusColor = getStatusColor(deployment.status)}
        <DropdownMenu.CheckboxItem
          checked={isSelected}
          href={getDeploymentHref(deployment)}
          on:click={() => handleClick(deployment)}
          class="flex items-center gap-x-2"
        >
          <div class="flex items-center gap-x-2 truncate">
            <span
              class="inline-block size-1.5 rounded-full flex-none {statusColor}"
              style="background-color: currentColor;"
            />
            <span class="truncate">
              {isProd ? `${deployment.branch} (production)` : deployment.branch}
            </span>
          </div>
          {#if statusLabel}
            <span class="text-[10px] {statusColor} flex-none"
              >{statusLabel}</span
            >
          {/if}
        </DropdownMenu.CheckboxItem>
      {/each}
    </DropdownMenu.SubContent>
  </DropdownMenu.Sub>
{/if}
