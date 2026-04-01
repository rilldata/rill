<script lang="ts">
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { type ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import ResourceErrorMessage from "@rilldata/web-common/features/projects/status/ResourceErrorMessage.svelte";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { ExternalLinkIcon } from "lucide-svelte";

  export let organization: string;

  type OrgResource = {
    projectName: string;
    kind: string;
    name: string;
    reconcileStatus: string;
    reconcileError: string;
    stateUpdatedOn: string;
  };

  export let resources: OrgResource[];

  let searchText = "";
  let selectedProject = "";
  let selectedType = "";

  $: projectNames = [...new Set(resources.map((r) => r.projectName))].sort();
  $: resourceTypes = [...new Set(resources.map((r) => r.kind))].sort();

  $: filteredResources = resources.filter((r) => {
    if (selectedProject && r.projectName !== selectedProject) return false;
    if (selectedType && r.kind !== selectedType) return false;
    if (
      searchText &&
      !r.name.toLowerCase().includes(searchText.toLowerCase()) &&
      !r.projectName.toLowerCase().includes(searchText.toLowerCase())
    )
      return false;
    return true;
  });

  function mapReconcileStatus(status: string): V1ReconcileStatus {
    switch (status) {
      case "RECONCILE_STATUS_IDLE":
        return V1ReconcileStatus.RECONCILE_STATUS_IDLE;
      case "RECONCILE_STATUS_PENDING":
        return V1ReconcileStatus.RECONCILE_STATUS_PENDING;
      case "RECONCILE_STATUS_RUNNING":
        return V1ReconcileStatus.RECONCILE_STATUS_RUNNING;
      default:
        return V1ReconcileStatus.RECONCILE_STATUS_IDLE;
    }
  }

  function prettyKind(kind: string): string {
    return kind.replace("rill.runtime.v1.", "");
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return "—";
    return new Date(dateStr).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "2-digit",
    });
  }

  let openDropdownKey = "";
</script>

<div class="flex flex-col gap-y-4">
  <div class="flex flex-row items-center gap-x-3">
    <div class="flex-1 min-w-0">
      <Search
        bind:value={searchText}
        placeholder="Search resources"
        large
        autofocus={false}
        showBorderOnFocus={false}
      />
    </div>

    <select
      class="h-9 rounded-sm border border-border bg-surface-base px-2 text-sm text-fg-secondary font-medium hover:bg-surface-hover"
      bind:value={selectedProject}
    >
      <option value="">All projects</option>
      {#each projectNames as project}
        <option value={project}>{project}</option>
      {/each}
    </select>

    <select
      class="h-9 rounded-sm border border-border bg-surface-base px-2 text-sm text-fg-secondary font-medium hover:bg-surface-hover"
      bind:value={selectedType}
    >
      <option value="">All types</option>
      {#each resourceTypes as type}
        <option value={type}>{prettyKind(type)}</option>
      {/each}
    </select>

    {#if searchText || selectedProject || selectedType}
      <button
        class="shrink-0 text-xs text-primary-500 hover:text-primary-600"
        on:click={() => {
          searchText = "";
          selectedProject = "";
          selectedType = "";
        }}
      >
        Clear
      </button>
    {/if}
  </div>

  {#if filteredResources.length === 0}
    <p class="text-fg-secondary text-sm py-8 text-center">
      No resources match the current filters
    </p>
  {:else}
    <div class="overflow-x-auto border border-border rounded-sm">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-border bg-surface-subtle">
            <th class="px-3 py-2 text-left font-medium text-fg-secondary text-xs">Type</th>
            <th class="px-3 py-2 text-left font-medium text-fg-secondary text-xs">Name</th>
            <th class="px-3 py-2 text-left font-medium text-fg-secondary text-xs">Project</th>
            <th class="px-3 py-2 text-center font-medium text-fg-secondary text-xs w-12">Status</th>
            <th class="px-3 py-2 text-left font-medium text-fg-secondary text-xs">Last refresh</th>
            <th class="px-3 py-2 w-10"></th>
          </tr>
        </thead>
        <tbody>
          {#each filteredResources as resource (`${resource.projectName}:${resource.kind}:${resource.name}`)}
            {@const resourceKey = `${resource.projectName}:${resource.kind}:${resource.name}`}
            <tr class="border-b border-border last:border-b-0 hover:bg-surface-hover">
              <td class="px-3 py-3">
                <ResourceTypeBadge kind={resource.kind} />
              </td>
              <td class="px-3 py-3 text-fg-primary truncate max-w-[200px] font-mono text-xs">
                {resource.name}
              </td>
              <td class="px-3 py-3 text-fg-secondary truncate max-w-[160px]">
                {resource.projectName}
              </td>
              <td class="px-3 py-3">
                <ResourceErrorMessage
                  message={resource.reconcileError}
                  status={resource.reconcileError
                    ? V1ReconcileStatus.RECONCILE_STATUS_IDLE
                    : mapReconcileStatus(resource.reconcileStatus)}
                />
              </td>
              <td class="px-3 py-3 text-fg-secondary text-xs">
                {formatDate(resource.stateUpdatedOn)}
              </td>
              <td class="px-3 py-3">
                <DropdownMenu.Root
                  open={openDropdownKey === resourceKey}
                  onOpenChange={(isOpen) => {
                    openDropdownKey = isOpen ? resourceKey : "";
                  }}
                >
                  <DropdownMenu.Trigger class="flex-none" aria-label="Resource actions">
                    <IconButton rounded active={openDropdownKey === resourceKey} size={20}>
                      <ThreeDot size="16px" />
                    </IconButton>
                  </DropdownMenu.Trigger>
                  <DropdownMenu.Content align="start">
                    <DropdownMenu.Item
                      class="font-normal flex items-center"
                      href="/{organization}/{resource.projectName}/-/status/resources"
                    >
                      <div class="flex items-center">
                        <ExternalLinkIcon size="12px" />
                        <span class="ml-2">View in project</span>
                      </div>
                    </DropdownMenu.Item>
                  </DropdownMenu.Content>
                </DropdownMenu.Root>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>
