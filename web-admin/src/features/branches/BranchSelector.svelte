<script lang="ts">
  import { page } from "$app/stores";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import {
    extractBranchFromPath,
    injectBranchIntoPath,
    removeBranchFromPath,
    requestSkipBranchInjection,
  } from "./branch-utils";
  import { getStatusDotClass } from "../projects/status/display-utils";
  import {
    V1DeploymentStatus,
    createAdminServiceListDeployments,
    type V1Deployment,
  } from "../../client";

  export let organization: string;
  export let project: string;
  export let primaryBranch: string | undefined = undefined;

  let open = false;

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

  $: rawDeployments = $deploymentsQuery.data?.deployments ?? [];

  // Deduplicate: keep only the most recently updated deployment per branch.
  $: deployments = (() => {
    const byBranch = new Map<string, V1Deployment>();
    for (const d of rawDeployments) {
      const branch = d.branch ?? "";
      const existing = byBranch.get(branch);
      // updatedOn is an ISO 8601 timestamp; lexicographic comparison is correct.
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

  // Current branch label for the trigger
  $: currentDeployment = isOnBranch
    ? deployments.find((d) => d.branch === activeBranch)
    : deployments.find((d) => d.branch === primaryBranch);
  $: triggerLabel = isOnBranch
    ? truncateBranch(activeBranch ?? "")
    : truncateBranch(primaryBranch ?? "");

  function truncateBranch(branch: string): string {
    if (branch.length <= 12) return branch;
    return branch.slice(0, 11) + "…";
  }

  function isProd(deployment: V1Deployment): boolean {
    return deployment.branch === primaryBranch || !deployment.branch;
  }

  function getDeploymentHref(deployment: V1Deployment): string {
    const basePath = removeBranchFromPath($page.url.pathname);
    if (isProd(deployment)) return basePath + $page.url.search;
    return (
      injectBranchIntoPath(basePath, deployment.branch!) + $page.url.search
    );
  }

  function handleClick(deployment: V1Deployment) {
    if (isProd(deployment)) {
      requestSkipBranchInjection();
    }
    open = false;
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
      <DropdownMenu.Trigger asChild let:builder>
        <button use:builder.action {...builder} class="chip">
          <span class="status-dot {statusDot(currentDeployment?.status)}" />
          <span>{triggerLabel}</span>
          <span class="caret" class:open>
            <CaretDownIcon size="10px" />
          </span>
        </button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="start" class="min-w-[200px] max-w-[300px]">
        <DropdownMenu.Label>All branches</DropdownMenu.Label>
        {#each sortedDeployments as deployment (deployment.id)}
          {@const prod = isProd(deployment)}
          {@const isSelected = prod
            ? !isOnBranch
            : activeBranch === deployment.branch}
          <DropdownMenu.CheckboxItem
            checked={isSelected}
            href={getDeploymentHref(deployment)}
            on:click={() => handleClick(deployment)}
            class="flex items-center gap-x-2"
          >
            <div class="flex items-center gap-x-2 truncate">
              <span
                class="inline-block size-1.5 rounded-full flex-none {statusDot(
                  deployment.status,
                )}"
              />
              <span class="truncate">
                {deployment.branch ?? primaryBranch ?? ""}
              </span>
              {#if prod}
                <span class="text-[10px] text-fg-muted flex-none">
                  production
                </span>
              {/if}
            </div>
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
