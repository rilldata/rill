<script lang="ts">
  import SearchInput from "@rilldata/web-admin/features/superuser/shared/SearchInput.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    searchUsers,
    createDeleteUserMutation,
  } from "@rilldata/web-admin/features/superuser/users/selectors";
  import { assumedUser } from "@rilldata/web-admin/features/superuser/users/assume-state";
  import { useQueryClient } from "@tanstack/svelte-query";

  let searchQuery = "";
  let dialogOpen = false;
  let dialogTitle = "";
  let dialogDescription = "";
  let dialogDestructive = false;
  let dialogAction: () => Promise<void> = async () => {};
  let dialogLoading = false;
  let actionInProgress = "";

  const queryClient = useQueryClient();
  const deleteUser = createDeleteUserMutation();

  $: usersQuery = searchUsers(searchQuery);

  function handleSearch(e: CustomEvent<string>) {
    searchQuery = e.detail;
  }

  function handleAssume(email: string) {
    dialogTitle = "Open as User";
    dialogDescription = `You will start browsing Rill Cloud as ${email}. The session will expire after 60 minutes. Use the banner to unassume when done.`;
    dialogDestructive = false;
    dialogAction = async () => {
      assumedUser.assume(email, {});
    };
    dialogOpen = true;
  }

  function handleUnassume() {
    assumedUser.unassume();
  }

  function handleDelete(email: string) {
    dialogTitle = "Delete User";
    dialogDescription = `This will permanently delete the user ${email}. This action cannot be undone.`;
    dialogDestructive = true;
    dialogAction = async () => {
      actionInProgress = `delete:${email}`;
      try {
        await $deleteUser.mutateAsync({ email });
        eventBus.emit("notification", {
          type: "success",
          message: `User ${email} deleted`,
        });
        await queryClient.invalidateQueries({
          predicate: (q) => q.queryKey[0] === "/v1/users/search",
        });
      } catch (err) {
        eventBus.emit("notification", {
          type: "error",
          message: `Failed to delete user: ${err}`,
        });
      } finally {
        actionInProgress = "";
      }
    };
    dialogOpen = true;
  }

  async function handleConfirm() {
    dialogLoading = true;
    try {
      await dialogAction();
      dialogOpen = false;
    } catch {
      // Keep dialog open for retry
    } finally {
      dialogLoading = false;
    }
  }
</script>

<p class="text-sm text-fg-secondary mb-4">Search for users by email across all organizations.</p>

{#if $assumedUser}
  <div
    class="flex items-center gap-3 mb-4 px-4 py-2 rounded-md bg-yellow-100 border border-yellow-300 text-yellow-800 text-sm"
  >
    <span>Currently assumed as <strong>{$assumedUser}</strong></span>
    <Button large class="font-normal" type="tertiary" onClick={handleUnassume}
      >Unassume</Button
    >
  </div>
{/if}

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
                    onClick={() => handleAssume(user.email ?? "")}
                    >Open as User</Button
                  >
                {/if}
                <Button
                  large
                  class="font-normal"
                  type="secondary-destructive"
                  disabled={actionInProgress === `delete:${user.email}`}
                  loading={actionInProgress === `delete:${user.email}`}
                  onClick={() => handleDelete(user.email ?? "")}
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

<AlertDialog bind:open={dialogOpen}>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{dialogTitle}</AlertDialogTitle>
      <AlertDialogDescription>{dialogDescription}</AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        large
        class="font-normal"
        type="tertiary"
        onClick={() => (dialogOpen = false)}>Cancel</Button
      >
      <Button
        large
        class="font-normal"
        type={dialogDestructive ? "destructive" : "primary"}
        onClick={handleConfirm}
        loading={dialogLoading}
      >
        Confirm
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
