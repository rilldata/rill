<script lang="ts">
  import Search from "@rilldata/web-common/components/search/Search.svelte";

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

  $: projectNames = [
    ...new Set(resources.map((r) => r.projectName)),
  ].sort();
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

  function getStatusDot(status: string): string {
    switch (status) {
      case "RECONCILE_STATUS_IDLE":
        return "bg-green-500";
      case "RECONCILE_STATUS_PENDING":
      case "RECONCILE_STATUS_RUNNING":
        return "bg-yellow-500";
      default:
        return "bg-red-500";
    }
  }

  function getStatusLabel(status: string): string {
    switch (status) {
      case "RECONCILE_STATUS_IDLE":
        return "OK";
      case "RECONCILE_STATUS_PENDING":
        return "Pending";
      case "RECONCILE_STATUS_RUNNING":
        return "Running";
      default:
        return "Error";
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
      year: "numeric",
      hour: "numeric",
      minute: "2-digit",
    });
  }
</script>

<section class="flex flex-col gap-y-4">
  <div class="flex flex-row items-center gap-x-4 min-h-9">
    <div class="flex-1 min-w-0 min-h-9">
      <Search
        bind:value={searchText}
        placeholder="Search resources"
        large
        autofocus={false}
        showBorderOnFocus={false}
      />
    </div>

    <select
      class="min-h-9 rounded-sm border bg-input px-2 py-1 text-sm text-fg-secondary font-medium hover:bg-surface-hover"
      bind:value={selectedProject}
    >
      <option value="">All projects</option>
      {#each projectNames as project}
        <option value={project}>{project}</option>
      {/each}
    </select>

    <select
      class="min-h-9 rounded-sm border bg-input px-2 py-1 text-sm text-fg-secondary font-medium hover:bg-surface-hover"
      bind:value={selectedType}
    >
      <option value="">All types</option>
      {#each resourceTypes as type}
        <option value={type}>{prettyKind(type)}</option>
      {/each}
    </select>

    {#if searchText || selectedProject || selectedType}
      <button
        class="shrink-0 text-sm text-primary-500 hover:text-primary-600 whitespace-nowrap"
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
    <div class="overflow-x-auto rounded-lg border border-border">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-border bg-surface-subtle">
            <th class="px-4 py-3 text-left font-medium text-fg-secondary"
              >Project</th
            >
            <th class="px-4 py-3 text-left font-medium text-fg-secondary"
              >Type</th
            >
            <th class="px-4 py-3 text-left font-medium text-fg-secondary"
              >Name</th
            >
            <th class="px-4 py-3 text-left font-medium text-fg-secondary"
              >Status</th
            >
            <th class="px-4 py-3 text-left font-medium text-fg-secondary"
              >Last Updated</th
            >
          </tr>
        </thead>
        <tbody>
          {#each filteredResources as resource (`${resource.projectName}:${resource.kind}:${resource.name}`)}
            <tr
              class="border-b border-border last:border-b-0 {resource.reconcileError
                ? 'bg-red-50'
                : ''}"
            >
              <td class="px-4 py-3 text-fg-primary font-medium">
                {resource.projectName}
              </td>
              <td class="px-4 py-3 text-fg-secondary">
                {prettyKind(resource.kind)}
              </td>
              <td class="px-4 py-3 text-fg-primary font-mono text-xs">
                {resource.name}
              </td>
              <td class="px-4 py-3">
                <span class="flex items-center gap-2">
                  <span
                    class="inline-block h-2 w-2 rounded-full {getStatusDot(
                      resource.reconcileStatus,
                    )}"
                  ></span>
                  <span class="text-fg-primary"
                    >{getStatusLabel(resource.reconcileStatus)}</span
                  >
                </span>
                {#if resource.reconcileError}
                  <p class="text-xs text-red-600 mt-1 truncate max-w-xs">
                    {resource.reconcileError}
                  </p>
                {/if}
              </td>
              <td class="px-4 py-3 text-fg-secondary">
                {formatDate(resource.stateUpdatedOn)}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</section>
