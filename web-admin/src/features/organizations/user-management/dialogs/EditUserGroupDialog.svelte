<script lang="ts">
  import { page } from "$app/stores";
  import type { V1OrganizationMemberUser } from "@rilldata/web-admin/client";
  import {
    createAdminServiceAddUsergroupMemberUser,
    createAdminServiceListOrganizationMemberUsers,
    createAdminServiceListUsergroupMemberUsers,
    createAdminServiceRemoveUsergroupMemberUser,
    createAdminServiceRenameUsergroup,
    getAdminServiceListOrganizationMemberUsergroupsQueryKey,
    getAdminServiceListUsergroupMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import AvatarListItem from "@rilldata/web-common/components/avatar/AvatarListItem.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import Combobox from "@rilldata/web-common/components/combobox/Combobox.svelte";
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import { SLUG_REGEX } from "@rilldata/web-admin/features/organizations/user-management/constants.ts";

  export let open = false;
  export let groupName: string;
  export let currentUserEmail: string = "";

  let searchInput = "";
  let debouncedSearchText = "";
  let debounceTimer: ReturnType<typeof setTimeout> | undefined;
  let selectedUsers: V1OrganizationMemberUser[] = [];
  let pendingAdditions: string[] = [];
  let pendingRemovals: string[] = [];
  let initialized = false;

  // Debounce search input to avoid too many API calls
  $: {
    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      debouncedSearchText = searchInput;
    }, 300);
  }

  $: organization = $page.params.organization;
  $: listUsergroupMemberUsers = createAdminServiceListUsergroupMemberUsers(
    organization,
    groupName,
  );

  $: userGroupMembersUsers = $listUsergroupMemberUsers.data?.members ?? [];

  // Query organization users when user types (debounced)
  // Use a more stable pattern - create query once and let params drive it
  $: organizationUsersQuery = createAdminServiceListOrganizationMemberUsers(
    organization,
    debouncedSearchText
      ? {
          pageSize: 50,
          searchPattern: `${debouncedSearchText}%`,
        }
      : { pageSize: 50 }, // Pass params even when no search to keep query stable
    {
      query: {
        enabled: debouncedSearchText.length > 0,
      },
    },
  );

  $: allOrganizationUsers =
    $organizationUsersQuery.data?.members?.filter(
      (u) =>
        !selectedUsers.some((selected) => selected.userEmail === u.userEmail),
    ) ?? [];

  $: if (
    userGroupMembersUsers.length > 0 &&
    selectedUsers.length === 0 &&
    !initialized
  ) {
    selectedUsers = [...userGroupMembersUsers];
    initialized = true;
  }

  // TODO: we need to get role from a separate query and fill in selectedUsers
  //       organizationUsers is not guaranteed to have the user present in the group.

  const queryClient = useQueryClient();
  const addUsergroupMemberUser = createAdminServiceAddUsergroupMemberUser();
  const removeUserGroupMember = createAdminServiceRemoveUsergroupMemberUser();
  const renameUserGroup = createAdminServiceRenameUsergroup();

  function handleRemove(email: string) {
    selectedUsers = selectedUsers.filter((user) => user.userEmail !== email);
    pendingRemovals = [...pendingRemovals, email];
    pendingAdditions = pendingAdditions.filter((e) => e !== email);
  }

  async function handleRename(groupName: string, newName: string) {
    try {
      await $renameUserGroup.mutateAsync({
        org: organization,
        usergroup: groupName,
        data: {
          name: newName,
        },
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListOrganizationMemberUsergroupsQueryKey(
          organization,
          {
            includeCounts: true,
          },
        ),
      });

      eventBus.emit("notification", { message: "User group renamed" });
    } catch (error) {
      eventBus.emit("notification", {
        message: `Error: ${error.response.data.message}`,
        type: "error",
      });
    }
  }

  async function handleAdd(email: string) {
    const user = allOrganizationUsers.find((u) => u.userEmail === email);

    // Don't add if already in selectedUsers
    if (
      user &&
      !selectedUsers.some((selected) => selected.userEmail === email)
    ) {
      selectedUsers = [...selectedUsers, user];
      pendingAdditions = [...pendingAdditions, email];
      pendingRemovals = pendingRemovals.filter((e) => e !== email);
    }
  }

  async function applyPendingChanges() {
    try {
      for (const email of pendingAdditions) {
        await $addUsergroupMemberUser.mutateAsync({
          org: organization,
          usergroup: groupName,
          email: email,
          data: {},
        });
      }

      for (const email of pendingRemovals) {
        await $removeUserGroupMember.mutateAsync({
          org: organization,
          usergroup: groupName,
          email: email,
        });
      }

      // Invalidate only the necessary queries
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListUsergroupMemberUsersQueryKey(
          organization,
          groupName,
        ),
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListOrganizationMemberUsergroupsQueryKey(
          organization,
          {
            includeCounts: true,
          },
        ),
      });

      pendingAdditions = [];
      pendingRemovals = [];

      eventBus.emit("notification", {
        message: "User group changes saved successfully",
      });
    } catch (error) {
      eventBus.emit("notification", {
        message: `Error: ${error.response.data.message}`,
        type: "error",
      });
    }
  }

  const formId = "edit-user-group-form";

  const initialValues = {
    newName: groupName,
  };

  const schema = yup(
    object({
      newName: string()
        .required("Name is required")
        .min(1, "Name must be at least 1 character")
        .max(40, "Name must be at most 40 characters")
        .matches(
          SLUG_REGEX,
          "Name can only include letters, numbers, underscores, and hyphens — no spaces or special characters",
        ),
    }),
  );

  const { form, enhance, submit, errors, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      validationMethod: "oninput",
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        try {
          await handleRename(groupName, values.newName);
          await applyPendingChanges();
          open = false;
        } catch (error) {
          console.error(error);
        }
      },
    },
  );

  $: coercedUsersToOptions = [
    ...selectedUsers.map((user) => ({
      value: user.userEmail,
      label: user.userName,
    })),
    ...allOrganizationUsers.map((user) => ({
      value: user.userEmail,
      label: user.userName,
    })),
  ];

  function getMetadata(email: string) {
    const user =
      selectedUsers.find((u) => u.userEmail === email) ||
      allOrganizationUsers.find((u) => u.userEmail === email);
    return user
      ? { name: user.userName, photoUrl: user.userPhotoUrl }
      : undefined;
  }

  // Check if form has been modified
  $: hasFormChanges = $form.newName !== initialValues.newName;

  function handleClose() {
    open = false;
    searchInput = "";
    selectedUsers = [];
    pendingAdditions = [];
    pendingRemovals = [];
    initialized = false;
    // Only reset the form if it has been modified
    if (hasFormChanges) {
      $form.newName = initialValues.newName;
    }
  }
</script>

<Dialog
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
  }}
  onOpenChange={(open) => {
    if (!open) {
      handleClose();
    }
  }}
>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Edit group</DialogTitle>
    </DialogHeader>
    <form
      id={formId}
      class="w-full"
      on:submit|preventDefault={submit}
      use:enhance
    >
      <div class="flex flex-col gap-4 w-full">
        <Input
          bind:value={$form.newName}
          id="edit-user-group-name"
          label="Name"
          placeholder="Untitled"
          errors={$errors.newName}
          alwaysShowError={true}
        />

        <div class="flex flex-col gap-y-1">
          <label
            for="user-group-users"
            class="line-clamp-1 text-sm font-medium text-gray-800"
          >
            Users
          </label>
          <Combobox
            bind:searchValue={searchInput}
            options={coercedUsersToOptions}
            placeholder="Search to add/remove users"
            {getMetadata}
            enableClientFiltering={false}
            selectedValues={[
              ...new Set(
                [
                  ...selectedUsers.map((user) => user.userEmail),
                  ...pendingAdditions,
                ].filter((email) => !pendingRemovals.includes(email)),
              ),
            ]}
            onSelectedChange={(values) => {
              if (!values) return;

              const newEmails = values.map((v) => v.value);
              const currentEmails = selectedUsers.map((u) => u.userEmail);

              // Find emails to add (in new but not in current)
              newEmails
                .filter((email) => !currentEmails.includes(email))
                .forEach((email) => handleAdd(email));

              // Find emails to remove (in current but not in new)
              currentEmails
                .filter((email) => !newEmails.includes(email))
                .forEach((email) => handleRemove(email));
            }}
          />
        </div>
      </div>
    </form>

    <div class="flex flex-col gap-2 w-full">
      {#if selectedUsers.length > 0}
        <div class="flex flex-row items-center gap-x-1">
          <div class="text-xs font-semibold uppercase text-gray-500">
            {selectedUsers.length} User{selectedUsers.length === 1 ? "" : "s"}
          </div>
        </div>
      {/if}
      <div class="max-h-[208px] overflow-y-auto">
        <div class="flex flex-col gap-2">
          {#each selectedUsers as user (user.userEmail)}
            <div class="flex flex-row justify-between gap-2 items-center">
              <AvatarListItem
                name={user.userName}
                email={user.userEmail}
                photoUrl={user.userPhotoUrl}
                isCurrentUser={user.userEmail === currentUserEmail}
                role={user.roleName}
              />
              <Button
                type="text"
                danger
                onClick={() => handleRemove(user.userEmail)}
              >
                Remove
              </Button>
            </div>
          {/each}
        </div>
      </div>
    </div>

    <DialogFooter>
      <Button type="plain" onClick={handleClose}>Cancel</Button>
      <Button
        type="primary"
        disabled={$submitting ||
          $form.newName.trim() === "" ||
          !!$errors.newName}
        form={formId}
        submitForm
      >
        Save
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
