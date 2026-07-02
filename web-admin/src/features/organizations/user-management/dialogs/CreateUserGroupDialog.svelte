<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { page } from "$app/stores";
  import type { V1OrganizationMemberUser } from "@rilldata/web-admin/client";
  import {
    createAdminServiceAddUsergroupMemberUser,
    createAdminServiceCreateUsergroup,
    createAdminServiceListOrganizationMemberUsersInfinite,
    getAdminServiceListOrganizationMemberUsergroupsQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
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
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
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
      open = false;

      eventBus.emit("notification", { message: m.groups_created() });
    } catch (error) {
      eventBus.emit("notification", {
        message: `Error: ${error.response.data.message}`,
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
        message: m.groups_changes_saved(),
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

  const { form, enhance, submit, errors, submitting } = superForm(
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

  // Check if form has been modified
  $: hasFormChanges = $form.name !== initialValues.name;

  function handleClose() {
    open = false;
    searchInput = "";
    selectedUsers = [];
    pendingAdditions = [];
    pendingRemovals = [];
    $errors = {};
    // Only reset the form if it has been modified
    if (hasFormChanges) {
      $form.name = initialValues.name;
    }
  }
</script>

<Dialog
  bind:open
  onOpenChange={(open) => {
    if (!open) {
      handleClose();
    }
  }}
>
  <DialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]" interactOutsideBehavior="ignore">
    <DialogHeader>
      <DialogTitle>{m.groups_create_a_group()}</DialogTitle>
    </DialogHeader>
    <form
      id={formId}
      class="w-full"
      onsubmit={(e) => {
        e.preventDefault();
        submit(e);
      }}
      use:enhance
    >
      <div class="flex flex-col gap-4 w-full">
        <Input
          bind:value={$form.name}
          id="create-user-group-name"
          label={m.users_form_name()}
          placeholder={m.users_form_untitled()}
          errors={$errors.name}
          alwaysShowError={true}
        />

        <div class="flex flex-col gap-y-1">
          <label
            for="user-group-users"
            class="line-clamp-1 text-sm font-medium text-fg-primary"
          >
            {m.users_form_users()}
          </label>
          <Combobox
            bind:searchValue={searchInput}
            options={coercedUsersToOptions}
            placeholder={m.org_search_add_remove_users()}
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

              const newEmails = values;
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
            {m.users_user_count({ count: selectedUsers.length })}
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
                {m.users_remove()}
              </Button>
            </div>
          {/each}
        </div>
      </div>
    </div>

    <DialogFooter>
      <Button type="tertiary" onClick={handleClose}>{m.users_cancel()}</Button>
      <Button
        type="primary"
        disabled={$submitting || $form.name.trim() === "" || !!$errors.name}
        form={formId}
        submitForm
      >
        {m.users_create()}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
