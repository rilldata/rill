<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import SearchInput from "@rilldata/web-admin/features/admin/shared/SearchInput.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import {
    notifySuccess,
    notifyError,
  } from "@rilldata/web-admin/features/admin/shared/notify";
  import {
    searchUsers,
    createDeleteUserMutation,
  } from "@rilldata/web-admin/features/admin/users/selectors";
  import { assumedUser } from "@rilldata/web-admin/features/admin/users/assume-state";
  import { useQueryClient } from "@tanstack/svelte-query";

  let searchQuery = "";
  let confirmOpen = false;
  let confirmTitle = "";
  let confirmDescription = "";
  let confirmDestructive = false;
  let confirmAction: () => Promise<void> = async () => {};
  let actionInProgress = "";

  const queryClient = useQueryClient();
  const deleteUser = createDeleteUserMutation();

  $: usersQuery = searchUsers(searchQuery);

  function handleSearch(e: CustomEvent<string>) {
    searchQuery = e.detail;
  }

  function handleAssume(email: string) {
    confirmTitle = "Open as User";
    confirmDescription = `You will start browsing Rill Cloud as ${email}. The session will expire after 60 minutes. Use the banner to unassume when done.`;
    confirmDestructive = false;
    confirmAction = async () => {
      // assume() navigates away via window.location.href, so no code after it runs
      assumedUser.assume(email);
    };
    confirmOpen = true;
  }

  function handleUnassume() {
    assumedUser.unassume();
  }

  function handleDelete(email: string) {
    confirmTitle = "Delete User";
    confirmDescription = `This will permanently delete the user ${email}. This action cannot be undone.`;
    confirmDestructive = true;
    confirmAction = async () => {
      actionInProgress = `delete:${email}`;
      try {
        await $deleteUser.mutateAsync({
          email,
        });
        notifySuccess(`User ${email} deleted`);
        await queryClient.invalidateQueries({
          predicate: (q) => q.queryKey[0] === "/v1/users/search",
        });
      } catch (err) {
        notifyError(`Failed to delete user: ${err}`);
      } finally {
        actionInProgress = "";
      }
    };
    confirmOpen = true;
  }
</script>

<AdminPageHeader
  title="Users"
  description="Search for users by email across all organizations."
/>

{#if $assumedUser}
  <div class="assume-banner">
    <span>Currently assumed as <strong>{$assumedUser}</strong></span>
    <button class="unassume-btn" on:click={handleUnassume}>
      Unassume
    </button>
  </div>
{/if}

<div class="mb-4 max-w-md">
  <SearchInput
    placeholder="Search by email (min 3 characters)..."
    on:search={handleSearch}
  />
</div>

{#if $usersQuery.isFetching && searchQuery.length >= 3}
  <div class="loading">
    <div class="spinner" />
    <span class="text-sm text-slate-500">Searching users...</span>
  </div>
{:else if $usersQuery.data?.users?.length}
  <p class="text-xs text-slate-500 mb-2">
    {$usersQuery.data.users.length} result{$usersQuery.data.users.length === 1
      ? ""
      : "s"}
  </p>
  <div class="results-table">
    <table class="w-full">
      <thead>
        <tr>
          <th>Email</th>
          <th>Display Name</th>
          <th>Created</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each $usersQuery.data.users as user}
          {@const isAssumed = $assumedUser === user.email}
          <tr>
            <td class="font-mono text-xs">{user.email}</td>
            <td>{user.displayName ?? "-"}</td>
            <td class="text-xs text-slate-500">
              {user.createdOn
                ? new Date(user.createdOn).toLocaleDateString()
                : "-"}
            </td>
            <td>
              <div class="flex gap-2">
                {#if isAssumed}
                  <button
                    class="action-btn assume-active"
                    on:click={handleUnassume}
                  >
                    Unassume
                  </button>
                {:else}
                  <button
                    class="action-btn"
                    on:click={() => handleAssume(user.email ?? "")}
                  >
                    Open as User
                  </button>
                {/if}
                <button
                  class="action-btn destructive"
                  disabled={actionInProgress === `delete:${user.email}`}
                  on:click={() => handleDelete(user.email ?? "")}
                >
                  {actionInProgress === `delete:${user.email}`
                    ? "Deleting..."
                    : "Delete"}
                </button>
              </div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{:else if searchQuery.length >= 3 && $usersQuery.isSuccess}
  <p class="text-sm text-slate-500">No users found for "{searchQuery}"</p>
{:else if searchQuery.length < 3}
  <p class="text-sm text-slate-400">
    Type at least 3 characters to search across all organizations.
  </p>
{/if}

<ConfirmDialog
  bind:open={confirmOpen}
  title={confirmTitle}
  description={confirmDescription}
  destructive={confirmDestructive}
  onConfirm={confirmAction}
/>

<style lang="postcss">
  table {
    @apply border-collapse;
  }

  th {
    @apply text-left text-xs font-medium text-slate-500 dark:text-slate-400
      uppercase tracking-wider px-4 py-2 border-b border-slate-200 dark:border-slate-700;
  }

  td {
    @apply px-4 py-3 text-sm text-slate-700 dark:text-slate-300
      border-b border-slate-100 dark:border-slate-800;
  }

  tr:hover td {
    @apply bg-slate-50 dark:bg-slate-800/50;
  }

  .action-btn {
    @apply text-xs px-2 py-1 rounded border border-slate-300 dark:border-slate-600
      text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700;
  }

  .action-btn:disabled {
    @apply opacity-50 cursor-not-allowed;
  }

  .action-btn.assume-active {
    @apply border-amber-400 bg-amber-50 text-amber-700
      hover:bg-amber-100
      dark:border-amber-600 dark:bg-amber-900/20 dark:text-amber-300
      dark:hover:bg-amber-900/40;
  }

  .action-btn.destructive {
    @apply border-red-300 text-red-600 hover:bg-red-50
      dark:border-red-700 dark:text-red-400 dark:hover:bg-red-900/20;
  }

  .assume-banner {
    @apply flex items-center gap-3 mb-4 px-4 py-2 rounded-md
      bg-amber-50 border border-amber-200 text-amber-800 text-sm
      dark:bg-amber-900/20 dark:border-amber-700 dark:text-amber-300;
  }

  .unassume-btn {
    @apply text-xs px-3 py-1 rounded border border-amber-400 bg-white text-amber-700
      hover:bg-amber-50
      dark:bg-amber-900/30 dark:border-amber-600 dark:text-amber-300
      dark:hover:bg-amber-900/50;
  }

  .loading {
    @apply flex items-center gap-2 py-4;
  }

  .spinner {
    @apply w-4 h-4 border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin;
  }
</style>
