<script lang="ts">
  import { page } from "$app/stores";
  import type { V1OrganizationMemberUser } from "@rilldata/web-admin/client";
  import {
    createAdminServiceAddUsergroupMemberUser,
    createAdminServiceListUsergroupMemberUsers,
    createAdminServiceRemoveUsergroupMemberUser,
    createAdminServiceRenameUsergroup,
    getAdminServiceListOrganizationMemberUsergroupsQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
    getAdminServiceListUsergroupMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
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
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  export let open = false;
  export let groupName: string;
  export let currentUserEmail: string;
  export let searchUsersList: V1OrganizationMemberUser[];

  let searchText = "";

  $: organization = $page.params.organization;
  $: listUsergroupMemberUsers = createAdminServiceListUsergroupMemberUsers(
    organization,
    groupName,
  );

  const queryClient = useQueryClient();
  const removeUserGroupMember = createAdminServiceRemoveUsergroupMemberUser();
  const addUsergroupMemberUser = createAdminServiceAddUsergroupMemberUser();
  const renameUserGroup = createAdminServiceRenameUsergroup();

  let pendingAdditions: string[] = [];
  let pendingRemovals: string[] = [];

  async function handleAddUsergroupMemberUser(email: string) {
    pendingAdditions = [...pendingAdditions, email];
    pendingRemovals = pendingRemovals.filter((e) => e !== email);
  }

  async function handleRename(groupName: string, newName: string) {
    try {
      await $renameUserGroup.mutateAsync({
        organization: organization,
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

  async function handleRemoveUser(groupName: string, email: string) {
    pendingRemovals = [...pendingRemovals, email];
    pendingAdditions = pendingAdditions.filter((e) => e !== email);
  }

  async function applyPendingChanges() {
    try {
      for (const email of pendingAdditions) {
        await $addUsergroupMemberUser.mutateAsync({
          organization: organization,
          usergroup: groupName,
          email: email,
          data: {},
        });
      }

      for (const email of pendingRemovals) {
        await $removeUserGroupMember.mutateAsync({
          organization: organization,
          usergroup: groupName,
          email: email,
        });
      }

      await queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationMemberUsersQueryKey(organization),
      });

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
        .min(3, "Name must be at least 3 characters")
        .matches(
          /^[a-z0-9]+(-[a-z0-9]+)*$/,
          "Name must be lowercase and can contain letters, numbers, and hyphens (slug)",
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

  $: availableSearchUsersList = searchUsersList.filter(
    (user) =>
      !$listUsergroupMemberUsers.data?.members.some(
        (member) => member.userEmail === user.userEmail,
      ) && !pendingAdditions.includes(user.userEmail),
  );

  $: displayedMembers = [
    ...($listUsergroupMemberUsers.data?.members.filter(
      (member) => !pendingRemovals.includes(member.userEmail),
    ) || []),
    ...pendingAdditions.map((email) => ({
      userEmail: email,
      userName:
        searchUsersList.find((u) => u.userEmail === email)?.userName || email,
    })),
  ];

  $: coercedUsersToOptions = availableSearchUsersList.map((user) => ({
    value: user.userEmail,
    label: user.userEmail,
    name: user.userName,
  }));

  function handleClose() {
    open = false;
    searchText = "";
    pendingAdditions = [];
    pendingRemovals = [];
  }
</script>

<Dialog
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    handleClose();
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
          placeholder="New name"
          id="user-group-name"
          label="Name"
          errors={$errors.newName}
          alwaysShowError
        />

        <div class="flex flex-col gap-y-1">
          <label
            for="user-group-users"
            class="line-clamp-1 text-sm font-medium text-gray-800"
          >
            Users
          </label>
          <Combobox
            bind:inputValue={searchText}
            options={coercedUsersToOptions}
            placeholder="Search for users"
            onSelectedChange={(value) => {
              if (value) {
                handleAddUsergroupMemberUser(value.value);
              }
            }}
          />
        </div>
      </div>
    </form>

    <div class="flex flex-col gap-2 w-full">
      {#if displayedMembers.length > 0}
        <div class="flex flex-row items-center gap-x-1">
          <div class="text-xs font-semibold uppercase text-gray-500">
            {displayedMembers.length} User{displayedMembers.length === 1
              ? ""
              : "s"}
          </div>
        </div>
      {/if}
      <div class="max-h-[208px] overflow-y-auto">
        <div class="flex flex-col gap-2">
          {#each displayedMembers as member}
            <div class="flex flex-row justify-between gap-2 items-center">
              <div class="flex items-center gap-2">
                <Avatar avatarSize="h-7 w-7" alt={member.userName} />
                <div class="flex flex-col text-left">
                  <span class="text-sm font-medium text-gray-900">
                    {member.userName}
                    <span class="text-gray-500 font-normal">
                      {member.userEmail === currentUserEmail ? "(You)" : ""}
                    </span>
                  </span>
                  <span class="text-xs text-gray-500">{member.userEmail}</span>
                </div>
              </div>
              <Button
                type="text"
                danger
                on:click={() => {
                  handleRemoveUser(groupName, member.userEmail);
                }}
              >
                Remove
              </Button>
            </div>
          {/each}
        </div>
      </div>
    </div>
    <DialogFooter>
      <Button type="plain" on:click={handleClose}>Cancel</Button>
      <Button
        type="primary"
        disabled={$submitting ||
          ($form.newName.trim() === groupName &&
            pendingAdditions.length === 0 &&
            pendingRemovals.length === 0)}
        form={formId}
        submitForm
      >
        Save
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
