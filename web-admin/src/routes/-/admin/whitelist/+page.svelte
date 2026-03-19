<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import {
    getWhitelistedDomains,
    createAddWhitelistMutation,
    createRemoveWhitelistMutation,
  } from "@rilldata/web-admin/features/admin/whitelist/selectors";
  import { useQueryClient } from "@tanstack/svelte-query";

  let bannerRef: ActionResultBanner;
  let org = "";
  let activeOrg = "";
  let newDomain = "";
  let newRole = "viewer";
  let confirmOpen = false;
  let confirmTitle = "";
  let confirmDescription = "";
  let confirmAction: () => Promise<void> = async () => {};

  const queryClient = useQueryClient();
  const addWhitelist = createAddWhitelistMutation();
  const removeWhitelist = createRemoveWhitelistMutation();

  $: domainsQuery = getWhitelistedDomains(activeOrg);

  function handleLookup() {
    activeOrg = org;
  }

  async function handleAdd() {
    if (!activeOrg || !newDomain) return;
    try {
      await $addWhitelist.mutateAsync({
        org: activeOrg,
        data: { domain: newDomain, role: newRole },
      });
      bannerRef.show("success", `Domain ${newDomain} whitelisted for ${activeOrg}`);
      newDomain = "";
      await queryClient.invalidateQueries();
    } catch (err) {
      bannerRef.show("error", `Failed: ${err}`);
    }
  }

  function handleRemove(domain: string) {
    confirmTitle = "Remove Whitelisted Domain";
    confirmDescription = `Remove "${domain}" from the whitelist for ${activeOrg}?`;
    confirmAction = async () => {
      try {
        await $removeWhitelist.mutateAsync({ org: activeOrg, domain });
        bannerRef.show("success", `Domain ${domain} removed from whitelist`);
        await queryClient.invalidateQueries();
      } catch (err) {
        bannerRef.show("error", `Failed: ${err}`);
      }
    };
    confirmOpen = true;
  }
</script>

<AdminPageHeader
  title="Domain Whitelist"
  description="Manage whitelisted email domains for organizations."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="sections">
  <section class="card">
    <div class="form-row mb-4">
      <input type="text" class="input" placeholder="Organization name" bind:value={org}
        on:keydown={(e) => e.key === "Enter" && handleLookup()} />
      <button class="btn-primary" on:click={handleLookup}>Lookup</button>
    </div>

    {#if activeOrg}
      <div class="form-row mb-4">
        <input type="text" class="input" placeholder="Domain (e.g. acme.com)" bind:value={newDomain} />
        <select class="input" bind:value={newRole}>
          <option value="admin">Admin</option>
          <option value="editor">Editor</option>
          <option value="viewer">Viewer</option>
        </select>
        <button class="btn-primary" on:click={handleAdd}>Add Domain</button>
      </div>

      {#if $domainsQuery.data?.domains?.length}
        <table class="w-full">
          <thead>
            <tr><th>Domain</th><th>Role</th><th>Actions</th></tr>
          </thead>
          <tbody>
            {#each $domainsQuery.data.domains as d}
              <tr>
                <td class="font-mono text-xs">{d.domain}</td>
                <td class="text-xs">{d.role}</td>
                <td>
                  <button class="action-btn destructive" on:click={() => handleRemove(d.domain ?? "")}>
                    Remove
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {:else if $domainsQuery.isSuccess}
        <p class="text-sm text-slate-500">No whitelisted domains.</p>
      {/if}
    {/if}
  </section>
</div>

<ConfirmDialog bind:open={confirmOpen} title={confirmTitle} description={confirmDescription} onConfirm={confirmAction} />

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
  th { @apply text-left text-xs font-medium text-slate-500 uppercase tracking-wider px-4 py-2 border-b border-slate-200 dark:border-slate-700; }
  td { @apply px-4 py-3 text-sm text-slate-700 dark:text-slate-300 border-b border-slate-100 dark:border-slate-800; }
  .action-btn { @apply text-xs px-2 py-1 rounded border border-slate-300 dark:border-slate-600 text-slate-600 dark:text-slate-300; }
  .action-btn.destructive { @apply border-red-300 text-red-600 hover:bg-red-50 dark:border-red-700 dark:text-red-400; }
</style>
