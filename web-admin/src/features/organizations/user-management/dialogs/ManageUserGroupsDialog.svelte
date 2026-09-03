<script lang="ts">
  import {
    createAdminServiceAddUsergroupMemberUser,
    createAdminServiceListUsergroupsForOrganizationAndUser,
    createAdminServiceRemoveUsergroupMemberUser,
    getAdminServiceListUsergroupsForOrganizationAndUserQueryKey,
  } from "@rilldata/web-admin/client";
  import UserGroupsMultiSelect from "@rilldata/web-admin/features/organizations/user-management/UserGroupsMultiSelect.svelte";
  import {
    invalidateOrgInvites,
    invalidateOrgMemberUsers,
    invalidateOrgUsergroups,
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

  $: memberGroupsQuery = createAdminServiceListUsergroupsForOrganizationAndUser(
    organization,
    { userId },
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
    open = false;
    initialized = false;
    selectedGroups = [];
    initialGroups = [];
  }

  async function handleSave() {
    saving = true;
    try {
      for (const group of additions) {
        await $addUsergroupMemberUser.mutateAsync({
          org: organization,
          usergroup: group,
          email,
          data: {},
        });
      }
      for (const group of removals) {
        await $removeUsergroupMemberUser.mutateAsync({
          org: organization,
          usergroup: group,
          email,
        });
      }

      await Promise.all([
        invalidateOrgMemberUsers(queryClient, organization),
        invalidateOrgInvites(queryClient, organization),
        invalidateOrgUsergroups(queryClient, organization),
        queryClient.invalidateQueries({
          queryKey: getAdminServiceListUsergroupsForOrganizationAndUserQueryKey(
            organization,
            { userId },
          ),
        }),
      ]);

      eventBus.emit("notification", { message: m.users_groups_updated() });
      handleClose();
    } catch (error) {
      eventBus.emit("notification", {
        message: m.users_error_updating_groups({
          message: error?.response?.data?.message ?? String(error),
        }),
        type: "error",
      });
    } finally {
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
  <DialogContent class="translate-y-[-200px]" interactOutsideBehavior="ignore">
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
      <Button type="tertiary" onClick={handleClose}>{m.users_cancel()}</Button>
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
