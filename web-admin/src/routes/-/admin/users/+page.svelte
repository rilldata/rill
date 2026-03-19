<!-- web-admin/src/routes/-/admin/users/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import SearchInput from "@rilldata/web-admin/features/admin/shared/SearchInput.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import {
    searchUsers,
    createAssumeUserMutation,
    createDeleteUserMutation,
  } from "@rilldata/web-admin/features/admin/users/selectors";
  import { useQueryClient } from "@tanstack/svelte-query";

  let searchQuery = "";
  let bannerRef: ActionResultBanner;
  let confirmOpen = false;
  let confirmTitle = "";
  let confirmDescription = "";
  let confirmDestructive = false;
  let confirmAction: () => Promise<void> = async () => {};

  const queryClient = useQueryClient();
  const assumeUser = createAssumeUserMutation();
  const deleteUser = createDeleteUserMutation();

  $: usersQuery = searchUsers(searchQuery);

  function handleSearch(e: CustomEvent<string>) {
    searchQuery = e.detail;
  }

  function handleAssume(email: string) {
    confirmTitle = "Assume User Identity";
    confirmDescription = `You will browse Rill Cloud as ${email}. Use "Unassume" to return to your own identity.`;
    confirmDestructive = false;
    confirmAction = async () => {
      try {
        await $assumeUser.mutateAsync({ data: { email } });
        bannerRef.show("success", `Now browsing as ${email}`);
      } catch (err) {
        bannerRef.show("error", `Failed to assume user: ${err}`);
      }
    };
    confirmOpen = true;
  }

  function handleDelete(email: string) {
    confirmTitle = "Delete User";
    confirmDescription = `This will permanently delete the user ${email}. This action cannot be undone.`;
    confirmDestructive = true;
    confirmAction = async () => {
      try {
        await $deleteUser.mutateAsync({
          email,
          superuserForceAccess: true,
        });
        bannerRef.show("success", `User ${email} deleted`);
        await queryClient.invalidateQueries({
          predicate: (q) => q.queryKey[0] === "/v1/users/search",
        });
      } catch (err) {
        bannerRef.show("error", `Failed to delete user: ${err}`);
      }
    };
    confirmOpen = true;
  }

  async function handleOpenAsUser(email: string) {
    // Assume the user's identity first, then open the main page
    try {
      await $assumeUser.mutateAsync({ data: { email } });
      window.open("/", "_blank");
      bannerRef.show("success", `Opened as ${email} in a new tab. Remember to unassume when done.`);
    } catch (err) {
      bannerRef.show("error", `Failed to assume user: ${err}`);
    }
  }
</script>

<AdminPageHeader
  title="Users"
  description="Search for users by email, assume their identity for debugging, or manage their accounts."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="mb-4 max-w-md">
  <SearchInput
    placeholder="Search by email (min 2 characters)..."
    on:search={handleSearch}
  />
</div>

{#if $usersQuery.isLoading && searchQuery.length >= 2}
  <p class="text-sm text-slate-500">Searching...</p>
{:else if $usersQuery.data?.users?.length}
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
                <button
                  class="action-btn"
                  on:click={() => handleAssume(user.email ?? "")}
                >
                  Assume
                </button>
                <button
                  class="action-btn"
                  on:click={() => handleOpenAsUser(user.email ?? "")}
                >
                  Open as User
                </button>
                <button
                  class="action-btn destructive"
                  on:click={() => handleDelete(user.email ?? "")}
                >
                  Delete
                </button>
              </div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{:else if searchQuery.length >= 2 && $usersQuery.isSuccess}
  <p class="text-sm text-slate-500">No users found for "{searchQuery}"</p>
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

  .action-btn.destructive {
    @apply border-red-300 text-red-600 hover:bg-red-50
      dark:border-red-700 dark:text-red-400 dark:hover:bg-red-900/20;
  }
</style>
