<script lang="ts">
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog-v2";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import { page } from "$app/stores";
  import { createAdminServiceListUsergroupMemberUsers } from "@rilldata/web-admin/client";
  import type { V1MemberUser } from "@rilldata/web-admin/client";
  import Combobox from "@rilldata/web-common/components/combobox/Combobox.svelte";

  export let open = false;
  export let groupName: string;
  export let searchUsersList: V1MemberUser[];
  export let onCreate: (name: string) => void;
  export let onAddUsergroupMemberUser: (
    email: string,
    usergroup: string,
  ) => void;

  let searchText = "";

  // TODO: if user is already in group, disable the option
  // Otherwise, "user is already a member of the usergroup"
  // $: console.log("yes we can search users now: ", searchUsersList);

  $: organization = $page.params.organization;
  $: listUsergroupMemberUsers = createAdminServiceListUsergroupMemberUsers(
    organization,
    groupName,
  );

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
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        try {
          await onCreate(values.name);
          open = false;
        } catch (error) {
          console.error(error);
        }
      },
    },
  );

  // TODO: rework this to a list view
  $: coercedUsersToOptions = searchUsersList.map((user) => ({
    value: user.userEmail,
    label: user.userEmail,
  }));
</script>

<Dialog
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    open = false;
    groupName = "";
  }}
  onOpenChange={(open) => {
    if (!open) {
      groupName = "";
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
          label="Group label"
          placeholder="Untitled"
          errors={$errors.name}
          alwaysShowError
        />

        <!-- TODO: revisit; what happens when users are added but groupName is empty? -->
        <!-- NOTE: Client side search current users until we have a different endpoint to search users server-side -->
        <Combobox
          bind:inputValue={searchText}
          options={coercedUsersToOptions}
          id="create-user-group-users"
          label="Users"
          name="searchUsers"
          placeholder="Search for users"
          onSelectedChange={(value) => {
            if (value) {
              onAddUsergroupMemberUser(value.value, groupName);
            }
          }}
        />

        <!-- TODO: List users, remove -->
        {#if $listUsergroupMemberUsers.data?.members.length > 0}
          <div class="text-xs font-semibold uppercase text-gray-500">
            {$listUsergroupMemberUsers.data?.members.length} Users
          </div>
        {/if}
      </div>
    </form>
    <DialogFooter>
      <Button
        type="plain"
        on:click={() => {
          open = false;
        }}>Cancel</Button
      >
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
