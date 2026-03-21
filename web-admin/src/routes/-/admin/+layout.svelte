<script lang="ts">
  import AdminSidebar from "@rilldata/web-admin/features/admin/layout/AdminSidebar.svelte";
</script>

<svelte:head>
  <title>Admin Console | Rill</title>
</svelte:head>

<div class="flex h-screen overflow-hidden bg-white dark:bg-slate-900">
  <AdminSidebar />
  <div class="flex-1 overflow-y-auto p-8 bg-slate-50 dark:bg-slate-950">
    <slot />
  </div>
</div>
