<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";

  let bannerRef: ActionResultBanner;
  let org = "";
  let project = "";

  // Virtual files management will be implemented in a follow-up.
  // The API calls needed are:
  // - PullVirtualRepo (list)
  // - GetVirtualFile (read)
  // - DeleteVirtualFile (delete)
</script>

<AdminPageHeader
  title="Virtual Files"
  description="Browse, read, and delete virtual files in project deployments."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="card">
  <div class="form-row mb-4">
    <input type="text" class="input" placeholder="Organization name" bind:value={org} />
    <input type="text" class="input" placeholder="Project name" bind:value={project} />
    <button class="btn-primary" disabled>List Files (Coming Soon)</button>
  </div>
  <p class="text-sm text-slate-500">Virtual file management will be available in a future update.</p>
</div>

<style lang="postcss">
  .card { @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700; }
  .form-row { @apply flex gap-3 items-center flex-wrap; }
  .input {
    @apply px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600
      bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500;
  }
  .btn-primary { @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50; }
</style>
