<!-- web-admin/src/routes/-/admin/quotas/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import OrgSearchInput from "@rilldata/web-admin/features/admin/shared/OrgSearchInput.svelte";
  import {
    notifySuccess,
    notifyError,
  } from "@rilldata/web-admin/features/admin/shared/notify";
  import {
    getOrgForQuotas,
    createUpdateOrgQuotasMutation,
  } from "@rilldata/web-admin/features/admin/quotas/selectors";
  import { useQueryClient } from "@tanstack/svelte-query";

  let orgValue = "";
  let activeOrg = "";

  const queryClient = useQueryClient();
  const updateOrgQuotas = createUpdateOrgQuotasMutation();

  // Quota fields for editing
  let projects = "";
  let deployments = "";
  let slotsTotal = "";
  let slotsPerDeployment = "";
  let outstandingInvites = "";
  let storageLimitBytes = "";

  $: orgQuery = getOrgForQuotas(activeOrg);

  function handleOrgSelect(e: CustomEvent<string>) {
    activeOrg = e.detail;
  }

  // Populate fields when org data loads
  $: if ($orgQuery.data?.organization?.quotas) {
    const q = $orgQuery.data.organization.quotas;
    projects = q.projects != null ? String(q.projects) : "";
    deployments = q.deployments != null ? String(q.deployments) : "";
    slotsTotal = q.slotsTotal != null ? String(q.slotsTotal) : "";
    slotsPerDeployment = q.slotsPerDeployment != null ? String(q.slotsPerDeployment) : "";
    outstandingInvites = q.outstandingInvites != null ? String(q.outstandingInvites) : "";
    storageLimitBytes = q.storageLimitBytesPerDeployment ?? "";
  }

  async function handleSaveQuotas() {
    try {
      await $updateOrgQuotas.mutateAsync({
        data: {
          org: activeOrg,
          projects: projects ? Number(projects) : undefined,
          deployments: deployments ? Number(deployments) : undefined,
          slotsTotal: slotsTotal ? Number(slotsTotal) : undefined,
          slotsPerDeployment: slotsPerDeployment ? Number(slotsPerDeployment) : undefined,
          outstandingInvites: outstandingInvites ? Number(outstandingInvites) : undefined,
          storageLimitBytesPerDeployment: storageLimitBytes ? storageLimitBytes : undefined,
        },
      });
      notifySuccess(`Quotas updated for org: ${activeOrg}`);
      await queryClient.invalidateQueries({
        predicate: (q) =>
          (q.queryKey[0] as string)?.includes("/v1/superuser/quotas") ||
          (q.queryKey[0] as string)?.includes("/v1/organizations"),
      });
    } catch (err) {
      notifyError(`Failed to update quotas: ${err}`);
    }
  }
</script>

<AdminPageHeader
  title="Quotas"
  description="View and adjust resource quotas for organizations."
/>

<div class="sections">
  <section class="card">
    <div class="form-row mb-4">
      <div class="w-64">
        <OrgSearchInput
          bind:value={orgValue}
          placeholder="Search organization..."
          on:select={handleOrgSelect}
        />
      </div>
    </div>

    {#if activeOrg && $orgQuery.isFetching}
      <div class="loading">
        <div class="spinner" />
        <span class="text-sm text-slate-500">Loading quotas...</span>
      </div>
    {:else if activeOrg && $orgQuery.data?.organization}
      <div class="quota-grid">
        <div class="quota-field">
          <label class="quota-label" for="projects">Projects</label>
          <input id="projects" type="number" class="input" bind:value={projects} />
        </div>
        <div class="quota-field">
          <label class="quota-label" for="deployments">Deployments</label>
          <input id="deployments" type="number" class="input" bind:value={deployments} />
        </div>
        <div class="quota-field">
          <label class="quota-label" for="slotsTotal">Total Slots</label>
          <input id="slotsTotal" type="number" class="input" bind:value={slotsTotal} />
        </div>
        <div class="quota-field">
          <label class="quota-label" for="slotsPerDeployment">Slots per Deployment</label>
          <input id="slotsPerDeployment" type="number" class="input" bind:value={slotsPerDeployment} />
        </div>
        <div class="quota-field">
          <label class="quota-label" for="outstandingInvites">Outstanding Invites</label>
          <input id="outstandingInvites" type="number" class="input" bind:value={outstandingInvites} />
        </div>
        <div class="quota-field">
          <label class="quota-label" for="storageLimitBytes">Storage Limit (bytes)</label>
          <input id="storageLimitBytes" type="text" class="input" bind:value={storageLimitBytes} />
        </div>
      </div>

      <div class="mt-4">
        <button class="btn-primary" on:click={handleSaveQuotas}>Save Quotas</button>
      </div>
    {/if}
  </section>
</div>

<style lang="postcss">
  .sections { @apply flex flex-col gap-6; }
  .card { @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700; }
  .form-row { @apply flex gap-3 items-center flex-wrap; }
  .input {
    @apply px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600
      bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500;
  }
  .btn-primary { @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 whitespace-nowrap; }
  .quota-grid { @apply grid grid-cols-2 lg:grid-cols-3 gap-4; }
  .quota-field { @apply flex flex-col gap-1; }
  .quota-label { @apply text-xs font-medium text-slate-500 dark:text-slate-400; }
  .loading { @apply flex items-center gap-2 py-4; }
  .spinner { @apply w-4 h-4 border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin; }
</style>
