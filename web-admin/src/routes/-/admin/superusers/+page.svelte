<!-- web-admin/src/routes/-/admin/superusers/+page.svelte -->
<script lang="ts">
  import {
    createAdminServiceListSuperusers,
    createAdminServiceSetSuperuser,
  } from "@rilldata/web-admin/client";
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";

  let bannerRef: ActionResultBanner;
  let newEmail = "";
  let confirmOpen = false;
  let confirmTitle = "";
  let confirmDescription = "";
  let confirmDestructive = false;
  let confirmAction: () => Promise<void> = async () => {};

  const queryClient = useQueryClient();
  const superusersQuery = createAdminServiceListSuperusers();
  const setSuperuser = createAdminServiceSetSuperuser();

  async function handleAdd() {
    if (!newEmail) return;
    try {
      await $setSuperuser.mutateAsync({ data: { email: newEmail, superuser: true } });
      bannerRef.show("success", `${newEmail} added as superuser`);
      newEmail = "";
      await queryClient.invalidateQueries();
    } catch (err) {
      bannerRef.show("error", `Failed: ${err}`);
    }
  }

  function handleRemove(email: string) {
    confirmTitle = "Remove Superuser";
    confirmDescription = `Remove superuser access for ${email}? They will lose access to this admin console.`;
    confirmDestructive = true;
    confirmAction = async () => {
      try {
        await $setSuperuser.mutateAsync({ data: { email, superuser: false } });
        bannerRef.show("success", `${email} removed as superuser`);
        await queryClient.invalidateQueries();
      } catch (err) {
        bannerRef.show("error", `Failed: ${err}`);
      }
    };
    confirmOpen = true;
  }
</script>

<AdminPageHeader
  title="Superusers"
  description="Manage who has superuser (super admin) access to Rill Cloud."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="card mb-6">
  <h2 class="card-title">Add Superuser</h2>
  <div class="form-row">
    <input type="email" class="input" placeholder="Email address" bind:value={newEmail}
      on:keydown={(e) => e.key === "Enter" && handleAdd()} />
    <button class="btn-primary" on:click={handleAdd}>Add Superuser</button>
  </div>
</div>

{#if $superusersQuery.data?.users?.length}
  <table class="w-full">
    <thead>
      <tr><th>Email</th><th>Display Name</th><th>Actions</th></tr>
    </thead>
    <tbody>
      {#each $superusersQuery.data.users as user}
        <tr>
          <td class="font-mono text-xs">{user.email}</td>
          <td>{user.displayName ?? "-"}</td>
          <td>
            <button class="action-btn destructive" on:click={() => handleRemove(user.email ?? "")}>
              Remove
            </button>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{/if}

<ConfirmDialog bind:open={confirmOpen} title={confirmTitle} description={confirmDescription}
  destructive={confirmDestructive} onConfirm={confirmAction} />

<style lang="postcss">
  .card { @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700; }
  .card-title { @apply text-sm font-semibold text-slate-900 dark:text-slate-100 mb-3; }
  .form-row { @apply flex gap-3 items-center flex-wrap; }
  .input {
    @apply px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600
      bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500;
  }
  .btn-primary { @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 whitespace-nowrap; }
  th { @apply text-left text-xs font-medium text-slate-500 uppercase tracking-wider px-4 py-2 border-b border-slate-200 dark:border-slate-700; }
  td { @apply px-4 py-3 text-sm text-slate-700 dark:text-slate-300 border-b border-slate-100 dark:border-slate-800; }
  .action-btn.destructive { @apply text-xs px-2 py-1 rounded border border-red-300 text-red-600 hover:bg-red-50 dark:border-red-700 dark:text-red-400; }
</style>
