<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { page } from "$app/stores";
  import { createAdminServiceAddOrganizationMemberUser } from "@rilldata/web-admin/client";
  import {
    buildInviteAttributes,
    invalidateOrgInvites,
    invalidateOrgMemberUsers,
  } from "@rilldata/web-admin/features/organizations/user-management/utils";
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
  import KeyValueInput from "@rilldata/web-common/components/forms/KeyValueInput.svelte";
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
  let showAttributes = false;

  async function handleCreate(
    newEmail: string,
    newRole: string,
    attributes: Record<string, string> | undefined,
    isSuperUser: boolean = false,
  ) {
    await $addOrganizationMemberUser.mutateAsync({
      org: organization,
      data: {
        email: newEmail,
        role: newRole,
        attributes,
        superuserForceAccess: isSuperUser,
      },
    });

    await invalidateOrgMemberUsers(queryClient, organization);

    await invalidateOrgInvites(queryClient, organization);

    email = "";
    role = "";
    isSuperUser = false;
  }

  const formId = "add-user-form";

  const initialValues: {
    emails: string[];
    role: string;
    attributes: Array<{ key: string; value: string }>;
  } = {
    emails: [""],
    role: OrgUserRoles.Viewer,
    attributes: [],
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
      attributes: array(
        object({
          key: string().defined(),
          value: string().defined(),
        }),
      ),
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

        // The same attributes apply to every invited email
        const attributes = buildInviteAttributes(values.attributes);

        const results = await Promise.all(
          emails.map(async (email, index) => {
            try {
              await handleCreate(email, values.role, attributes, isSuperUser);
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
            message: m.users_org_invited_success({
              count: succeeded.length,
              role: values.role,
            }),
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
      $form.attributes = [];
      showAttributes = false;
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
      $form.attributes = [];
      showAttributes = false;
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

      <div class="mt-3 flex flex-col gap-y-1">
        <button
          type="button"
          class="flex items-center gap-x-1 text-xs font-medium text-fg-secondary hover:text-fg-primary w-fit"
          onclick={() => (showAttributes = !showAttributes)}
        >
          <CaretDownIcon
            size="12px"
            className={showAttributes ? "" : "-rotate-90"}
          />
          {m.users_custom_attributes()}
        </button>
        {#if showAttributes}
          <div class="text-[11px] text-fg-secondary">
            {m.users_custom_attributes_hint()}
          </div>
          <KeyValueInput
            id="invite-attributes"
            bind:value={$form.attributes}
            keyPlaceholder={m.users_attribute_key_placeholder()}
            valuePlaceholder={m.users_attribute_value_placeholder()}
            itemLabel={m.users_custom_attributes()}
            addLabel={m.users_add_attribute()}
          />
        {/if}
      </div>

      {#if failedInvites.length > 0}
        <div class="text-sm text-red-500 py-2">
          {m.users_already_member({
            emails: failedInvites.join(", "),
            count: failedInvites.length,
          })}
        </div>
      {/if}
    </form>
  </DialogContent>
</Dialog>
