<!-- web-admin/src/routes/-/admin/superusers/+page.svelte -->
<script lang="ts">
  import {
    createAdminServiceListSuperusers,
    createAdminServiceSetSuperuser,
  } from "@rilldata/web-admin/client";
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import {
    notifySuccess,
    notifyError,
  } from "@rilldata/web-admin/features/admin/shared/notify";
  import { useQueryClient } from "@tanstack/svelte-query";

  let newEmail = "";
  let addLoading = false;
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
    addLoading = true;
    try {
      await $setSuperuser.mutateAsync({ data: { email: newEmail, superuser: true } });
      notifySuccess( `${newEmail} added as superuser`);
      newEmail = "";
      await queryClient.invalidateQueries({
        predicate: (q) =>
          (q.queryKey[0] as string)?.includes("/v1/superuser"),
      });
    } catch (err) {
      notifyError( `Failed: ${err}`);
    } finally {
      addLoading = false;
    }
  }

  function handleRemove(email: string) {
    confirmTitle = "Remove Superuser";
    confirmDescription = `Remove superuser access for ${email}? They will lose access to this admin console.`;
    confirmDestructive = true;
    confirmAction = async () => {
      try {
        await $setSuperuser.mutateAsync({ data: { email, superuser: false } });
        notifySuccess( `${email} removed as superuser`);
        await queryClient.invalidateQueries({
          predicate: (q) =>
            (q.queryKey[0] as string)?.includes("/v1/superuser"),
        });
      } catch (err) {
        notifyError( `Failed: ${err}`);
      }
    };
    confirmOpen = true;
  }
</script>

<AdminPageHeader
  title="Superusers"
  description="Manage who has superuser (super admin) access across all of Rill Cloud."
/>

<div class="p-5 rounded-lg border border-slate-200 dark:border-slate-700 mb-6">
  <h2 class="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-3">
    Add Superuser
  </h2>
  <div class="flex gap-3 items-center flex-wrap">
    <input
      type="email"
      class="px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
      placeholder="Email address"
      bind:value={newEmail}
      on:keydown={(e) => e.key === "Enter" && handleAdd()}
    />
    <button
      class="px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 whitespace-nowrap flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
      on:click={handleAdd}
      disabled={addLoading || !newEmail}
    >
      {#if addLoading}
        <span
          class="inline-block w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"
        />
        Adding...
      {:else}
        Add Superuser
      {/if}
    </button>
  </div>
</div>

{#if $superusersQuery.isFetching}
  <div class="flex items-center gap-2 py-4">
    <div
      class="w-4 h-4 border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin"
    />
    <span class="text-sm text-slate-500">Loading superusers...</span>
  </div>
{:else if $superusersQuery.data?.users?.length}
  <p class="text-xs text-slate-500 dark:text-slate-400 mb-2">
    {$superusersQuery.data.users.length} superuser{$superusersQuery.data.users.length === 1 ? "" : "s"}
  </p>
  <table class="w-full">
    <thead>
      <tr>
        <th
          class="text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider px-4 py-2 border-b border-slate-200 dark:border-slate-700"
        >
          Email
        </th>
        <th
          class="text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider px-4 py-2 border-b border-slate-200 dark:border-slate-700"
        >
          Display Name
        </th>
        <th
          class="text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider px-4 py-2 border-b border-slate-200 dark:border-slate-700"
        >
          Actions
        </th>
      </tr>
    </thead>
    <tbody>
      {#each $superusersQuery.data.users as user}
        <tr>
          <td
            class="px-4 py-3 text-sm text-slate-700 dark:text-slate-300 border-b border-slate-100 dark:border-slate-800 font-mono text-xs"
          >
            {user.email}
          </td>
          <td
            class="px-4 py-3 text-sm text-slate-700 dark:text-slate-300 border-b border-slate-100 dark:border-slate-800"
          >
            {user.displayName ?? "-"}
          </td>
          <td
            class="px-4 py-3 text-sm text-slate-700 dark:text-slate-300 border-b border-slate-100 dark:border-slate-800"
          >
            <button
              class="text-xs px-2 py-1 rounded border border-red-300 text-red-600 hover:bg-red-50 dark:border-red-700 dark:text-red-400"
              on:click={() => handleRemove(user.email ?? "")}
            >
              Remove
            </button>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{/if}

<ConfirmDialog
  bind:open={confirmOpen}
  title={confirmTitle}
  description={confirmDescription}
  destructive={confirmDestructive}
  onConfirm={confirmAction}
/>
