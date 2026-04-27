<script lang="ts">
  import AssumeUserDialog from "@rilldata/web-admin/features/superuser/dialogs/AssumeUserDialog.svelte";
  import DeleteUserDialog from "@rilldata/web-admin/features/superuser/dialogs/DeleteUserDialog.svelte";
  import SearchInput from "@rilldata/web-admin/features/superuser/shared/SearchInput.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { searchUsers } from "@rilldata/web-admin/features/superuser/users/selectors";
  import { assumedUser } from "@rilldata/web-admin/features/superuser/users/assume-state";

  let searchQuery = "";

  let assumeDialogOpen = false;
  let assumeEmail = "";

  let deleteDialogOpen = false;
  let deleteEmail = "";

  $: usersQuery = searchUsers(searchQuery);

  function handleSearch(e: CustomEvent<string>) {
    searchQuery = e.detail;
  }
</script>

<h1 class="text-lg font-semibold text-fg-primary">Users</h1>
<p class="text-sm text-fg-secondary mb-4">
  Search for users by email across all organizations.
</p>

<div class="mb-4 max-w-md">
  <SearchInput
    placeholder="Search by email (min 3 characters)..."
    on:search={handleSearch}
  />
  {#if searchQuery.length < 3}
    <p class="text-sm text-fg-muted mt-2">
      Type at least 3 characters to search across all organizations.
    </p>
  {/if}
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
                    onClick={() => assumedUser.unassume()}>Unassume</Button
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
{/if}

<AssumeUserDialog bind:open={assumeDialogOpen} email={assumeEmail} />
<DeleteUserDialog bind:open={deleteDialogOpen} email={deleteEmail} />
