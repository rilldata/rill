<script lang="ts">
  import { page } from "$app/stores";
  import type { V1OrganizationMemberUser } from "@rilldata/web-admin/client";
  import {
    createAdminServiceCreateUsergroup,
    createAdminServiceAddUsergroupMemberUser,
    getAdminServiceListOrganizationMemberUsergroupsQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
    getAdminServiceListUsergroupMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import { Button } from "@rilldata/web-common/components/button/index.js";
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
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  export let open = false;
  export let groupName: string;
  export let searchUsersList: V1OrganizationMemberUser[] = [];
  export let currentUserEmail: string = "";

  let searchText = "";
  let selectedUsers: V1OrganizationMemberUser[] = [];

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const createUserGroup = createAdminServiceCreateUsergroup();
  const addUsergroupMemberUser = createAdminServiceAddUsergroupMemberUser();

  async function handleAddUsergroupMemberUser(
    email: string,
    usergroup: string,
  ) {
    try {
      await $addUsergroupMemberUser.mutateAsync({
        organization: organization,
        usergroup: usergroup,
        email: email,
        data: {},
      });

      await queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationMemberUsersQueryKey(organization),
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListUsergroupMemberUsersQueryKey(
          organization,
          usergroup,
        ),
      });

      eventBus.emit("notification", {
        message: "User added to user group",
      });
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error adding user to user group",
        type: "error",
      });
    }
  }

  async function handleCreate(newName: string) {
    try {
      await $createUserGroup.mutateAsync({
        organization: organization,
        data: {
          name: newName,
        },
      });

      // Add selected users to the newly created group
      for (const user of selectedUsers) {
        await handleAddUsergroupMemberUser(user.userEmail, newName);
      }

      await queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      });

      groupName = "";
      selectedUsers = [];
      open = false;

      eventBus.emit("notification", { message: "User group created" });
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error creating user group",
        type: "error",
      });
    }
  }

  function handleRemoveUser(email: string) {
    selectedUsers = selectedUsers.filter((user) => user.userEmail !== email);
  }

  const formId = "create-user-group-form";

  const initialValues = {
    name: "",
  };

  const schema = yup(
    object({
      name: string()
        .required("User group name is required")
        .min(3, "User group name must be at least 3 characters")
        .matches(
          /^[a-z0-9]+(-[a-z0-9]+)*$/,
          "User group name must be lowercase and can contain letters, numbers, and hyphens (slug)",
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
          await handleCreate(values.name);
          open = false;
          // eslint-disable-next-line @typescript-eslint/no-unused-vars
        } catch (error) {
          console.error(error);
        }
      },
    },
  );

  $: coercedUsersToOptions = searchUsersList
    .filter(
      (user) =>
        !selectedUsers.some(
          (selected) => selected.userEmail === user.userEmail,
        ),
    )
    .map((user) => ({
      value: user.userEmail,
      label: user.userEmail,
      name: user.userName,
    }));

  function handleClose() {
    open = false;
    searchText = "";
    selectedUsers = [];
  }
</script>

<Dialog
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    handleClose();
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
      <DialogTitle>Create a group</DialogTitle>
    </DialogHeader>
    <form
      id={formId}
      class="w-full"
      on:submit|preventDefault={submit}
      use:enhance
    >
      <div class="flex flex-col gap-4 w-full">
        <Input
          bind:value={$form.name}
          id="create-user-group-name"
          label="Name"
          placeholder="Untitled"
          errors={$errors.name}
        />

        <Combobox
          bind:inputValue={searchText}
          options={coercedUsersToOptions}
          id="user-group-users"
          label="Users"
          name="searchUsers"
          placeholder="Search for users"
          emptyText="No users found"
          onSelectedChange={(value) => {
            if (value) {
              const selectedUser = searchUsersList.find(
                (user) => user.userEmail === value.value,
              );
              if (selectedUser) {
                selectedUsers = [...selectedUsers, selectedUser];
              }
            }
          }}
        />
      </div>
    </form>

    {#if selectedUsers.length > 0}
      <div class="flex flex-col gap-2 w-full">
        <div class="flex flex-row items-center gap-x-1">
          <div class="text-xs font-semibold uppercase text-gray-500">
            {selectedUsers.length} Users
          </div>
          <Tooltip location="right" alignment="middle" distance={8}>
            <div class="text-gray-500">
              <InfoCircle size="12px" />
            </div>
            <TooltipContent maxWidth="400px" slot="tooltip-content">
              Users in this group will inherit the group's permissions.
            </TooltipContent>
          </Tooltip>
        </div>
        <div class="max-h-[208px] overflow-y-auto">
          <div class="flex flex-col gap-2">
            {#each selectedUsers as user}
              <div class="flex flex-row justify-between gap-2 items-center">
                <div class="flex items-center gap-2">
                  <Avatar avatarSize="h-7 w-7" alt={user.userName} />
                  <div class="flex flex-col text-left">
                    <span class="text-sm font-medium text-gray-900">
                      {user.userName}
                      <span class="text-gray-500 font-normal">
                        {user.userEmail === currentUserEmail ? "(You)" : ""}
                      </span>
                    </span>
                    <span class="text-xs text-gray-500">{user.userEmail}</span>
                  </div>
                </div>
                <Button
                  type="text"
                  danger
                  on:click={() => handleRemoveUser(user.userEmail)}
                >
                  Remove
                </Button>
              </div>
            {/each}
          </div>
        </div>
      </div>
    {:else}
      <div class="flex flex-col gap-2 w-full">
        <div class="text-xs font-semibold uppercase text-gray-500">Users</div>
        <div class="text-gray-500">No users selected</div>
      </div>
    {/if}

    <DialogFooter>
      <Button type="plain" on:click={handleClose}>Cancel</Button>
      <Button
        type="primary"
        disabled={$submitting || $form.name.trim() === ""}
        form={formId}
        submitForm
      >
        Create
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
