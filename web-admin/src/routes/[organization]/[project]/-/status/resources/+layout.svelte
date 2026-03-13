<script lang="ts">
  import { page } from "$app/stores";

  $: basePath = `/${$page.params.organization}/${$page.params.project}/-/status/resources`;
  $: isGraphView =
    $page.route.id?.endsWith("/graph") ?? false;
</script>

<div class="flex flex-col size-full gap-y-4">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Resources</h2>
    <div class="view-toggle">
      <a
        href={basePath}
        class="toggle-btn"
        class:active={!isGraphView}
      >
        List
      </a>
      <a
        href="{basePath}/graph"
        class="toggle-btn"
        class:active={isGraphView}
      >
        Graph
      </a>
    </div>
  </div>
  <slot />
</div>

<style lang="postcss">
  .view-toggle {
    @apply flex rounded-sm border border-gray-200 overflow-hidden;
  }
  .toggle-btn {
    @apply px-3 py-1 text-sm font-medium text-fg-secondary no-underline;
  }
  .toggle-btn:hover {
    @apply bg-surface-hover;
  }
  .toggle-btn.active {
    @apply bg-primary-100 text-primary-600;
  }
</style>
