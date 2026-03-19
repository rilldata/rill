<!-- web-admin/src/routes/-/admin/quotas/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";
  import {
    getOrgForQuotas,
    createUpdateOrgQuotasMutation,
    createUpdateUserQuotasMutation,
  } from "@rilldata/web-admin/features/admin/quotas/selectors";
  import { useQueryClient } from "@tanstack/svelte-query";

  let bannerRef: ActionResultBanner;
  let quotaType: "org" | "user" = "org";
  let lookupValue = "";
  let activeOrg = "";
  let activeUser = "";
  let lookupDone = false;

  const queryClient = useQueryClient();
  const updateOrgQuotas = createUpdateOrgQuotasMutation();
  const updateUserQuotas = createUpdateUserQuotasMutation();

  // Quota fields for editing (org quotas)
  let projects = "";
  let deployments = "";
  let slotsTotal = "";
  let slotsPerDeployment = "";
  let outstandingInvites = "";
  let storageLimitBytes = "";

  // Quota fields for editing (user quotas)
  let singleuserOrgs = "";

  $: orgQuery = getOrgForQuotas(activeOrg);

  function handleLookup() {
    lookupDone = true;
    if (quotaType === "org") {
      activeOrg = lookupValue;
      activeUser = "";
    } else {
      activeUser = lookupValue;
      activeOrg = "";
    }
  }

  // Populate fields when org data loads
  $: if ($orgQuery.data?.organization?.quotas) {
    const q = $orgQuery.data.organization.quotas;
    projects = q.projects ?? "";
    deployments = q.deployments ?? "";
    slotsTotal = q.slotsTotal ?? "";
    slotsPerDeployment = q.slotsPerDeployment ?? "";
    outstandingInvites = q.outstandingInvites ?? "";
    storageLimitBytes = q.storageLimitBytesPerDeployment ?? "";
  }

  async function handleSaveQuotas() {
    try {
      if (quotaType === "org") {
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
        bannerRef.show("success", `Quotas updated for org: ${activeOrg}`);
      } else {
        await $updateUserQuotas.mutateAsync({
          data: {
            email: activeUser,
            singleuserOrgs: singleuserOrgs ? Number(singleuserOrgs) : undefined,
          },
        });
        bannerRef.show("success", `Quotas updated for user: ${activeUser}`);
      }
      await queryClient.invalidateQueries({
        predicate: (q) =>
          (q.queryKey[0] as string)?.includes("/v1/superuser/quotas") ||
          (q.queryKey[0] as string)?.includes("/v1/organizations"),
      });
    } catch (err) {
      bannerRef.show("error", `Failed to update quotas: ${err}`);
    }
  }
</script>

<AdminPageHeader
  title="Quotas"
  description="View and adjust resource quotas for organizations and users."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="sections">
  <section class="card">
    <div class="flex gap-4 mb-4">
      <label class="flex items-center gap-2 text-sm">
        <input type="radio" value="org" bind:group={quotaType} />
        Organization
      </label>
      <label class="flex items-center gap-2 text-sm">
        <input type="radio" value="user" bind:group={quotaType} />
        User
      </label>
    </div>

    <div class="form-row mb-4">
      <input
        type="text"
        class="input"
        placeholder={quotaType === "org" ? "Organization name" : "User email"}
        bind:value={lookupValue}
        on:keydown={(e) => e.key === "Enter" && handleLookup()}
      />
      <button class="btn-primary" on:click={handleLookup}>Lookup</button>
    </div>

    {#if quotaType === "org" && activeOrg && $orgQuery.data?.organization}
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
    {:else if quotaType === "user" && activeUser && lookupDone}
      <div class="quota-grid">
        <div class="quota-field">
          <label class="quota-label" for="singleuserOrgs">Single-user Orgs Limit</label>
          <input id="singleuserOrgs" type="number" class="input" bind:value={singleuserOrgs} />
        </div>
      </div>
      <p class="text-xs text-slate-500 mt-2">User quotas are limited to the single-user orgs field. Other quotas are managed at the org level.</p>

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
</style>
