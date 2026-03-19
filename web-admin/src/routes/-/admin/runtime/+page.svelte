<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";

  let bannerRef: ActionResultBanner;

  // Runtime management will be implemented in a follow-up.
  // The API calls needed are:
  // - SudoIssueRuntimeManagerToken
  // - ListInstances (via runtime client)
  // - DeleteInstance (via runtime client)
</script>

<AdminPageHeader
  title="Runtime"
  description="Manage runtime instances, issue manager tokens, and view deployment infrastructure."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="card">
  <p class="text-sm text-slate-500">Runtime management will be available in a future update. Use the CLI for now:</p>
  <pre class="mt-2 p-3 bg-slate-100 dark:bg-slate-800 rounded text-xs font-mono">rill sudo runtime list-instances &lt;host&gt;
rill sudo runtime delete-instance &lt;host&gt; &lt;instance_id&gt;
rill sudo runtime manager-token &lt;host&gt;</pre>
</div>

<style lang="postcss">
  .card { @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700; }
</style>
