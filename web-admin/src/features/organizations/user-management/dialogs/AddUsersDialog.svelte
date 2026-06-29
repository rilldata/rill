<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import { page } from "$app/stores";
  import {
    createAdminServiceAddOrganizationMemberUser,
    getAdminServiceListOrganizationInvitesQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import { RFC5322EmailRegex } from "@rilldata/web-common/components/forms/validation.ts";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { array, object, string } from "yup";

  import { getOrgRolesOptions } from "../../constants";

  export let open = false;
  export let email: string;
  export let role: string;
  export let isSuperUser: boolean;

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const addOrganizationMemberUser =
    createAdminServiceAddOrganizationMemberUser();

  let failedInvites: string[] = [];

  async function handleCreate(
    newEmail: string,
    newRole: string,
    isSuperUser: boolean = false,
  ) {
    await $addOrganizationMemberUser.mutateAsync({
      org: organization,
      data: {
        email: newEmail,
        role: newRole,
        superuserForceAccess: isSuperUser,
      },
    });

    await queryClient.invalidateQueries({
      queryKey:
        getAdminServiceListOrganizationMemberUsersQueryKey(organization),
    });

    await queryClient.invalidateQueries({
      queryKey: getAdminServiceListOrganizationInvitesQueryKey(organization),
    });

    email = "";
    role = "";
    isSuperUser = false;
  }

  const formId = "add-user-form";

  const initialValues: {
    emails: string[];
    role: string;
  } = {
    emails: [""],
    role: OrgUserRoles.Viewer,
  };
  const schema = yup(
    object({
      emails: array(
        string().matches(RFC5322EmailRegex, {
          excludeEmptyString: true,
          message: m.users_invalid_email(),
        }),
      ), // yup's email regex is too simple
      role: string().required(),
    }),
  );

  const { form, errors, enhance, submit, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        failedInvites = [];
        let succeeded = [];
        let failed = [];

        if (!form.valid) return;
        const values = form.data;
        const emails = values.emails.map((e) => e.trim()).filter(Boolean);
        if (emails.length === 0) return;

        const results = await Promise.all(
          emails.map(async (email, index) => {
            try {
              await handleCreate(email, values.role, isSuperUser);
              return { index, email, success: true };
            } catch (error) {
              console.error("Error adding user to organization", error);
              return { index, email, success: false };
            }
          }),
        );

        results
          .sort((a, b) => a.index - b.index)
          .forEach(({ email, success }) => {
            if (success) {
              succeeded.push(email);
            } else {
              failed.push(email);
            }
          });

        // Only show success notification if any invites succeeded
        if (succeeded.length > 0) {
          eventBus.emit("notification", {
            type: "success",
            message: m.users_invited_success({ count: succeeded.length, role: values.role }),
          });
        }

        // Show error notification if any invites failed
        if (failed.length > 0) {
          failedInvites = failed; // Store failed emails
        }

        // Close dialog after showing notifications
        if (failedInvites.length === 0) {
          open = false;
        }
      },
      validationMethod: "oninput",
    },
  );

  $: orgRolesOptions = getOrgRolesOptions();

  $: hasInvalidEmails = $form.emails.some(
    (e, i) => e.length > 0 && $errors.emails?.[i] !== undefined,
  );
</script>

<Dialog
  bind:open
  onOpenChange={(open) => {
    if (!open) {
      email = "";
      role = "";
      isSuperUser = false;
      failedInvites = [];
      $form.emails = [""];
    }
  }}
>
  <DialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </DialogTrigger>
  <DialogContent
    class="translate-y-[-200px]"
    onInteractOutside={(e) => {
      e.preventDefault();
      open = false;
      email = "";
      role = "";
      isSuperUser = false;
      failedInvites = [];
      $form.emails = [""];
    }}
  >
    <DialogHeader>
      <DialogTitle>{m.users_add_users()}</DialogTitle>
    </DialogHeader>
    <form
      id={formId}
      onsubmit={(e) => {
        e.preventDefault();
        submit(e);
      }}
      class="w-full"
      use:enhance
    >
      <MultiInput
        id="emails"
        placeholder={m.users_email_placeholder()}
        contentClassName="relative [&>div:first-child]:max-h-[120px] [&>div:first-child]:overflow-y-auto"
        bind:values={$form.emails}
        errors={$errors.emails}
        singular="email"
        plural="emails"
      >
        <div slot="within-input" class="flex items-center h-full">
          <DropdownMenu>
            <DropdownMenuTrigger
              class="w-18 flex flex-row gap-1 items-center rounded-sm px-2 py-1 hover:bg-surface-hover"
            >
              <div class="text-xs">
                {orgRolesOptions.find((o) => o.value === $form.role)?.label}
              </div>
              <CaretDownIcon size="12px" />
            </DropdownMenuTrigger>
            <DropdownMenuContent
              side="bottom"
              align="end"
              class="w-[260px]"
              strategy="fixed"
            >
              {#each orgRolesOptions as { value, label, description } (value)}
                <DropdownMenuItem
                  onclick={() => ($form.role = value)}
                  class="text-xs hover:bg-surface-hover {$form.role === value
                    ? 'bg-surface-active'
                    : ''}"
                >
                  <div class="flex flex-col">
                    <div class="text-xs font-medium text-fg-primary">
                      {label}
                    </div>
                    <div class="text-fg-secondary text-[11px]">
                      {description}
                    </div>
                  </div>
                </DropdownMenuItem>
              {/each}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <svelte:fragment slot="beside-input" let:hasSomeValue>
          <Button
            submitForm
            type="primary"
            form={formId}
            loading={$submitting}
            disabled={hasInvalidEmails || !hasSomeValue}
            forcedStyle="height: 32px !important;"
          >
            {m.users_invite()}
          </Button>
        </svelte:fragment>
      </MultiInput>

      {#if failedInvites.length > 0}
        <div class="text-sm text-red-500 py-2">
          {m.users_already_member({ emails: failedInvites.join(", "), count: failedInvites.length })}
        </div>
      {/if}
    </form>
  </DialogContent>
</Dialog>
