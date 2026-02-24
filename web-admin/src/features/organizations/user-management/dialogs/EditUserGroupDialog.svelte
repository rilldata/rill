<script lang="ts">
  import { page } from "$app/stores";
  import type { V1OrganizationMemberUser } from "@rilldata/web-admin/client";
  import {
    adminServiceListProjectMemberUsergroups,
    adminServiceListProjectsForOrganization,
    createAdminServiceAddProjectMemberUsergroup,
    createAdminServiceAddUsergroupMemberUser,
    createAdminServiceListOrganizationMemberUsers,
    createAdminServiceListUsergroupMemberUsers,
    createAdminServiceRemoveProjectMemberUsergroup,
    createAdminServiceRemoveUsergroupMemberUser,
    createAdminServiceSetProjectMemberUsergroupRole,
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
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { PROJECT_ROLES_OPTIONS } from "@rilldata/web-admin/features/projects/constants";
  import { ProjectUserRoles } from "@rilldata/web-common/features/users/roles";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import { SLUG_REGEX } from "@rilldata/web-admin/features/organizations/user-management/constants.ts";

  export let open = false;
  export let groupName: string;
  export let currentUserEmail: string = "";

  // ── Members state ──────────────────────────────────────────────────
  let memberSearchInput = "";
  let debouncedMemberSearch = "";
  let memberDebounceTimer: ReturnType<typeof setTimeout> | undefined;
  let selectedUsers: V1OrganizationMemberUser[] = [];
  let pendingMemberAdditions: string[] = [];
  let pendingMemberRemovals: string[] = [];
  let membersInitialized = false;

  $: {
    if (memberDebounceTimer) clearTimeout(memberDebounceTimer);
    memberDebounceTimer = setTimeout(() => {
      debouncedMemberSearch = memberSearchInput;
    }, 300);
  }

  // ── Projects state ─────────────────────────────────────────────────
  type ProjectWithRole = { name: string; role: ProjectUserRoles };

  let allOrgProjectNames: string[] = [];
  let initialProjects: ProjectWithRole[] = [];
  let selectedProjects: ProjectWithRole[] = [];
  let projectsInitialized = false;
  let projectsLoading = false;
  let projectSearchInput = "";

  // ── Org / queries ──────────────────────────────────────────────────
  $: organization = $page.params.organization;

  $: listUsergroupMemberUsers = createAdminServiceListUsergroupMemberUsers(
    organization,
    groupName,
  );
  $: userGroupMembersUsers = $listUsergroupMemberUsers.data?.members ?? [];

  $: organizationUsersQuery = createAdminServiceListOrganizationMemberUsers(
    organization,
    debouncedMemberSearch
      ? { pageSize: 50, searchPattern: `${debouncedMemberSearch}%` }
      : { pageSize: 50 },
    { query: { enabled: debouncedMemberSearch.length > 0 } },
  );

  $: availableMemberOptions =
    ($organizationUsersQuery.data?.members?.filter(
      (u): u is typeof u & { userEmail: string; userName: string } =>
        Boolean(u.userEmail) &&
        Boolean(u.userName) &&
        !selectedUsers.some((s) => s.userEmail === u.userEmail),
    ) ?? []).map((u) => ({ value: u.userEmail, label: u.userName }));

  $: allOrganizationUsers = $organizationUsersQuery.data?.members ?? [];

  $: if (userGroupMembersUsers.length > 0 && !membersInitialized) {
    selectedUsers = [...userGroupMembersUsers];
    membersInitialized = true;
  }

  // Load all org projects + this group's current access when the dialog opens.
  $: if (open && !projectsInitialized) {
    void loadProjectsForGroup();
  }

  async function loadProjectsForGroup() {
    projectsLoading = true;
    try {
      const projectsResponse =
        await adminServiceListProjectsForOrganization(organization);
      const allProjects = projectsResponse.projects ?? [];
      allOrgProjectNames = allProjects.map((p) => p.name ?? "").filter(Boolean);

      const results = await Promise.all(
        allProjects.map(async (project) => {
          try {
            const res = await adminServiceListProjectMemberUsergroups(
              organization,
              project.name ?? "",
            );
            const match = (res.members ?? []).find(
              (m) => m.groupName === groupName,
            );
            return match
              ? {
                  name: project.name ?? "",
                  role: (match.roleName ??
                    ProjectUserRoles.Viewer) as ProjectUserRoles,
                }
              : null;
          } catch {
            return null;
          }
        }),
      );

      const loaded = results.filter((r): r is ProjectWithRole => r !== null);
      initialProjects = loaded;
      selectedProjects = [...loaded];
      projectsInitialized = true;
    } catch {
      // Silently ignore; the section will render empty.
    } finally {
      projectsLoading = false;
    }
  }

  $: availableProjectOptions = allOrgProjectNames
    .filter(
      (name) =>
        !selectedProjects.some((p) => p.name === name) &&
        (!projectSearchInput ||
          name.toLowerCase().includes(projectSearchInput.toLowerCase())),
    )
    .map((name) => ({ value: name, label: name }));

  // ── Mutations ──────────────────────────────────────────────────────
  const queryClient = useQueryClient();
  const addUsergroupMemberUser = createAdminServiceAddUsergroupMemberUser();
  const removeUsergroupMemberUser =
    createAdminServiceRemoveUsergroupMemberUser();
  const updateUserGroup = createAdminServiceUpdateUsergroup();
  const addProjectMemberUsergroup =
    createAdminServiceAddProjectMemberUsergroup();
  const removeProjectMemberUsergroup =
    createAdminServiceRemoveProjectMemberUsergroup();
  const setProjectMemberUsergroupRole =
    createAdminServiceSetProjectMemberUsergroupRole();

  // ── Member handlers ────────────────────────────────────────────────
  function handleMemberRemove(email: string) {
    selectedUsers = selectedUsers.filter((u) => u.userEmail !== email);
    pendingMemberRemovals = [...pendingMemberRemovals, email];
    pendingMemberAdditions = pendingMemberAdditions.filter((e) => e !== email);
  }

  function handleMemberAdd(email: string) {
    const user = allOrganizationUsers.find((u) => u.userEmail === email);
    if (user && !selectedUsers.some((s) => s.userEmail === email)) {
      selectedUsers = [...selectedUsers, user];
      pendingMemberAdditions = [...pendingMemberAdditions, email];
      pendingMemberRemovals = pendingMemberRemovals.filter((e) => e !== email);
      memberSearchInput = "";
    }
  }

  // ── Project handlers ───────────────────────────────────────────────
  function handleProjectAdd(name: string) {
    if (!selectedProjects.some((p) => p.name === name)) {
      selectedProjects = [
        ...selectedProjects,
        { name, role: ProjectUserRoles.Viewer },
      ];
      projectSearchInput = "";
    }
  }

  function handleProjectRemove(name: string) {
    selectedProjects = selectedProjects.filter((p) => p.name !== name);
  }

  function handleProjectRoleChange(name: string, role: string) {
    selectedProjects = selectedProjects.map((p) =>
      p.name === name ? { ...p, role: role as ProjectUserRoles } : p,
    );
  }

  // ── Save ───────────────────────────────────────────────────────────
  async function applyPendingChanges() {
    // Members
    for (const email of pendingMemberAdditions) {
      await $addUsergroupMemberUser.mutateAsync({
        org: organization,
        usergroup: groupName,
        email,
        data: {},
      });
    }
    for (const email of pendingMemberRemovals) {
      await $removeUsergroupMemberUser.mutateAsync({
        org: organization,
        usergroup: groupName,
        email,
      });
    }

    // Projects: additions
    const added = selectedProjects.filter(
      (p) => !initialProjects.some((i) => i.name === p.name),
    );
    for (const { name, role } of added) {
      await $addProjectMemberUsergroup.mutateAsync({
        org: organization,
        project: name,
        usergroup: groupName,
        data: { role },
      });
    }

    // Projects: removals
    const removed = initialProjects.filter(
      (i) => !selectedProjects.some((p) => p.name === i.name),
    );
    for (const { name } of removed) {
      await $removeProjectMemberUsergroup.mutateAsync({
        org: organization,
        project: name,
        usergroup: groupName,
      });
    }

    // Projects: role changes
    const roleChanged = selectedProjects.filter((p) => {
      const initial = initialProjects.find((i) => i.name === p.name);
      return initial && initial.role !== p.role;
    });
    for (const { name, role } of roleChanged) {
      await $setProjectMemberUsergroupRole.mutateAsync({
        org: organization,
        project: name,
        usergroup: groupName,
        data: { role },
      });
    }

    // Invalidate caches
    await queryClient.invalidateQueries({
      queryKey: getAdminServiceListUsergroupMemberUsersQueryKey(
        organization,
        groupName,
      ),
    });
    await queryClient.invalidateQueries({
      queryKey: getAdminServiceListOrganizationMemberUsergroupsQueryKey(
        organization,
        { includeCounts: true },
      ),
    });
    const affectedProjects = [
      ...new Set([
        ...added.map((p) => p.name),
        ...removed.map((p) => p.name),
        ...roleChanged.map((p) => p.name),
      ]),
    ];
    for (const project of affectedProjects) {
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectMemberUsergroupsQueryKey(
          organization,
          project,
        ),
      });
    }

    pendingMemberAdditions = [];
    pendingMemberRemovals = [];
  }

  // ── Form ───────────────────────────────────────────────────────────
  const formId = "edit-user-group-form";
  const initialValues = { newName: groupName };

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
        try {
          if (form.data.newName !== groupName) {
            await $updateUserGroup.mutateAsync({
              org: organization,
              usergroup: groupName,
              data: { newName: form.data.newName },
            });
            await queryClient.invalidateQueries({
              queryKey: getAdminServiceListOrganizationMemberUsergroupsQueryKey(
                organization,
                { includeCounts: true },
              ),
            });
          }
          await applyPendingChanges();
          eventBus.emit("notification", {
            message: "User group changes saved successfully",
          });
          open = false;
        } catch (error) {
          eventBus.emit("notification", {
            message: `Error: ${error.response?.data?.message ?? error.message}`,
            type: "error",
          });
        }
      },
    },
  );

  function getMemberMetadata(email: string) {
    const user =
      selectedUsers.find((u) => u.userEmail === email) ||
      allOrganizationUsers.find((u) => u.userEmail === email);
    return user
      ? { name: user.userName ?? "", photoUrl: user.userPhotoUrl ?? undefined }
      : undefined;
  }

  // ── Close / reset ──────────────────────────────────────────────────
  function handleClose() {
    open = false;
    memberSearchInput = "";
    projectSearchInput = "";
    selectedUsers = [];
    pendingMemberAdditions = [];
    pendingMemberRemovals = [];
    membersInitialized = false;
    allOrgProjectNames = [];
    initialProjects = [];
    selectedProjects = [];
    projectsInitialized = false;
    if ($form.newName !== initialValues.newName) {
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

    <div class="flex flex-col gap-4 w-full">
      <!-- Name -->
      <form
        id={formId}
        class="w-full"
        on:submit|preventDefault={submit}
        use:enhance
      >
        <Input
          bind:value={$form.newName}
          id="edit-user-group-name"
          label="Name"
          placeholder="Untitled"
          errors={$errors.newName}
          alwaysShowError={true}
        />
      </form>

      <!-- Projects -->
      <div class="flex flex-col gap-1">
        <div class="text-sm font-medium text-fg-primary">Projects</div>
        <div class="rounded-md border border-gray-200">
          {#if projectsLoading}
            <div class="px-3 py-2 text-sm text-fg-secondary">
              Loading projects…
            </div>
          {:else}
            {#if selectedProjects.length > 0}
              <div class="max-h-40 overflow-y-auto divide-y divide-gray-100">
                {#each selectedProjects as project (project.name)}
                  <div class="flex items-center gap-2 px-3 py-2">
                    <span class="flex-1 truncate text-sm">{project.name}</span>
                    <select
                      class="h-7 rounded-sm border border-gray-300 bg-white px-2 text-xs text-gray-700 cursor-pointer focus:outline-none focus:border-primary-500"
                      value={project.role}
                      on:change={(e) =>
                        handleProjectRoleChange(
                          project.name,
                          e.currentTarget.value,
                        )}
                    >
                      {#each PROJECT_ROLES_OPTIONS as opt (opt.value)}
                        <option value={opt.value}>{opt.label}</option>
                      {/each}
                    </select>
                    <button
                      type="button"
                      class="text-gray-400 hover:text-red-500 text-xs leading-none p-1 rounded hover:bg-red-50"
                      on:click={() => handleProjectRemove(project.name)}
                      aria-label="Remove {project.name}"
                    >
                      ✕
                    </button>
                  </div>
                {/each}
              </div>
            {/if}
            <div
              class="border-t border-gray-200 bg-gray-50"
              class:border-t-0={selectedProjects.length === 0}
            >
              <Combobox
                bind:searchValue={projectSearchInput}
                options={availableProjectOptions}
                placeholder="Search projects…"
                enableClientFiltering={false}
                selectedValues={[]}
                onSelectedChange={(values) => {
                  if (values?.length) handleProjectAdd(values[0].value);
                }}
              />
            </div>
          {/if}
        </div>
      </div>

      <!-- Members -->
      <div class="flex flex-col gap-1">
        <div class="text-sm font-medium text-fg-primary">Members</div>
        <div class="rounded-md border border-gray-200">
          {#if selectedUsers.length > 0}
            <div class="max-h-40 overflow-y-auto divide-y divide-gray-100">
              {#each selectedUsers as user (user.userEmail)}
                <div class="flex items-center gap-2 px-3 py-2">
                  <div class="flex-1 min-w-0">
                    <AvatarListItem
                      name={user.userName ?? ""}
                      email={user.userEmail ?? ""}
                      photoUrl={user.userPhotoUrl}
                      isCurrentUser={user.userEmail === currentUserEmail}
                      role={user.roleName ?? ""}
                    />
                  </div>
                  <button
                    type="button"
                    class="text-gray-400 hover:text-red-500 text-xs leading-none p-1 rounded hover:bg-red-50 shrink-0"
                    on:click={() => handleMemberRemove(user.userEmail ?? "")}
                    aria-label="Remove {user.userName ?? ''}"
                  >
                    ✕
                  </button>
                </div>
              {/each}
            </div>
          {/if}
          <div
            class="border-t border-gray-200 bg-gray-50"
            class:border-t-0={selectedUsers.length === 0}
          >
            <Combobox
              bind:searchValue={memberSearchInput}
              options={availableMemberOptions}
              placeholder="Search members…"
              getMetadata={getMemberMetadata}
              enableClientFiltering={false}
              selectedValues={[]}
              onSelectedChange={(values) => {
                if (values?.length) handleMemberAdd(values[0].value);
              }}
            />
          </div>
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
