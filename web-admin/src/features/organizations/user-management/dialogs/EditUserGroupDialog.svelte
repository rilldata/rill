<script lang="ts">
  import { page } from "$app/stores";
  import type {
    V1OrganizationMemberUser,
    V1Project,
  } from "@rilldata/web-admin/client";
  import {
    adminServiceListProjectMemberUsergroups,
    adminServiceListProjectsForOrganization,
    createAdminServiceAddProjectMemberUsergroup,
    createAdminServiceAddUsergroupMemberUser,
    createAdminServiceListOrganizationMemberUsers,
    createAdminServiceListUsergroupMemberUsers,
    createAdminServiceRemoveProjectMemberUsergroup,
    createAdminServiceRemoveUsergroupMemberUser,
    createAdminServiceUpdateUsergroup,
    getAdminServiceListOrganizationMemberUsergroupsQueryKey,
    getAdminServiceListProjectMemberUsergroupsQueryKey,
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
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { PROJECT_ROLES_OPTIONS } from "@rilldata/web-admin/features/projects/constants";
  import { ProjectUserRoles } from "@rilldata/web-common/features/users/roles";
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

  let allProjects: V1Project[] = [];
  let selectedProjects: string[] = [];
  let originalProjects: string[] = [];
  let projectDropdownOpen = false;
  let selectedRole: ProjectUserRoles = ProjectUserRoles.Viewer;
  let roleDropdownOpen = false;
  let projectsLoaded = false;
  let projectsLoading = false;

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

  // Load projects when dialog opens
  async function loadProjects() {
    if (projectsLoaded || projectsLoading) return;
    projectsLoading = true;

    try {
      const projectsResponse =
        await adminServiceListProjectsForOrganization(organization);
      allProjects = projectsResponse.projects ?? [];

      // Check which projects this group has access to
      const projectAccessResults = await Promise.all(
        allProjects.map(async (project) => {
          try {
            const usergroupsResponse =
              await adminServiceListProjectMemberUsergroups(
                organization,
                project.name ?? "",
              );
            const members = usergroupsResponse.members ?? [];
            const hasAccess = members.some((m) => m.groupName === groupName);
            return { projectName: project.name ?? "", hasAccess };
          } catch {
            return { projectName: project.name ?? "", hasAccess: false };
          }
        }),
      );

      const accessibleProjectNames = projectAccessResults
        .filter((r) => r.hasAccess)
        .map((r) => r.projectName);

      selectedProjects = [...accessibleProjectNames];
      originalProjects = [...accessibleProjectNames];
      projectsLoaded = true;
    } catch {
      // Ignore errors
    } finally {
      projectsLoading = false;
    }
  }

  $: if (open && !projectsLoaded) {
    void loadProjects();
  }

  $: selectedProjectsLabel = (() => {
    if (selectedProjects.length === 0) return "Select projects";
    if (selectedProjects.length === 1) return selectedProjects[0];
    return `${selectedProjects.length} Projects`;
  })();

  $: selectedRoleLabel =
    PROJECT_ROLES_OPTIONS.find((o) => o.value === selectedRole)?.label ??
    "Viewer";

  function toggleProjectSelection(projectName: string) {
    const idx = selectedProjects.indexOf(projectName);
    if (idx >= 0) {
      selectedProjects = selectedProjects.filter((name) => name !== projectName);
    } else {
      selectedProjects = [...selectedProjects, projectName];
    }
    projectDropdownOpen = true;
  }

  const queryClient = useQueryClient();
  const addUsergroupMemberUser = createAdminServiceAddUsergroupMemberUser();
  const removeUserGroupMember = createAdminServiceRemoveUsergroupMemberUser();
  const updateUserGroup = createAdminServiceUpdateUsergroup();
  const addProjectMemberUsergroup = createAdminServiceAddProjectMemberUsergroup();
  const removeProjectMemberUsergroup = createAdminServiceRemoveProjectMemberUsergroup();

  function handleRemove(email: string) {
    selectedUsers = selectedUsers.filter((user) => user.userEmail !== email);
    pendingRemovals = [...pendingRemovals, email];
    pendingAdditions = pendingAdditions.filter((e) => e !== email);
  }

  async function handleRename(groupName: string, newName: string) {
    try {
      await $updateUserGroup.mutateAsync({
        org: organization,
        usergroup: groupName,
        data: {
          newName: newName,
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

  async function applyProjectChanges() {
    const projectsToAdd = selectedProjects.filter(
      (p) => !originalProjects.includes(p),
    );
    const projectsToRemove = originalProjects.filter(
      (p) => !selectedProjects.includes(p),
    );

    try {
      // Add group to new projects
      for (const projectName of projectsToAdd) {
        await $addProjectMemberUsergroup.mutateAsync({
          org: organization,
          project: projectName,
          usergroup: groupName,
          data: {
            role: selectedRole,
          },
        });

        await queryClient.invalidateQueries({
          queryKey: getAdminServiceListProjectMemberUsergroupsQueryKey(
            organization,
            projectName,
          ),
        });
      }

      // Remove group from projects
      for (const projectName of projectsToRemove) {
        await $removeProjectMemberUsergroup.mutateAsync({
          org: organization,
          project: projectName,
          usergroup: groupName,
        });

        await queryClient.invalidateQueries({
          queryKey: getAdminServiceListProjectMemberUsergroupsQueryKey(
            organization,
            projectName,
          ),
        });
      }

      originalProjects = [...selectedProjects];
    } catch (error) {
      eventBus.emit("notification", {
        message: `Error updating project access: ${error.response?.data?.message || error.message}`,
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
          await applyProjectChanges();
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
    allProjects = [];
    selectedProjects = [];
    originalProjects = [];
    projectsLoaded = false;
    projectsLoading = false;
    selectedRole = ProjectUserRoles.Viewer;
    projectDropdownOpen = false;
    roleDropdownOpen = false;
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
            class="line-clamp-1 text-sm font-medium text-fg-primary"
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

        <!-- Project access multi-select -->
        <div class="flex flex-col gap-y-1">
          <label
            for="project-access"
            class="line-clamp-1 text-sm font-medium text-fg-primary"
          >
            Project access
          </label>
          {#if projectsLoading}
            <div class="text-sm text-fg-secondary">Loading projects...</div>
          {:else if allProjects.length === 0}
            <div class="text-sm text-fg-secondary">No projects available</div>
          {:else}
            <Dropdown.Root
              bind:open={projectDropdownOpen}
              closeOnItemClick={false}
            >
              <Dropdown.Trigger
                class="w-full min-h-[36px] flex flex-row justify-between gap-1 items-center rounded-sm border border-gray-300 bg-surface-background text-sm px-3 {projectDropdownOpen
                  ? 'bg-gray-200'
                  : 'hover:bg-surface-hover'}"
              >
                <span class="truncate">
                  {selectedProjectsLabel}
                </span>
                {#if projectDropdownOpen}
                  <CaretUpIcon size="12px" />
                {:else}
                  <CaretDownIcon size="12px" />
                {/if}
              </Dropdown.Trigger>
              <Dropdown.Content align="start" class="w-full max-h-60 overflow-y-auto">
                {#each allProjects as p (p.id)}
                  <Dropdown.CheckboxItem
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

        <!-- Access level selector for new projects -->
        {#if selectedProjects.some((p) => !originalProjects.includes(p))}
          <div class="flex flex-col gap-y-1">
            <label
              for="access-level"
              class="line-clamp-1 text-sm font-medium text-fg-primary"
            >
              Access level (for new projects)
            </label>
            <Dropdown.Root bind:open={roleDropdownOpen}>
              <Dropdown.Trigger
                class="w-full min-h-[36px] flex flex-row justify-between gap-1 items-center rounded-sm border border-gray-300 bg-surface-background text-sm px-3 {roleDropdownOpen
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
              <Dropdown.Content align="start" class="w-full">
                {#each PROJECT_ROLES_OPTIONS as option}
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
        {/if}
      </div>
    </form>

    <div class="flex flex-col gap-2 w-full">
      {#if selectedUsers.length > 0}
        <div class="flex flex-row items-center gap-x-1">
          <div class="text-xs font-semibold uppercase text-fg-secondary">
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
                type="destructive"
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
      <Button type="tertiary" onClick={handleClose}>Cancel</Button>
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
