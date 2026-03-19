<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";
  import {
    getOrganization,
    getOrgAdmins,
    createSetCustomDomainMutation,
    createJoinOrgMutation,
  } from "@rilldata/web-admin/features/admin/organizations/selectors";

  let bannerRef: ActionResultBanner;
  let orgName = "";
  let lookupOrg = "";
  let customDomainOrg = "";
  let customDomain = "";
  let joinOrg = "";
  let joinEmail = "";
  let joinRole = "admin";

  const setCustomDomain = createSetCustomDomainMutation();
  const joinOrgMutation = createJoinOrgMutation();

  $: orgQuery = getOrganization(lookupOrg);
  $: adminsQuery = getOrgAdmins(lookupOrg);

  function handleLookup() {
    lookupOrg = orgName;
  }

  async function handleSetCustomDomain() {
    if (!customDomainOrg || !customDomain) return;
    try {
      await $setCustomDomain.mutateAsync({
        data: { name: customDomainOrg, customDomain },
      });
      bannerRef.show("success", `Custom domain set for ${customDomainOrg}`);
    } catch (err) {
      bannerRef.show("error", `Failed: ${err}`);
    }
  }

  async function handleJoinOrg() {
    if (!joinOrg || !joinEmail) return;
    try {
      await $joinOrgMutation.mutateAsync({
        org: joinOrg,
        data: { email: joinEmail, role: joinRole, superuserForceAccess: true },
      });
      bannerRef.show("success", `${joinEmail} added to ${joinOrg} as ${joinRole}`);
    } catch (err) {
      bannerRef.show("error", `Failed: ${err}`);
    }
  }
</script>

<AdminPageHeader
  title="Organizations"
  description="Lookup organizations, view their details, set custom domains, and add users."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="sections">
  <!-- Org Lookup -->
  <section class="card">
    <h2 class="card-title">Organization Lookup</h2>
    <div class="form-row mb-4">
      <input
        type="text"
        class="input"
        placeholder="Organization name"
        bind:value={orgName}
        on:keydown={(e) => e.key === "Enter" && handleLookup()}
      />
      <button class="btn-primary" on:click={handleLookup}>Lookup</button>
    </div>

    {#if $orgQuery.data?.organization}
      {@const org = $orgQuery.data.organization}
      <div class="detail-grid">
        <div class="detail-item">
          <span class="detail-label">ID</span>
          <span class="detail-value font-mono">{org.id}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Name</span>
          <span class="detail-value">{org.name}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Description</span>
          <span class="detail-value">{org.description ?? "-"}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Billing Plan</span>
          <span class="detail-value">{org.billingPlanDisplayName ?? "-"}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Custom Domain</span>
          <span class="detail-value">{org.customDomain ?? "None"}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Created</span>
          <span class="detail-value">
            {org.createdOn ? new Date(org.createdOn).toLocaleDateString() : "-"}
          </span>
        </div>
      </div>

      {#if $adminsQuery.data?.members?.length}
        <h3 class="mt-4 text-xs font-semibold text-slate-500 uppercase">Members</h3>
        <div class="mt-2">
          {#each $adminsQuery.data.members as member}
            <div class="member-row">
              <span class="text-sm">{member.userEmail}</span>
              <span class="text-xs text-slate-500">{member.roleName}</span>
            </div>
          {/each}
        </div>
      {/if}
    {/if}
  </section>

  <!-- Set Custom Domain -->
  <section class="card">
    <h2 class="card-title">Set Custom Domain</h2>
    <div class="form-row">
      <input type="text" class="input" placeholder="Organization name" bind:value={customDomainOrg} />
      <input type="text" class="input" placeholder="Custom domain (e.g. analytics.acme.com)" bind:value={customDomain} />
      <button class="btn-primary" on:click={handleSetCustomDomain}>Set Domain</button>
    </div>
  </section>

  <!-- Join Organization -->
  <section class="card">
    <h2 class="card-title">Add User to Organization</h2>
    <div class="form-row">
      <input type="text" class="input" placeholder="Organization name" bind:value={joinOrg} />
      <input type="email" class="input" placeholder="User email" bind:value={joinEmail} />
      <select class="input" bind:value={joinRole}>
        <option value="admin">Admin</option>
        <option value="editor">Editor</option>
        <option value="viewer">Viewer</option>
      </select>
      <button class="btn-primary" on:click={handleJoinOrg}>Add User</button>
    </div>
  </section>
</div>

<style lang="postcss">
  .sections { @apply flex flex-col gap-6; }
  .card { @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700; }
  .card-title { @apply text-sm font-semibold text-slate-900 dark:text-slate-100 mb-3; }
  .form-row { @apply flex gap-3 items-center flex-wrap; }
  .input {
    @apply px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600
      bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500;
  }
  .btn-primary { @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 whitespace-nowrap; }
  .detail-grid { @apply grid grid-cols-2 lg:grid-cols-3 gap-3; }
  .detail-item { @apply flex flex-col; }
  .detail-label { @apply text-[11px] text-slate-500 dark:text-slate-400 uppercase tracking-wider; }
  .detail-value { @apply text-sm text-slate-900 dark:text-slate-100; }
  .member-row { @apply flex justify-between items-center px-3 py-1.5 rounded bg-slate-50 dark:bg-slate-800 mb-1; }
</style>
