<script lang="ts">
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog-v2";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import { createAdminServiceListUsergroupMemberUsers } from "@rilldata/web-admin/client";
  import { page } from "$app/stores";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";

  export let open = false;
  export let groupName: string;
  export let currentUserEmail: string;
  export let onRename: (groupName: string, newName: string) => void;
  export let onRemoveUser: (groupName: string, email: string) => void;

  $: organization = $page.params.organization;
  $: listUsergroupMemberUsers = createAdminServiceListUsergroupMemberUsers(
    organization,
    groupName,
  );

  const formId = "rename-user-group-form";

  const initialValues = {
    newName: groupName,
  };

  const schema = yup(
    object({
      newName: string()
        .required("New user group name is required")
        .min(3, "New user group name must be at least 3 characters")
        .matches(
          /^[a-z0-9]+(-[a-z0-9]+)*$/,
          "New user group name must be lowercase and can contain letters, numbers, and hyphens (slug)",
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
          await onRename(groupName, values.newName);
          open = false;
        } catch (error) {
          console.error(error);
        }
      },
    },
  );
</script>

<Dialog
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    open = false;
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
      <div class="flex flex-col gap-2 w-full">
        <Input
          bind:value={$form.newName}
          placeholder="New user group name"
          errors={$errors.newName}
          alwaysShowError
        />
      </div>
    </form>
    {#if $listUsergroupMemberUsers.data?.members.length > 0}
      <div class="flex flex-col gap-2 w-full">
        <div class="flex flex-row items-center gap-x-1">
          <div class="text-xs font-semibold uppercase text-gray-500">Users</div>
          <Tooltip location="right" alignment="middle" distance={8}>
            <div class="text-gray-500">
              <InfoCircle size="12px" />
            </div>
            <TooltipContent maxWidth="400px" slot="tooltip-content">
              Users in this group will inherit the group's permissions.
            </TooltipContent>
          </Tooltip>
        </div>
        <div class="flex flex-col gap-2">
          {#each $listUsergroupMemberUsers.data?.members as member}
            <div class="flex flex-row justify-between gap-2 items-center">
              <div class="flex items-center gap-2">
                <Avatar size="h-7 w-7" alt={member.userName} />
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
                type="plain"
                on:click={() => {
                  onRemoveUser(groupName, member.userEmail);
                }}
              >
                Remove
              </Button>
            </div>
          {/each}
        </div>
      </div>
    {:else}
      <div class="flex flex-col gap-2 w-full">
        <div class="text-xs font-semibold uppercase text-gray-500">Users</div>
        <div class="text-gray-500">No users in this group</div>
      </div>
    {/if}
    <DialogFooter>
      <Button
        type="plain"
        on:click={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button
        type="primary"
        disabled={$submitting || $form.newName.trim() === groupName}
        form={formId}
        submitForm
      >
        Save
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
