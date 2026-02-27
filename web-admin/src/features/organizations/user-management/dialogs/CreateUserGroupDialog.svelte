<script lang="ts">
  import { page } from "$app/stores";
  import type {
    V1OrganizationMemberUser,
    V1Project,
  } from "@rilldata/web-admin/client";
  import {
    createAdminServiceAddProjectMemberUsergroup,
    createAdminServiceAddUsergroupMemberUser,
    createAdminServiceCreateUsergroup,
    createAdminServiceListOrganizationMemberUsersInfinite,
    createAdminServiceListProjectsForOrganization,
    getAdminServiceListOrganizationMemberUsergroupsQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
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

  let selectedProjects: string[] = [];
  let projectDropdownOpen = false;
  let selectedRole: ProjectUserRoles = ProjectUserRoles.Viewer;
  let roleDropdownOpen = false;

  // Debounce search input to avoid too many API calls.
  // Use a standard Svelte reactive block: it re-runs whenever `searchInput` changes.
  // We capture `searchInput` into a local constant to avoid race conditions in the timeout.
  $: {
    const current = searchInput;
    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      debouncedSearchText = current;
    }, 300);
  }

  $: organization = $page.params.organization;

  // Projects list
  $: projectsQuery = createAdminServiceListProjectsForOrganization(
    organization,
    undefined,
    {
      query: {
        enabled: open && !!organization,
      },
    },
  );
  $: projects = $projectsQuery?.data?.projects ?? ([] as V1Project[]);

  $: selectedRoleLabel =
    PROJECT_ROLES_OPTIONS.find((o) => o.value === selectedRole)?.label ??
    "Viewer";

  $: selectedProjectsLabel = (() => {
    if (selectedProjects.length === 0) return "Select projects";
    if (selectedProjects.length === 1) return selectedProjects[0];
    return `${selectedProjects.length} Projects`;
  })();

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

  // Infinite query for organization users (debounced by search)
  $: organizationUsersInfiniteQuery =
    createAdminServiceListOrganizationMemberUsersInfinite(
      organization,
      debouncedSearchText
        ? {
            pageSize: 50,
            searchPattern: `${debouncedSearchText}%`,
          }
        : { pageSize: 50 },
      {
        query: {
          enabled: open,
          getNextPageParam: (lastPage) => {
            return lastPage.nextPageToken !== ""
              ? lastPage.nextPageToken
              : undefined;
          },
        },
      },
    );

  $: organizationUsers = $organizationUsersInfiniteQuery.data?.pages
    ? $organizationUsersInfiniteQuery.data.pages.flatMap((p) => p.members ?? [])
    : [];

  $: hasMoreUsers =
    // Prefer built-in flag if available
    ($organizationUsersInfiniteQuery?.hasNextPage ??
      (($organizationUsersInfiniteQuery?.data?.pages?.length ?? 0) > 0 &&
        ($organizationUsersInfiniteQuery?.data?.pages?.[
          ($organizationUsersInfiniteQuery?.data?.pages?.length ?? 1) - 1
        ]?.nextPageToken ?? "") !== "")) ||
    false;

  $: isLoadingMoreUsers =
    $organizationUsersInfiniteQuery?.isFetchingNextPage ?? false;

  function loadMoreUsers() {
    const fetchNext = $organizationUsersInfiniteQuery?.fetchNextPage;
    if (typeof fetchNext === "function") {
      fetchNext();
    }
  }

  const queryClient = useQueryClient();
  const createUserGroup = createAdminServiceCreateUsergroup();
  const addUsergroupMemberUser = createAdminServiceAddUsergroupMemberUser();
  const addProjectMemberUsergroup =
    createAdminServiceAddProjectMemberUsergroup();

  async function handleCreate(newName: string) {
    try {
      await $createUserGroup.mutateAsync({
        org: organization,
        data: {
          name: newName,
        },
      });

      // Apply pending user changes after group creation
      await applyPendingChanges(newName);

      // Add group to selected projects
      await applyProjectAccess(newName);

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListOrganizationMemberUsergroupsQueryKey(
          organization,
          {
            includeCounts: true,
          },
        ),
      });

      groupName = "";
      selectedUsers = [];
      pendingAdditions = [];
      pendingRemovals = [];
      selectedProjects = [];
      selectedRole = ProjectUserRoles.Viewer;
      open = false;

      eventBus.emit("notification", { message: "User group created" });
    } catch (error) {
      eventBus.emit("notification", {
        message: `Error: ${error.response.data.message}`,
        type: "error",
      });
    }
  }

  async function applyProjectAccess(usergroup: string) {
    if (selectedProjects.length === 0) return;

    try {
      await Promise.all(
        selectedProjects.map(async (projectName) => {
          await $addProjectMemberUsergroup.mutateAsync({
            org: organization,
            project: projectName,
            usergroup: usergroup,
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
        }),
      );
    } catch (error) {
      eventBus.emit("notification", {
        message: `Error adding group to projects: ${error.response?.data?.message || error.message}`,
        type: "error",
      });
    }
  }

  async function applyPendingChanges(usergroup: string) {
    try {
      // Add pending users to the group
      for (const email of pendingAdditions) {
        await $addUsergroupMemberUser.mutateAsync({
          org: organization,
          usergroup: usergroup,
          email: email,
          data: {},
        });
      }

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
        message: "User group changes saved successfully",
      });
    } catch (error) {
      eventBus.emit("notification", {
        message: `Error: ${error.response.data.message}`,
        type: "error",
      });
    }
  }

  async function handleRemove(email: string) {
    selectedUsers = selectedUsers.filter((user) => user.userEmail !== email);
    pendingRemovals = [...pendingRemovals, email];
    pendingAdditions = pendingAdditions.filter((e) => e !== email);
  }

  async function handleAdd(email: string) {
    const user = organizationUsers.find((u) => u.userEmail === email);

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

  const formId = "create-user-group-form";

  const initialValues = {
    name: "",
  };

  const schema = yup(
    object({
      name: string()
        .required("Name is required")
        .min(1, "Name must be at least 1 character")
        .max(40, "Name must be at most 40 characters")
        .matches(
          SLUG_REGEX,
          "Name can only include letters, numbers, underscores, and hyphens — no spaces or special characters",
        ),
    }),
  );

  const { form, enhance, submit, errors, submitting, reset } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      validationMethod: "oninput",
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        await handleCreate(values.name);
        open = false;
      },
    },
  );

  $: coercedUsersToOptions = organizationUsers.map((user) => ({
    value: user.userEmail,
    label: user.userName,
  }));

  function getMetadata(email: string) {
    const user = organizationUsers.find((user) => user.userEmail === email);
    return user
      ? { name: user.userName, photoUrl: user.userPhotoUrl }
      : undefined;
  }

  function handleClose() {
    open = false;
    searchInput = "";
    selectedUsers = [];
    pendingAdditions = [];
    pendingRemovals = [];
    selectedProjects = [];
    selectedRole = ProjectUserRoles.Viewer;
    projectDropdownOpen = false;
    roleDropdownOpen = false;
    reset();
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
      <DialogTitle>Create a group</DialogTitle>
    </DialogHeader>
    <form
      id={formId}
      class="w-full"
      on:submit|preventDefault={submit}
      use:enhance
    >
      <div class="flex flex-col gap-4 w-full">
        <!-- Name -->
        <Input
          bind:value={$form.name}
          id="create-user-group-name"
          label="Name"
          placeholder="Untitled"
          errors={$errors.name}
          alwaysShowError={true}
        />

        <!-- Project access multi-select -->
        <div class="flex flex-col gap-y-1">
          <label
            for="project-access"
            class="line-clamp-1 text-sm font-medium text-fg-primary"
          >
            Project access
          </label>
          {#if $projectsQuery?.isLoading}
            <div class="text-sm text-fg-secondary">Loading projects...</div>
          {:else if projects.length === 0}
            <div class="text-sm text-fg-secondary">No projects available</div>
          {:else}
            <Dropdown.Root
              bind:open={projectDropdownOpen}
              closeOnItemClick={false}
            >
              <Dropdown.Trigger
                class="min-h-[36px] flex flex-row justify-between gap-1 items-center rounded-sm border border-gray-300 bg-surface-background text-sm px-3 {projectDropdownOpen
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
              <Dropdown.Content
                align="start"
                sameWidth
                class="max-h-60 overflow-y-auto"
              >
                {#each projects as p (p.id)}
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

        <!-- Access level selector -->
        <div class="flex flex-col gap-y-1">
          <label
            for="access-level"
            class="line-clamp-1 text-sm font-medium text-fg-primary"
          >
            Access level
          </label>
          <Dropdown.Root bind:open={roleDropdownOpen}>
            <Dropdown.Trigger
              class="min-h-[36px] flex flex-row justify-between gap-1 items-center rounded-sm border border-gray-300 bg-surface-background text-sm px-3 {roleDropdownOpen
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
            <Dropdown.Content align="start" sameWidth>
              {#each PROJECT_ROLES_OPTIONS as option}
                <Dropdown.Item
                  class="font-normal flex flex-col items-start py-2 {selectedRole ===
                  option.value
                    ? 'bg-surface-active'
                    : ''}"
                  on:click={() => (selectedRole = option.value)}
                >
                  <span class="font-medium">{option.label}</span>
                  <span class="text-xs text-fg-secondary"
                    >{option.description}</span
                  >
                </Dropdown.Item>
              {/each}
            </Dropdown.Content>
          </Dropdown.Root>
        </div>

        <!-- Users -->
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
            loadMore={loadMoreUsers}
            hasMore={hasMoreUsers}
            isLoadingMore={isLoadingMoreUsers}
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
        disabled={$submitting || $form.name.trim() === "" || !!$errors.name}
        form={formId}
        submitForm
      >
        Create
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
