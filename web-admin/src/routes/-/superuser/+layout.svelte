<script lang="ts">
  import SuperuserSidebar from "@rilldata/web-admin/features/superuser/layout/SuperuserSidebar.svelte";
</script>

<svelte:head>
  <title>Superuser Console | Rill</title>
</svelte:head>

<div class="flex h-screen overflow-hidden bg-surface-base">
  <SuperuserSidebar />
  <div class="flex-1 overflow-y-auto p-8">
    <slot />
  </div>
</div>
