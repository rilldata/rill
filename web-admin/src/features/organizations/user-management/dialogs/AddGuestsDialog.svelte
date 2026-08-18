<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { page } from "$app/stores";
  import {
    createAdminServiceAddProjectMemberUser,
    type V1Project,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { getOrgRolesOptions } from "@rilldata/web-admin/features/organizations/constants";
  import {
    buildInviteAttributes,
    invalidateOrgInvites,
    invalidateOrgMemberUsers,
    type AttributeRow,
  } from "@rilldata/web-admin/features/organizations/user-management/utils";
  import { listProjectsForOrgQueryOptions } from "@rilldata/web-admin/features/projects/list-projects-query-options";
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
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import KeyValueInput from "@rilldata/web-common/components/forms/KeyValueInput.svelte";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import { RFC5322EmailRegex } from "@rilldata/web-common/components/forms/validation";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { createQuery, useQueryClient } from "@tanstack/svelte-query";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { array, mixed, object, string } from "yup";

  export let open = false;

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const addProjectMemberUser = createAdminServiceAddProjectMemberUser();

  let failedInvites: string[] = [];
  let selectedProjects: string[] = [];
  let projectDropdownOpen = false;
  let selectedRole: OrgUserRoles = OrgUserRoles.Viewer;
  let roleDropdownOpen = false;
  let hasAutoSelectedProject = false;
  let showAttributes = false;

  function resetDialogState() {
    failedInvites = [];
    selectedProjects = [];
    selectedRole = OrgUserRoles.Viewer;
    hasAutoSelectedProject = false;
    projectDropdownOpen = false;
    roleDropdownOpen = false;
    showAttributes = false;
    $form.attributes = [];
  }

  // Projects list
  $: projectsQuery = createQuery({
    ...listProjectsForOrgQueryOptions(organization),
    refetchOnWindowFocus: true,
  });
  $: projects = $projectsQuery?.data?.projects ?? ([] as V1Project[]);
  $: projectsErrorMessage =
    getRpcErrorMessage($projectsQuery?.error) ?? $projectsQuery?.error?.message;

  $: orgRolesOptions = getOrgRolesOptions();
  $: selectedRoleLabel =
    orgRolesOptions.find((o) => o.value === selectedRole)?.label ?? "";

  $: selectedProjectsLabel = (() => {
    if (selectedProjects.length === 0) return m.users_select_projects();
    if (selectedProjects.length === 1) return selectedProjects[0];
    return m.users_project_count({ count: selectedProjects.length });
  })();

  $: if (open && !hasAutoSelectedProject && projects.length > 0) {
    const firstProjectName = projects[0]?.name;
    if (firstProjectName) {
      selectedProjects = [firstProjectName];
      hasAutoSelectedProject = true;
    }
  }

  // Collapsing the section reads as discarding what was typed there, so drop the rows
  // rather than silently applying them to every invite.
  function toggleAttributes() {
    showAttributes = !showAttributes;
    if (!showAttributes) $form.attributes = [];
  }

  function toggleProjectSelection(projectName: string) {
    const idx = selectedProjects.indexOf(projectName);
    if (idx >= 0) {
      selectedProjects = selectedProjects.filter(
        (name) => name !== projectName,
      );
    } else {
      selectedProjects = [...selectedProjects, projectName];
    }
    projectDropdownOpen = true;
  }

  async function handleCreate(
    email: string,
    attributes: Record<string, string> | undefined,
  ) {
    // Loop selected projects and add as selectedRole
    await Promise.all(
      selectedProjects.map((projectName) =>
        $addProjectMemberUser.mutateAsync({
          org: organization,
          project: projectName,
          data: { email, role: selectedRole, attributes },
        }),
      ),
    );
  }

  const formId = "create-guests-form";
  const initialValues: {
    emails: string[];
    attributes: AttributeRow[];
  } = { emails: [""], attributes: [] };
  const schema = yup(
    object({
      emails: array(
        string().matches(RFC5322EmailRegex, {
          excludeEmptyString: true,
          message: m.users_invalid_email(),
        }),
      ),
      // Modelled as mixed() rather than array(object(...)):
      // yup infers object fields as optional, which does not match KeyValueInput's row type.
      attributes: mixed<AttributeRow[]>().defined(),
    }),
  );

  const { form, errors, enhance, submit, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        failedInvites = [];
        if (!form.valid) return;

        const emails = form.data.emails.map((e) => e.trim()).filter(Boolean);
        if (emails.length === 0) return;
        if (selectedProjects.length === 0) return;

        // The same attributes apply to every invited email
        const attributes = buildInviteAttributes(form.data.attributes);

        const results = await Promise.all(
          emails.map(async (email, index) => {
            try {
              await handleCreate(email, attributes);
              return { index, email, success: true };
            } catch {
              return { index, email, success: false };
            }
          }),
        );

        const succeeded: string[] = [];
        const failed: string[] = [];
        results
          .sort((a, b) => a.index - b.index)
          .forEach(({ email, success }) =>
            success ? succeeded.push(email) : failed.push(email),
          );

        if (succeeded.length > 0) {
          eventBus.emit("notification", {
            type: "success",
            message: m.users_guests_invited_success({
              count: succeeded.length,
              role: selectedRole,
            }),
          });
        }

        if (failed.length > 0) failedInvites = failed;

        // Invalidate lists
        await invalidateOrgMemberUsers(queryClient, organization);
        await invalidateOrgInvites(queryClient, organization);

        if (failedInvites.length === 0) {
          open = false;
          resetDialogState();
        }
      },
      validationMethod: "oninput",
    },
  );

  $: hasInvalidEmails = $form.emails.some(
    (e, i) => e.length > 0 && $errors.emails?.[i] !== undefined,
  );

  // role label is derived from orgRolesOptions
</script>

<Dialog
  bind:open
  onOpenChange={(dialogOpen) => {
    if (!dialogOpen) {
      resetDialogState();
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
      resetDialogState();
    }}
  >
    <DialogHeader>
      <DialogTitle>{m.users_add_guests_title()}</DialogTitle>
    </DialogHeader>
    <DialogDescription>
      {m.users_add_guests_description()}
    </DialogDescription>
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
        contentClassName="relative"
        bind:values={$form.emails}
        errors={$errors.emails}
        singular="email"
        plural="emails"
      />

      <!-- Project multi-select -->
      <div class="mt-3">
        <div class="text-xs font-medium mb-1">{m.users_project_access()}</div>
        {#if $projectsQuery?.isLoading}
          <DelayedSpinner isLoading={$projectsQuery?.isLoading} size="1rem" />
        {:else if $projectsQuery?.error}
          <div class="flex items-center gap-2">
            <div class="text-xs text-red-500">
              {m.users_failed_load_projects()}{projectsErrorMessage
                ? `: ${projectsErrorMessage}`
                : ""}
            </div>
            <Button type="tertiary" onClick={() => $projectsQuery?.refetch()}
              >{m.users_retry()}</Button
            >
          </div>
        {:else if projects.length === 0}
          <div class="text-xs text-fg-secondary">{m.users_no_projects()}</div>
        {:else}
          <Dropdown.Root bind:open={projectDropdownOpen}>
            <Dropdown.Trigger
              class="min-w-[260px] min-h-[32px] flex flex-row justify-between gap-1 items-center rounded-sm border border-gray-300 bg-surface-background text-sm px-3 {projectDropdownOpen
                ? 'bg-gray-200'
                : 'hover:bg-surface-hover'}"
            >
              <span>
                {selectedProjectsLabel}
              </span>
              {#if projectDropdownOpen}
                <CaretUpIcon size="12px" />
              {:else}
                <CaretDownIcon size="12px" />
              {/if}
            </Dropdown.Trigger>
            <Dropdown.Content align="start" class="w-[260px]">
              {#each projects as p (p.id)}
                <Dropdown.CheckboxItem
                  closeOnSelect={false}
                  class="font-normal flex items-center overflow-hidden"
                  checked={selectedProjects.includes(p.name)}
                  onCheckedChange={() => toggleProjectSelection(p.name)}
                >
                  <span class="truncate w-full" title={p.name}>{p.name}</span>
                </Dropdown.CheckboxItem>
              {/each}
            </Dropdown.Content>
          </Dropdown.Root>
        {/if}
      </div>

      <!-- Access level selector -->
      <div class="mt-3">
        <div class="text-xs font-medium mb-1">{m.users_access_level()}</div>
        <Dropdown.Root bind:open={roleDropdownOpen}>
          <Dropdown.Trigger
            class="min-w-[180px] min-h-[32px] flex flex-row justify-between gap-1 items-center rounded-sm border border-gray-300 bg-surface-background text-sm px-3 {roleDropdownOpen
              ? 'bg-gray-200'
              : 'hover:bg-surface-hover'}"
          >
            <span>{selectedRoleLabel}</span>
            {#if roleDropdownOpen}
              <CaretUpIcon size="12px" />
            {:else}
              <CaretDownIcon size="12px" />
            {/if}
          </Dropdown.Trigger>
          <Dropdown.Content align="start" class="w-[180px]">
            {#each orgRolesOptions as option}
              <Dropdown.CheckboxItem
                checked={selectedRole === option.value}
                onCheckedChange={(checked) => {
                  if (checked) selectedRole = option.value;
                }}
              >
                {option.label}
              </Dropdown.CheckboxItem>
            {/each}
          </Dropdown.Content>
        </Dropdown.Root>
      </div>

      <div class="mt-3 flex flex-col gap-y-1">
        <button
          type="button"
          class="flex items-center gap-x-1 text-xs font-medium text-fg-secondary hover:text-fg-primary w-fit"
          onclick={toggleAttributes}
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
            id="invite-guest-attributes"
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
          {m.users_failed_invite({
            emails: failedInvites.join(", "),
            count: failedInvites.length,
          })}
        </div>
      {/if}
    </form>
    <DialogFooter>
      <Button type="tertiary" onClick={() => (open = false)}
        >{m.users_cancel()}</Button
      >
      <Button
        type="primary"
        submitForm
        form={formId}
        loading={$submitting}
        disabled={hasInvalidEmails ||
          $form.emails.every((e) => !e.trim()) ||
          selectedProjects.length === 0}
      >
        {m.users_add_guests_button()}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
