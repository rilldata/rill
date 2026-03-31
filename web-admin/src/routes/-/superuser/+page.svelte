<script lang="ts">
  import SearchInput from "@rilldata/web-admin/features/superuser/shared/SearchInput.svelte";
  import ConfirmActionDialog from "@rilldata/web-admin/features/superuser/dialogs/ConfirmActionDialog.svelte";
  import GuardedDeleteDialog from "@rilldata/web-admin/features/superuser/dialogs/GuardedDeleteDialog.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    searchUsers,
    createDeleteUserMutation,
  } from "@rilldata/web-admin/features/superuser/users/selectors";
  import { getAdminServiceSearchUsersQueryKey } from "@rilldata/web-admin/client";
  import { assumedUser } from "@rilldata/web-admin/features/superuser/users/assume-state";
  import { useQueryClient } from "@tanstack/svelte-query";

  let searchQuery = "";

  // Assume user dialog state
  let assumeDialogOpen = false;
  let assumeEmail = "";

  // Delete user dialog state
  let deleteDialogOpen = false;
  let deleteEmail = "";
  let deleteLoading = false;
  let deleteError: string | undefined = undefined;

  const queryClient = useQueryClient();
  const deleteUser = createDeleteUserMutation();

  $: usersQuery = searchUsers(searchQuery);

  function handleSearch(e: CustomEvent<string>) {
    searchQuery = e.detail;
  }

  function handleUnassume() {
    assumedUser.unassume();
  }

  async function doAssume() {
    assumedUser.assume(assumeEmail, {});
  }

  async function doDelete() {
    deleteLoading = true;
    deleteError = undefined;
    try {
      await $deleteUser.mutateAsync({ email: deleteEmail });
      eventBus.emit("notification", {
        type: "success",
        message: `User ${deleteEmail} deleted`,
      });
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceSearchUsersQueryKey(),
      });
    } catch (err) {
      deleteError = `Failed to delete user: ${err}`;
      throw err;
    } finally {
      deleteLoading = false;
    }
  }
</script>

<p class="text-sm text-fg-secondary mb-4">
  Search for users by email across all organizations.
</p>

<div class="mb-4 max-w-md">
  <SearchInput
    placeholder="Search by email (min 3 characters)..."
    on:search={handleSearch}
  />
</div>

{#if $usersQuery.isFetching && searchQuery.length >= 3}
  <p class="text-sm text-fg-secondary py-4">Searching users...</p>
{:else if $usersQuery.data?.users?.length}
  <p class="text-sm text-fg-secondary mb-2">
    {$usersQuery.data.users.length} result{$usersQuery.data.users.length === 1
      ? ""
      : "s"}
  </p>
  <div class="rounded-lg border overflow-hidden">
    <table class="w-full border-collapse">
      <thead>
        <tr>
          <th
            class="text-left text-sm font-medium text-fg-secondary uppercase tracking-wider px-4 py-2 border-b"
            >Email</th
          >
          <th
            class="text-left text-sm font-medium text-fg-secondary uppercase tracking-wider px-4 py-2 border-b"
            >Display Name</th
          >
          <th
            class="text-left text-sm font-medium text-fg-secondary uppercase tracking-wider px-4 py-2 border-b"
            >Created</th
          >
          <th
            class="text-left text-sm font-medium text-fg-secondary uppercase tracking-wider px-4 py-2 border-b"
            >Actions</th
          >
        </tr>
      </thead>
      <tbody>
        {#each $usersQuery.data.users as user}
          {@const isAssumed = $assumedUser === user.email}
          <tr>
            <td class="px-4 py-3 text-sm font-mono text-fg-primary border-b"
              >{user.email}</td
            >
            <td class="px-4 py-3 text-sm text-fg-primary border-b"
              >{user.displayName ?? "-"}</td
            >
            <td class="px-4 py-3 text-sm text-fg-secondary border-b">
              {user.createdOn
                ? new Date(user.createdOn).toLocaleDateString()
                : "-"}
            </td>
            <td class="px-4 py-3 text-sm text-fg-primary border-b">
              <div class="flex gap-2">
                {#if isAssumed}
                  <Button
                    large
                    class="font-normal"
                    type="tertiary"
                    onClick={handleUnassume}>Unassume</Button
                  >
                {:else}
                  <Button
                    large
                    class="font-normal"
                    type="tertiary"
                    disabled={!user.email}
                    onClick={() => {
                      assumeEmail = user.email ?? "";
                      assumeDialogOpen = true;
                    }}>Open as User</Button
                  >
                {/if}
                <Button
                  large
                  class="font-normal"
                  type="secondary-destructive"
                  disabled={!user.email}
                  onClick={() => {
                    deleteEmail = user.email ?? "";
                    deleteDialogOpen = true;
                  }}
                >
                  Delete
                </Button>
              </div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{:else if searchQuery.length >= 3 && $usersQuery.isSuccess}
  <p class="text-sm text-fg-secondary">No users found for "{searchQuery}"</p>
{:else if searchQuery.length < 3}
  <p class="text-sm text-fg-muted">
    Type at least 3 characters to search across all organizations.
  </p>
{/if}

<ConfirmActionDialog
  bind:open={assumeDialogOpen}
  title="Open as User"
  description={`You will start browsing Rill Cloud as ${assumeEmail}. The session will expire after 60 minutes. Use the banner to unassume when done.`}
  onConfirm={doAssume}
/>

<GuardedDeleteDialog
  bind:open={deleteDialogOpen}
  title="Delete User"
  description={`This will permanently delete the user ${deleteEmail}. This action cannot be undone.`}
  confirmText={deleteEmail}
  confirmButtonText="Delete"
  loading={deleteLoading}
  error={deleteError}
  onConfirm={doDelete}
/>
