<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";

  interface Resource {
    name: string;
    kind: string;
    state?: string;
    error?: string;
    path?: string;
  }

  export let title: string;
  export let resources: Resource[];
  export let expanded: boolean;
  export let onToggle: () => void;
  export let onSelect: (resource: Resource) => void;
  export let hoveredResource: string | null;
  export let onHover: (name: string | null) => void;

  function getStatusColor(resource: Resource): string {
    if (resource.error) return "var(--status-error, #DC2626)";
    const state = resource.state?.toUpperCase();
    if (state === "RECONCILING" || state === "COMPILING") {
      return "var(--status-warning, #F59E0B)";
    }
    return "var(--status-success, #10B981)";
  }

  function getStatusText(resource: Resource): string {
    const state = (resource.state || "").toUpperCase();
    if (resource.error) return "Error";
    if (state === "RECONCILING" || state === "COMPILING") return "Compiling";
    if (state === "OK") return "Ready";
    return state.charAt(0) + state.slice(1).toLowerCase();
  }
</script>

<div class="py-2">
  <button
    on:click={onToggle}
    class="w-full flex items-center gap-1 px-3 py-1.5 transition-colors section-toggle"
  >
    <div class="flex-shrink-0 w-4 h-4 flex items-center justify-center">
      <CaretDownIcon
        size="12px"
        className={`transition-transform ${!expanded ? "-rotate-90" : ""}`}
      />
    </div>
    <h3
      class="text-xs font-semibold uppercase tracking-wide"
      style="color: var(--fg-muted)"
    >
      {title} ({resources.length})
    </h3>
  </button>

  {#if expanded}
    <ul class="list-none p-0 m-0 w-full">
      {#each resources as resource, idx (resource.name)}
        <li
          class={`block w-full border transition-colors resource-row relative ${
            idx === 0 ? "rounded-t-lg" : "border-t-0"
          } ${
            idx === resources.length - 1 ? "rounded-b-lg" : ""
          } ${resource.error ? "border-l-4 border-l-red-600" : ""}`}
          style="border-color: var(--border)"
          on:mouseenter={() => onHover(resource.name)}
          on:mouseleave={() => onHover(null)}
        >
          <button
            on:click={() => onSelect(resource)}
            class="flex items-center gap-x-3 group px-4 py-3 w-full"
          >
            <!-- Icon Container -->
            <div
              class="flex-shrink-0 h-10 w-10 rounded-md flex items-center justify-center"
              style="background: var(--surface-subtle)"
            >
              <svelte:component
                this={resourceIconMapping[resource.kind]}
                size="20px"
                color="var(--fg-secondary)"
              />
            </div>

            <!-- Content -->
            <div class="flex-1 min-w-0">
              <div
                class={`text-sm font-semibold truncate ${
                  resource.error
                    ? "text-red-600 dark:text-red-400"
                    : "resource-name"
                }`}
                style:color={!resource.error
                  ? "var(--fg-secondary)"
                  : undefined}
              >
                {resource.name}
              </div>
              <div class="text-xs truncate" style="color: var(--fg-muted)">
                {resource.path || "No path"}
              </div>
            </div>

            <!-- Status Circle -->
            <div class="flex-shrink-0 flex items-center gap-x-2">
              <div
                class="h-2.5 w-2.5 rounded-full"
                style:background-color={getStatusColor(resource)}
                title={getStatusText(resource)}
              />
            </div>
          </button>

          <!-- Error Tooltip -->
          {#if hoveredResource === resource.name && resource.error}
            <div
              class="absolute left-0 top-full mt-1 z-50 bg-red-600 dark:bg-red-700 text-white text-xs rounded px-2 py-1 max-w-xs break-words"
            >
              {resource.error}
            </div>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style lang="postcss">
  .section-toggle:hover {
    background: var(--surface-subtle);
  }

  .resource-row:hover {
    background: var(--surface-subtle);
  }

  .resource-name {
    color: var(--fg-secondary);
  }

  :global(.group):hover .resource-name {
    color: var(--fg-primary);
  }
</style>
