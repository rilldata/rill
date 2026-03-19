<script lang="ts">
  import AdminSidebar from "@rilldata/web-admin/features/admin/layout/AdminSidebar.svelte";
</script>

<svelte:head>
  <title>Admin Console | Rill</title>
</svelte:head>

<div class="admin-layout">
  <AdminSidebar />
  <div class="admin-content">
    <slot />
  </div>
</div>

<style lang="postcss">
  .admin-layout {
    @apply flex h-screen overflow-hidden;
  }

  .admin-content {
    @apply flex-1 overflow-y-auto p-8;
  }
</style>
