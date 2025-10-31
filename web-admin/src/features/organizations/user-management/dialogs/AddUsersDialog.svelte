<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceAddOrganizationMemberUser,
    createAdminServiceAddProjectMemberUser,
    createAdminServiceListProjectsForOrganization,
    getAdminServiceListOrganizationInvitesQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
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
  import type { V1Project } from "@rilldata/web-admin/client";
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";

  export let open = false;
  export let email: string;
  export let role: string;
  export let isSuperUser: boolean;

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const addOrganizationMemberUser =
    createAdminServiceAddOrganizationMemberUser();
  const addProjectMemberUser = createAdminServiceAddProjectMemberUser();

  // Org role options (includes Guest)
  const ORG_ROLES_OPTIONS = [
    {
      value: OrgUserRoles.Admin,
      label: "Admin",
      description: "Full control over organization settings and members",
    },
    {
      value: OrgUserRoles.Editor,
      label: "Editor",
      description: "Can manage projects and most org resources",
    },
    {
      value: OrgUserRoles.Viewer,
      label: "Viewer",
      description: "Read-only access to organization and projects",
    },
    {
      value: OrgUserRoles.Guest,
      label: "Guest",
      description: "Access only to selected projects",
    },
  ];

  let failedInvites: string[] = [];
  let selectedProjects: string[] = [];
  let projectDropdownOpen = false;

  // List projects for project access multi-select
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
  $: projectNames = projects.map((p) => p.name);
  $: allProjectNamesSet = new Set(projectNames);
  $: selectedProjectsSet = new Set(selectedProjects);

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

    // If inviting a guest, also add per-project access as Viewer
    if (newRole === OrgUserRoles.Guest && selectedProjects.length > 0) {
      await Promise.all(
        selectedProjects.map((projectName) =>
          $addProjectMemberUser.mutateAsync({
            org: organization,
            project: projectName,
            data: { email: newEmail, role: "viewer" },
          }),
        ),
      );
    }

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
    selectedProjects = [];
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
          message: "Invalid email",
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
        // If inviting a guest without projects selected, do not proceed
        if (values.role === OrgUserRoles.Guest && selectedProjects.length === 0)
          return;

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
            message: `Successfully invited ${succeeded.length} ${
              succeeded.length === 1 ? "person" : "people"
            } as ${values.role}`,
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

  $: hasInvalidEmails = $form.emails.some(
    (e, i) => e.length > 0 && $errors.emails?.[i] !== undefined,
  );
</script>

<Dialog
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    open = false;
    email = "";
    role = "";
    isSuperUser = false;
    failedInvites = [];
  }}
  onOpenChange={(open) => {
    if (!open) {
      email = "";
      role = "";
      isSuperUser = false;
      failedInvites = [];
      selectedProjects = [];
    }
  }}
>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Add users</DialogTitle>
    </DialogHeader>
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
      >
        <div slot="within-input" class="flex items-center h-full">
          <DropdownMenu typeahead={false}>
            <DropdownMenuTrigger
              class="w-18 flex flex-row gap-1 items-center rounded-sm px-2 py-1 hover:bg-slate-100"
            >
              <div class="text-xs">
                {ORG_ROLES_OPTIONS.find((o) => o.value === $form.role)?.label}
              </div>
              <CaretDownIcon size="12px" />
            </DropdownMenuTrigger>
            <DropdownMenuContent
              side="bottom"
              align="end"
              class="w-[260px]"
              strategy="fixed"
            >
              {#each ORG_ROLES_OPTIONS as { value, label, description } (value)}
                <DropdownMenuItem
                  on:click={() => ($form.role = value)}
                  class="text-xs hover:bg-slate-100 {$form.role === value
                    ? 'bg-slate-50'
                    : ''}"
                >
                  <div class="flex flex-col">
                    <div class="text-xs font-medium text-slate-700">
                      {label}
                    </div>
                    <div class="text-slate-500 text-[11px]">{description}</div>
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
            disabled={hasInvalidEmails ||
              !hasSomeValue ||
              ($form.role === OrgUserRoles.Guest &&
                selectedProjects.length === 0)}
            forcedStyle="height: 32px !important;"
          >
            Invite
          </Button>
        </svelte:fragment>
      </MultiInput>
      {#if $form.role === OrgUserRoles.Guest}
        <div class="mt-3">
          <div class="text-xs font-medium mb-1">Project access</div>
          {#if $projectsQuery?.isLoading}
            <div class="text-xs text-slate-500">Loading projects…</div>
          {:else if $projectsQuery?.error}
            <div class="text-xs text-red-500">Failed to load projects</div>
          {:else if projects.length === 0}
            <div class="text-xs text-slate-500">No projects</div>
          {:else}
            <Dropdown.Root bind:open={projectDropdownOpen}>
              <Dropdown.Trigger
                class="min-w-[224px] flex flex-row justify-between gap-1 items-center rounded-sm border border-gray-300 {projectDropdownOpen
                  ? 'bg-slate-200'
                  : 'hover:bg-slate-100'} px-2 py-1"
              >
                <span class="capitalize">
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
              <Dropdown.Content align="start" class="w-[224px]">
                {#each projects as p (p.id)}
                  <Dropdown.CheckboxItem
                    class="font-normal flex items-center"
                    checked={selectedProjects.includes(p.name)}
                    on:click={() => {
                      const idx = selectedProjects.indexOf(p.name);
                      if (idx >= 0)
                        selectedProjects = selectedProjects.filter(
                          (n) => n !== p.name,
                        );
                      else selectedProjects = [...selectedProjects, p.name];
                    }}
                  >
                    <span>{p.name}</span>
                  </Dropdown.CheckboxItem>
                {/each}
              </Dropdown.Content>
            </Dropdown.Root>
          {/if}
        </div>
      {/if}
      {#if failedInvites.length > 0}
        <div class="text-sm text-red-500 py-2">
          {failedInvites.length === 1
            ? `${failedInvites[0]} is already a member of this organization`
            : `${failedInvites.join(", ")} are already members of this organization`}
        </div>
      {/if}
    </form>
  </DialogContent>
</Dialog>
