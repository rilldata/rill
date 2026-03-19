<!-- web-admin/src/routes/-/admin/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
</script>

<AdminPageHeader
  title="Admin Console"
  description="Internal tools for managing users, billing, organizations, and more."
/>

<div class="quick-actions">
  <a href="/-/admin/users" class="action-card">
    <span class="action-title">Users</span>
    <span class="action-desc">Search, lookup, and manage user accounts</span>
  </a>
  <a href="/-/admin/billing" class="action-card">
    <span class="action-title">Billing</span>
    <span class="action-desc">Extend trials, repair billing, manage subscriptions</span>
  </a>
  <a href="/-/admin/organizations" class="action-card">
    <span class="action-title">Organizations</span>
    <span class="action-desc">Lookup orgs, set custom domains, manage plans</span>
  </a>
  <a href="/-/admin/projects" class="action-card">
    <span class="action-title">Projects</span>
    <span class="action-desc">Search projects, edit settings, hibernate or reset</span>
  </a>
  <a href="/-/admin/quotas" class="action-card">
    <span class="action-title">Quotas</span>
    <span class="action-desc">View and adjust organization and user quotas</span>
  </a>
  <a href="/-/admin/superusers" class="action-card">
    <span class="action-title">Superusers</span>
    <span class="action-desc">Manage superuser access</span>
  </a>
</div>

<style lang="postcss">
  .quick-actions {
    @apply grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4;
  }

  .action-card {
    @apply block p-4 rounded-lg border border-slate-200 dark:border-slate-700
      hover:border-blue-300 dark:hover:border-blue-600
      hover:shadow-sm transition-all;
  }

  .action-title {
    @apply block text-sm font-semibold text-slate-900 dark:text-slate-100 mb-1;
  }

  .action-desc {
    @apply block text-xs text-slate-500 dark:text-slate-400;
  }
</style>
