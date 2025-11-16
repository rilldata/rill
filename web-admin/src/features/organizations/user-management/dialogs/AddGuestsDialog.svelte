<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceAddProjectMemberUser,
    createAdminServiceListProjectsForOrganization,
    getAdminServiceListOrganizationInvitesQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
    type V1Project,
  } from "@rilldata/web-admin/client";
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
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import { RFC5322EmailRegex } from "@rilldata/web-common/components/forms/validation";
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { ORG_ROLES_OPTIONS } from "@rilldata/web-admin/features/organizations/constants";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { array, object, string } from "yup";

  export let open = false;

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const addProjectMemberUser = createAdminServiceAddProjectMemberUser();

  let failedInvites: string[] = [];
  let selectedProjects: string[] = [];
  let projectDropdownOpen = false;
  let selectedRole: "admin" | "editor" | "viewer" = "viewer";
  let roleDropdownOpen = false;

  // Projects list
  $: projectsQuery = createAdminServiceListProjectsForOrganization(
    organization,
    undefined,
    {
      query: {
        enabled: !!organization,
        refetchOnMount: true,
        refetchOnWindowFocus: true,
      },
    },
  );
  $: projects = $projectsQuery?.data?.projects ?? ([] as V1Project[]);
  $: projectsErrorMessage =
    getRpcErrorMessage($projectsQuery?.error) ?? $projectsQuery?.error?.message;

  $: selectedRoleLabel =
    ORG_ROLES_OPTIONS.find((o) => o.value === selectedRole)?.label ?? "";

  function toggleProjectSelection(projectName: string) {
    const idx = selectedProjects.indexOf(projectName);
    if (idx >= 0) {
      selectedProjects = selectedProjects.filter(
        (name) => name !== projectName,
      );
    } else {
      selectedProjects = [...selectedProjects, projectName];
    }
  }

  async function handleCreate(email: string) {
    // Loop selected projects and add as selectedRole
    await Promise.all(
      selectedProjects.map((projectName) =>
        $addProjectMemberUser.mutateAsync({
          org: organization,
          project: projectName,
          data: { email, role: selectedRole },
        }),
      ),
    );
  }

  const formId = "create-guests-form";
  const initialValues: { emails: string[] } = { emails: [""] };
  const schema = yup(
    object({
      emails: array(
        string().matches(RFC5322EmailRegex, {
          excludeEmptyString: true,
          message: "Invalid email",
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
        if (!form.valid) return;

        const emails = form.data.emails.map((e) => e.trim()).filter(Boolean);
        if (emails.length === 0) return;
        if (selectedProjects.length === 0) return;

        const results = await Promise.all(
          emails.map(async (email, index) => {
            try {
              await handleCreate(email);
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
            message: `Invited ${succeeded.length} guest${succeeded.length > 1 ? "s" : ""} as ${selectedRole}`,
          });
        }

        if (failed.length > 0) failedInvites = failed;

        // Invalidate lists
        await queryClient.invalidateQueries({
          queryKey:
            getAdminServiceListOrganizationMemberUsersQueryKey(organization),
        });
        await queryClient.invalidateQueries({
          queryKey:
            getAdminServiceListOrganizationInvitesQueryKey(organization),
        });

        if (failedInvites.length === 0) {
          open = false;
          selectedProjects = [];
          selectedRole = "viewer";
        }
      },
      validationMethod: "oninput",
    },
  );

  $: hasInvalidEmails = $form.emails.some(
    (e, i) => e.length > 0 && $errors.emails?.[i] !== undefined,
  );

  // role label is derived from ORG_ROLES_OPTIONS
</script>

<Dialog
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    open = false;
    failedInvites = [];
    selectedProjects = [];
    selectedRole = "viewer";
  }}
  onOpenChange={(open) => {
    if (!open) {
      failedInvites = [];
      selectedProjects = [];
      selectedRole = "viewer";
    }
  }}
>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Add guest users</DialogTitle>
    </DialogHeader>
    <DialogDescription>
      Guests can only access provisioned projects with assigned roles. They do
      not have organization-wide access.
    </DialogDescription>
    <form
      id={formId}
      on:submit|preventDefault={submit}
      class="w-full"
      use:enhance
    >
      <MultiInput
        id="emails"
        placeholder="Add emails, separated by commas"
        contentClassName="relative"
        bind:values={$form.emails}
        errors={$errors.emails}
        singular="email"
        plural="emails"
      />

      <!-- Project multi-select -->
      <div class="mt-3">
        <div class="text-xs font-medium mb-1">Project access</div>
        {#if $projectsQuery?.isLoading}
          <DelayedSpinner isLoading={$projectsQuery?.isLoading} size="1rem" />
        {:else if $projectsQuery?.error}
          <div class="flex items-center gap-2">
            <div class="text-xs text-red-500">
              Failed to load projects{projectsErrorMessage
                ? `: ${projectsErrorMessage}`
                : ""}
            </div>
            <Button type="plain" onClick={() => $projectsQuery?.refetch()}
              >Retry</Button
            >
          </div>
        {:else if projects.length === 0}
          <div class="text-xs text-slate-500">No projects</div>
        {:else}
          <Dropdown.Root bind:open={projectDropdownOpen}>
            <Dropdown.Trigger
              class="min-w-[260px] flex flex-row justify-between gap-1 items-center rounded-sm border border-gray-300 {projectDropdownOpen
                ? 'bg-slate-200'
                : 'hover:bg-slate-100'} px-2 py-1"
            >
              <span>
                {selectedProjects.length > 0
                  ? `${selectedProjects.length} Project${selectedProjects.length > 1 ? "s" : ""}`
                  : "Select projects"}
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
                  class="font-normal flex items-center overflow-hidden"
                  checked={selectedProjects.includes(p.name)}
                  on:click={() => toggleProjectSelection(p.name)}
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
        <div class="text-xs font-medium mb-1">Access level</div>
        <Dropdown.Root bind:open={roleDropdownOpen}>
          <Dropdown.Trigger
            class="min-w-[180px] flex flex-row justify-between gap-1 items-center rounded-sm border border-gray-300 {roleDropdownOpen
              ? 'bg-slate-200'
              : 'hover:bg-slate-100'} px-2 py-1"
          >
            <span>{selectedRoleLabel}</span>
            {#if roleDropdownOpen}
              <CaretUpIcon size="12px" />
            {:else}
              <CaretDownIcon size="12px" />
            {/if}
          </Dropdown.Trigger>
          <Dropdown.Content align="start" class="w-[180px]">
            <Dropdown.Item on:click={() => (selectedRole = "admin")}
              >Admin</Dropdown.Item
            >
            <Dropdown.Item on:click={() => (selectedRole = "editor")}
              >Editor</Dropdown.Item
            >
            <Dropdown.Item on:click={() => (selectedRole = "viewer")}
              >Viewer</Dropdown.Item
            >
          </Dropdown.Content>
        </Dropdown.Root>
      </div>

      {#if failedInvites.length > 0}
        <div class="text-sm text-red-500 py-2">
          {failedInvites.length === 1
            ? `Failed to invite ${failedInvites[0]}`
            : `Failed to invite: ${failedInvites.join(", ")}`}
        </div>
      {/if}
    </form>
    <DialogFooter>
      <Button type="plain" onClick={() => (open = false)}>Cancel</Button>
      <Button
        type="primary"
        submitForm
        form={formId}
        loading={$submitting}
        disabled={hasInvalidEmails ||
          $form.emails.every((e) => !e.trim()) ||
          selectedProjects.length === 0}
      >
        Add guests
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
