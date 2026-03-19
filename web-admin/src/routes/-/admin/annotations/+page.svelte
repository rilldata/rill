<script lang="ts">
  import {
    createAdminServiceSudoUpdateAnnotations,
  } from "@rilldata/web-admin/client";
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";

  let bannerRef: ActionResultBanner;
  let org = "";
  let project = "";
  let annotationsJson = "{}";
  let loaded = false;

  const updateAnnotations = createAdminServiceSudoUpdateAnnotations();

  async function handleLoad() {
    if (!org || !project) return;
    try {
      // Load existing project annotations via GetProject
      // The project's annotations field contains the current values
      const resp = await fetch(`/v1/organizations/${org}/projects/${project}`);
      if (resp.ok) {
        const data = await resp.json();
        const existing = data.project?.annotations ?? {};
        annotationsJson = JSON.stringify(existing, null, 2);
        loaded = true;
        bannerRef.show("success", `Loaded annotations for ${org}/${project}`);
      } else {
        bannerRef.show("error", `Project not found or access denied`);
      }
    } catch (err) {
      bannerRef.show("error", `Failed to load annotations: ${err}`);
    }
  }

  async function handleSave() {
    if (!org || !project) return;
    try {
      const annotations = JSON.parse(annotationsJson);
      await $updateAnnotations.mutateAsync({
        data: { organization: org, project, annotations },
      });
      bannerRef.show("success", `Annotations updated for ${org}/${project}`);
    } catch (err) {
      if (err instanceof SyntaxError) {
        bannerRef.show("error", "Invalid JSON format");
      } else {
        bannerRef.show("error", `Failed: ${err}`);
      }
    }
  }
</script>

<AdminPageHeader
  title="Annotations"
  description="View and update project annotations (key-value metadata used for billing, categorization, etc.)."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="card">
  <div class="form-row mb-4">
    <input type="text" class="input" placeholder="Organization name" bind:value={org} />
    <input type="text" class="input" placeholder="Project name" bind:value={project} />
    <button class="btn-primary" on:click={handleLoad}>Load Current</button>
  </div>
  <div class="mb-4">
    <label class="text-xs font-medium text-slate-500 mb-1 block">
      Annotations (JSON) {#if !loaded}<span class="text-yellow-600">— click "Load Current" first to avoid overwriting</span>{/if}
    </label>
    <textarea
      class="input w-full h-32 font-mono text-xs"
      placeholder={'{"key": "value"}'}
      bind:value={annotationsJson}
    ></textarea>
  </div>
  <button class="btn-primary" on:click={handleSave}>Save Annotations</button>
</div>

<style lang="postcss">
  .card { @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700; }
  .form-row { @apply flex gap-3 items-center flex-wrap; }
  .input {
    @apply px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600
      bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500;
  }
  .btn-primary { @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700; }
  textarea { @apply resize-y; }
</style>
