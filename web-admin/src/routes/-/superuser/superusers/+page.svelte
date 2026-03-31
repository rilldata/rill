<script lang="ts">
  import {
    createAdminServiceListSuperusers,
    createAdminServiceSetSuperuser,
  } from "@rilldata/web-admin/client";
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
  import { useQueryClient } from "@tanstack/svelte-query";

  let newEmail = "";
  let addLoading = false;
  let dialogOpen = false;
  let dialogAction: () => Promise<void> = async () => {};
  let dialogLoading = false;
  let removeTarget = "";

  const queryClient = useQueryClient();
  const superusersQuery = createAdminServiceListSuperusers();
  const setSuperuser = createAdminServiceSetSuperuser();

  async function handleAdd() {
    if (!newEmail) return;
    addLoading = true;
    try {
      await $setSuperuser.mutateAsync({
        data: { email: newEmail, superuser: true },
      });
      eventBus.emit("notification", {
        type: "success",
        message: `${newEmail} added as superuser`,
      });
      newEmail = "";
      await queryClient.invalidateQueries({
        predicate: (q) => (q.queryKey[0] as string)?.includes("/v1/superuser"),
      });
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed: ${err}`,
      });
    } finally {
      addLoading = false;
    }
  }

  function handleRemove(email: string) {
    removeTarget = email;
    dialogAction = async () => {
      try {
        await $setSuperuser.mutateAsync({ data: { email, superuser: false } });
        eventBus.emit("notification", {
          type: "success",
          message: `${email} removed as superuser`,
        });
        await queryClient.invalidateQueries({
          predicate: (q) =>
            (q.queryKey[0] as string)?.includes("/v1/superuser"),
        });
      } catch (err) {
        eventBus.emit("notification", {
          type: "error",
          message: `Failed: ${err}`,
        });
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
      // Keep open for retry
    } finally {
      dialogLoading = false;
    }
  }
</script>

<p class="text-sm text-fg-secondary mb-4">Manage who has superuser access across all of Rill Cloud.</p>

<div class="p-5 rounded-lg border mb-6">
  <h2 class="text-sm font-semibold text-fg-primary mb-3">Add Superuser</h2>
  <div class="flex gap-3 items-center flex-wrap">
    <input
      type="email"
      class="px-3 py-2 text-sm rounded-md border bg-input text-fg-primary placeholder:text-fg-muted focus:outline-none focus:ring-2 focus:ring-primary-500"
      placeholder="Email address"
      bind:value={newEmail}
      on:keydown={(e) => e.key === "Enter" && handleAdd()}
    />
    <Button
      large
      class="font-normal"
      type="primary"
      onClick={handleAdd}
      disabled={addLoading || !newEmail}
      loading={addLoading}
    >
      Add Superuser
    </Button>
  </div>
</div>

{#if $superusersQuery.isFetching}
  <p class="text-sm text-fg-secondary py-4">Loading superusers...</p>
{:else if $superusersQuery.data?.users?.length}
  <p class="text-sm text-fg-secondary mb-2">
    {$superusersQuery.data.users.length} superuser{$superusersQuery.data.users
      .length === 1
      ? ""
      : "s"}
  </p>
  <table class="w-full">
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
          >Actions</th
        >
      </tr>
    </thead>
    <tbody>
      {#each $superusersQuery.data.users as user}
        <tr>
          <td class="px-4 py-3 text-sm font-mono text-fg-primary border-b"
            >{user.email}</td
          >
          <td class="px-4 py-3 text-sm text-fg-primary border-b"
            >{user.displayName ?? "-"}</td
          >
          <td class="px-4 py-3 text-sm text-fg-primary border-b">
            <Button
              large
              class="font-normal"
              type="secondary-destructive"
              onClick={() => handleRemove(user.email ?? "")}
            >
              Remove
            </Button>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{/if}

<AlertDialog bind:open={dialogOpen}>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Remove Superuser</AlertDialogTitle>
      <AlertDialogDescription>
        Remove superuser access for {removeTarget}? They will lose access to
        this console.
      </AlertDialogDescription>
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
        type="destructive"
        onClick={handleConfirm}
        loading={dialogLoading}>Remove</Button
      >
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
