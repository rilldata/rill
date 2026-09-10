<script lang="ts">
  import {
    createAdminServiceAddUsergroupMemberUser,
    createAdminServiceListUsergroupsForOrganizationAndUser,
    createAdminServiceRemoveUsergroupMemberUser,
    getAdminServiceListUsergroupMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import UserGroupsMultiSelect from "@rilldata/web-admin/features/organizations/user-management/UserGroupsMultiSelect.svelte";
  import {
    invalidateOrgInvites,
    invalidateOrgMemberUsers,
    invalidateOrgUsergroups,
    invalidateUserGroupsForUser,
  } from "@rilldata/web-admin/features/organizations/user-management/utils.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useQueryClient } from "@tanstack/svelte-query";

  // Edits the user groups of one org user, whether they are a member or still have a pending invite.
  export let open = false;
  export let organization: string;
  export let email: string;
  // Empty for a pending invitee, who has no user yet
  export let userId: string = "";
  export let pendingAcceptance = false;
  // For a pending invitee, the groups stored on the invite; ignored for members, whose groups are fetched
  export let currentGroups: string[] = [];

  let selectedGroups: string[] = [];
  let initialGroups: string[] = [];
  let initialized = false;
  let saving = false;

  const queryClient = useQueryClient();
  const addUsergroupMemberUser = createAdminServiceAddUsergroupMemberUser();
  const removeUsergroupMemberUser =
    createAdminServiceRemoveUsergroupMemberUser();

  // The dialog does not paginate, so ask for every group in one page.
  // The server's default of 20 would silently truncate the baseline and make those groups unremovable here.
  $: listParams = { userId, pageSize: 1000 };

  $: memberGroupsQuery = createAdminServiceListUsergroupsForOrganizationAndUser(
    organization,
    listParams,
    { query: { enabled: open && !pendingAcceptance && !!userId } },
  );

  // Seed the selection once per opening, from the invite for pending users and from the query for members
  $: if (open && !initialized) {
    if (pendingAcceptance) {
      initialGroups = [...currentGroups];
      selectedGroups = [...currentGroups];
      initialized = true;
    } else if ($memberGroupsQuery.data) {
      initialGroups =
        $memberGroupsQuery.data.usergroups
          ?.filter((g) => !g.managed)
          .map((g) => g.groupName ?? "")
          .filter(Boolean) ?? [];
      selectedGroups = [...initialGroups];
      initialized = true;
    }
  }

  $: isLoading = open && !pendingAcceptance && $memberGroupsQuery.isLoading;

  $: additions = selectedGroups.filter((g) => !initialGroups.includes(g));
  $: removals = initialGroups.filter((g) => !selectedGroups.includes(g));
  $: hasChanges = additions.length > 0 || removals.length > 0;

  function handleClose() {
    // The dialog is locked while saving (see the markup), so this only guards direct calls
    if (saving) return;
    open = false;
    initialized = false;
    selectedGroups = [];
    initialGroups = [];
  }

  async function handleSave() {
    saving = true;
    // Snapshot the target and the changes before the first await:
    // the props and the derived additions/removals are reactive, so if they changed mid-save
    // the remaining requests would be sent for a different user or a different selection.
    const targetEmail = email;
    const targetUserId = userId;
    const toAdd = [...additions];
    const toRemove = [...removals];
    // The changes are applied one at a time, so track what actually landed:
    // on a partial failure the baseline has to move with them, otherwise a retry
    // resends a committed addition and the server rejects it as a duplicate.
    const added: string[] = [];
    const removed: string[] = [];
    try {
      for (const group of toAdd) {
        await $addUsergroupMemberUser.mutateAsync({
          org: organization,
          usergroup: group,
          email: targetEmail,
          data: {},
        });
        added.push(group);
      }
      for (const group of toRemove) {
        await $removeUsergroupMemberUser.mutateAsync({
          org: organization,
          usergroup: group,
          email: targetEmail,
        });
        removed.push(group);
      }

      eventBus.emit("notification", { message: m.users_groups_updated() });
      saving = false;
      handleClose();
    } catch (error) {
      initialGroups = [
        ...initialGroups.filter((g) => !removed.includes(g)),
        ...added,
      ];
      eventBus.emit("notification", {
        message: m.users_error_updating_groups({
          message: error?.response?.data?.message ?? String(error),
        }),
        type: "error",
      });
    } finally {
      if (added.length > 0 || removed.length > 0) {
        await Promise.all([
          invalidateOrgMemberUsers(queryClient, organization),
          invalidateOrgInvites(queryClient, organization),
          invalidateOrgUsergroups(queryClient, organization),
          invalidateUserGroupsForUser(queryClient, organization, targetUserId),
          // The member lists of the groups that changed, as shown in the group dialogs
          ...[...added, ...removed].map((group) =>
            queryClient.invalidateQueries({
              queryKey: getAdminServiceListUsergroupMemberUsersQueryKey(
                organization,
                group,
              ),
            }),
          ),
        ]);
      }
      saving = false;
    }
  }
</script>

<Dialog
  bind:open
  onOpenChange={(dialogOpen) => {
    if (!dialogOpen) handleClose();
  }}
>
  <DialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </DialogTrigger>
  <!-- Lock the dialog while saving: closing it would let the parent point it at another user before the requests finish -->
  <DialogContent
    class="translate-y-[-200px]"
    interactOutsideBehavior="ignore"
    escapeKeydownBehavior={saving ? "ignore" : "close"}
    noClose={saving}
  >
    <DialogHeader>
      <DialogTitle>{m.users_manage_groups()}</DialogTitle>
    </DialogHeader>
    <DialogDescription>
      {m.users_manage_groups_description({ email })}
    </DialogDescription>

    {#if pendingAcceptance}
      <div class="text-xs text-fg-secondary">
        {m.users_manage_groups_pending_hint()}
      </div>
    {/if}

    {#if isLoading}
      <DelayedSpinner isLoading size="1rem" />
    {:else if $memberGroupsQuery.isError && !pendingAcceptance}
      <div class="text-xs text-red-500">{m.users_error()}</div>
    {:else}
      <div class="flex flex-col gap-y-1">
        <label for="manage-usergroups" class="text-xs font-medium">
          {m.users_user_groups()}
        </label>
        <UserGroupsMultiSelect
          id="manage-usergroups"
          {organization}
          disabled={saving}
          bind:selected={selectedGroups}
        />
      </div>

      {#if selectedGroups.length > 0}
        <ul class="flex flex-col gap-1 max-h-[208px] overflow-y-auto">
          {#each selectedGroups as group (group)}
            <li class="flex flex-row justify-between items-center text-sm">
              <span class="truncate" title={group}>{group}</span>
              <Button
                type="destructive"
                disabled={saving}
                onClick={() =>
                  (selectedGroups = selectedGroups.filter((g) => g !== group))}
              >
                {m.users_remove()}
              </Button>
            </li>
          {/each}
        </ul>
      {:else}
        <div class="text-xs text-fg-secondary">{m.users_no_groups()}</div>
      {/if}
    {/if}

    <DialogFooter>
      <Button type="tertiary" disabled={saving} onClick={handleClose}>
        {m.users_cancel()}
      </Button>
      <Button
        type="primary"
        disabled={saving || isLoading || !hasChanges}
        loading={saving}
        onClick={handleSave}
      >
        {m.users_save()}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
