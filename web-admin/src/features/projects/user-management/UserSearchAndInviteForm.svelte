<script lang="ts">
  import {
    createAdminServiceAddProjectMemberUser,
    getAdminServiceListOrganizationMemberUsersQueryKey,
    getAdminServiceListProjectInvitesQueryKey,
    getAdminServiceListProjectMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import UserRoleSelect from "@rilldata/web-admin/features/projects/user-management/UserRoleSelect.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import { RFC5322EmailRegex } from "@rilldata/web-common/components/forms/validation";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { array, object, string } from "yup";
  import type {
    V1ProjectMemberUser,
    V1ProjectInvite,
    V1MemberUsergroup,
  } from "@rilldata/web-admin/client";
  import AvatarListItem from "@rilldata/web-admin/features/organizations/users/AvatarListItem.svelte";
  import { Combobox } from "bits-ui";
  import type { Selected } from "bits-ui";

  export let organization: string;
  export let project: string;
  export let onInvite: () => void = () => {};
  export let searchUsersList: {
    value: string;
    label: string;
    name: string;
    type: "member" | "invite" | "group";
    user?: V1ProjectMemberUser | V1ProjectInvite;
    group?: V1MemberUsergroup;
  }[] = [];

  $: console.log(searchUsersList);

  type PendingUser = {
    value: string;
    name: string;
    label: string;
  };

  let searchText = "";
  let comboboxOptions: Array<{
    value: string;
    label: string;
    type: string;
    name?: string;
    group?: V1MemberUsergroup;
    disabled?: boolean;
    isEmail?: boolean;
  }> = searchUsersList.map((user) => ({
    value: user.value,
    label: user.label,
    type: user.type,
    group: user.group,
    name: user.name,
  }));

  // Array to store pending selections
  let pendingSelections: string[] = [];
  let pendingUsers: PendingUser[] = [];

  function updatePendingUsers() {
    pendingUsers = pendingSelections.map((value) => {
      const user = searchUsersList.find((u) => u.value === value);
      if (user) {
        return {
          value: user.value,
          name: user.name,
          label: user.label,
        };
      } else {
        return {
          value,
          name: value,
          label: value,
        };
      }
    });
  }

  const queryClient = useQueryClient();
  const userInvite = createAdminServiceAddProjectMemberUser();

  const initialValues: {
    emails: string[];
    role: string;
  } = {
    emails: [""],
    role: "viewer",
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

  let needsReset = false;

  const { form, errors, enhance, submit, submitting, reset } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        // Combine text input emails with selected users/groups
        const emailsToInvite = [
          ...values.emails.map((e) => e.trim()).filter(Boolean),
          ...pendingSelections,
        ];

        if (emailsToInvite.length === 0) return;

        const succeeded = [];
        let errored = false;
        await Promise.all(
          emailsToInvite.map(async (email) => {
            try {
              await $userInvite.mutateAsync({
                organization,
                project,
                data: {
                  email,
                  role: values.role,
                },
              });
              succeeded.push(email);
            } catch {
              errored = true;
            }
          }),
        );

        await queryClient.invalidateQueries({
          queryKey: getAdminServiceListProjectMemberUsersQueryKey(
            organization,
            project,
          ),
        });

        await queryClient.invalidateQueries({
          queryKey: getAdminServiceListProjectInvitesQueryKey(
            organization,
            project,
          ),
        });

        await queryClient.invalidateQueries({
          queryKey:
            getAdminServiceListOrganizationMemberUsersQueryKey(organization),
          type: "all", // Clear regular and inactive queries
        });

        eventBus.emit("notification", {
          type: "success",
          message: `Invited ${succeeded.length} ${succeeded.length === 1 ? "person" : "people"} as ${values.role}`,
        });

        // Mark form for reset instead of directly changing the store
        needsReset = true;
        pendingSelections = [];
        pendingUsers = [];
        searchText = "";

        onInvite();
        if (errored) {
          // TODO: there no mocks for this yet, but will be added in future.
          //       the challenge here is how to show it for all the emails that fail
        }
      },
      validationMethod: "oninput",
    },
  );

  // Reset form at the top level when needed
  $: if (needsReset) {
    reset();
    needsReset = false;
  }

  $: hasInvalidEmails = $form.emails.some(
    (e, i) => e.length > 0 && $errors.emails?.[i] !== undefined,
  );

  function handleSelectedChange(selected: Selected<string>[] | undefined) {
    if (!selected || selected.length === 0) return;

    const lastSelected = selected[selected.length - 1];
    if (!lastSelected) return;

    const user = searchUsersList.find((u) => u.value === lastSelected.value);
    if (user) {
      // Add to pending selections instead of immediately inviting
      if (!pendingSelections.includes(user.value)) {
        pendingSelections = [...pendingSelections, user.value];
        updatePendingUsers();
      }
    } else if (RFC5322EmailRegex.test(lastSelected.value)) {
      // Valid email that's not in the search list
      if (!pendingSelections.includes(lastSelected.value)) {
        pendingSelections = [...pendingSelections, lastSelected.value];
        updatePendingUsers();
      }
    } else {
      // Regular text input
      $form.emails = [$form.emails[0] || lastSelected.value];
    }

    // Clear search text after selection
    searchText = "";
  }

  // Group search results by type
  $: groupedResults = {
    groups: searchUsersList.filter((item) => item.type === "group"),
    members: searchUsersList.filter((item) => item.type === "member"),
    guests: searchUsersList.filter((item) => item.type === "invite"),
  };

  // Filter based on search text
  $: filteredResults = searchText
    ? {
        groups: groupedResults.groups.filter(
          (item) =>
            item.name.toLowerCase().includes(searchText.toLowerCase()) ||
            item.value.toLowerCase().includes(searchText.toLowerCase()),
        ),
        members: groupedResults.members.filter(
          (item) =>
            item.name.toLowerCase().includes(searchText.toLowerCase()) ||
            item.value.toLowerCase().includes(searchText.toLowerCase()),
        ),
        guests: groupedResults.guests.filter(
          (item) =>
            item.name.toLowerCase().includes(searchText.toLowerCase()) ||
            item.value.toLowerCase().includes(searchText.toLowerCase()),
        ),
      }
    : groupedResults;

  // Combined options for combobox with section headers
  $: comboboxOptions = searchText
    ? [
        // If empty result, add "Invite email" option for valid emails
        ...(RFC5322EmailRegex.test(searchText) &&
        !searchUsersList.some((u) => u.value === searchText)
          ? [
              {
                value: searchText,
                label: `Invite ${searchText}`,
                type: "email",
                isEmail: true,
              },
            ]
          : []),

        // Groups section
        ...(filteredResults.groups.length > 0
          ? [
              {
                value: "groups-header",
                label: "GROUPS",
                type: "header",
                disabled: true,
              },
            ]
          : []),
        ...filteredResults.groups.map((g) => ({
          value: g.value,
          label: g.label,
          type: g.type,
          name: g.name,
          group: g.group,
        })),

        // Members section
        ...(filteredResults.members.length > 0
          ? [
              {
                value: "members-header",
                label: "MEMBERS",
                type: "header",
                disabled: true,
              },
            ]
          : []),
        ...filteredResults.members.map((m) => ({
          value: m.value,
          label: m.label,
          type: m.type,
          name: m.name,
        })),

        // Guests section
        ...(filteredResults.guests.length > 0
          ? [
              {
                value: "guests-header",
                label: "GUESTS",
                type: "header",
                disabled: true,
              },
            ]
          : []),
        ...filteredResults.guests.map((g) => ({
          value: g.value,
          label: g.label,
          type: g.type,
          name: g.name,
        })),
      ]
    : [];

  function removePendingSelection(value: string) {
    pendingSelections = pendingSelections.filter((v) => v !== value);
    updatePendingUsers();
  }

  function handleInvite() {
    // For direct text input
    if (searchText && RFC5322EmailRegex.test(searchText)) {
      if (!pendingSelections.includes(searchText)) {
        pendingSelections = [...pendingSelections, searchText];
        updatePendingUsers();
      }
      searchText = "";
    }

    submit();
  }

  // Instead of renderItemContent function, let's use AvatarListItem directly
  function getGroupCount(group: any): number {
    return group?.usersCount || 0;
  }

  function getAvatarShape(type: string): "circle" | "square" {
    if (type === "group") return "square";
    return "circle";
  }

  function getAvatarColor(type: string) {
    if (type === "group") return "bg-blue-100 text-blue-600";
    if (type === "member") return "bg-red-100 text-red-500";
    if (type === "invite") return "bg-blue-500 text-white";
    return "";
  }
</script>

<div class="flex flex-col gap-4 w-full">
  <form
    id="user-invite-form"
    on:submit|preventDefault={handleInvite}
    class="w-full"
    use:enhance
  >
    <div class="relative">
      <div class="flex items-center">
        <div class="flex-grow">
          <Combobox.Root
            items={comboboxOptions}
            onSelectedChange={handleSelectedChange}
            multiple={true}
            bind:inputValue={searchText}
          >
            <Combobox.Input
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-100 focus:border-primary-500"
              placeholder="Search users or enter email addresses"
              aria-label="Search users or enter email addresses"
            />

            <Combobox.Content
              class="w-full rounded border border-gray-200 bg-white shadow-md outline-none max-h-[400px] overflow-y-auto p-0"
              sideOffset={8}
            >
              {#if comboboxOptions.length === 0}
                <div class="px-4 py-2 text-xs text-gray-500">
                  No results found
                </div>
              {:else}
                {#each comboboxOptions as item (item.value)}
                  {#if item.type === "header"}
                    <div
                      class="text-sm font-medium text-gray-500 px-3 py-2 border-b border-gray-200 w-full"
                    >
                      {item.label}
                    </div>
                  {:else if item.isEmail}
                    <Combobox.Item
                      value={item.value}
                      label={item.label}
                      class="w-full select-none p-0 outline-none transition-all duration-75 data-[highlighted]:bg-gray-50"
                    >
                      <div
                        class="flex items-center px-3 py-3 border-b border-gray-100"
                      >
                        <div class="text-primary-600">{item.label}</div>
                      </div>
                    </Combobox.Item>
                  {:else}
                    <Combobox.Item
                      value={item.value}
                      label={item.label}
                      class="w-full select-none  outline-none transition-all duration-75 data-[highlighted]:bg-gray-50 p-2"
                    >
                      <AvatarListItem
                        parentDivClass="py-0"
                        name={item.name}
                        email={item.value}
                        photoUrl={null}
                        shape={getAvatarShape(item.type)}
                        count={item.type === "group"
                          ? getGroupCount(item.group)
                          : 0}
                        showGuestChip={item.type === "invite"}
                        leftSpacing={false}
                      />
                    </Combobox.Item>
                  {/if}
                {/each}
              {/if}
            </Combobox.Content>
          </Combobox.Root>
        </div>
        <div class="ml-2 h-full">
          <UserRoleSelect bind:value={$form.role} />
        </div>
        <div class="ml-2">
          <Button
            submitForm
            type="primary"
            form="user-invite-form"
            loading={$submitting}
            disabled={hasInvalidEmails &&
              pendingSelections.length === 0 &&
              searchText.length === 0}
            forcedStyle="height: 32px !important; padding-left: 20px; padding-right: 20px;"
          >
            Invite
          </Button>
        </div>
      </div>

      {#if pendingSelections.length > 0}
        <div class="mt-2 flex flex-wrap gap-2">
          {#each pendingUsers as user}
            <div
              class="flex items-center bg-gray-100 rounded-md px-2 py-1 text-sm"
            >
              <span>{user.name}</span>
              <button
                type="button"
                class="ml-2 text-gray-500 hover:text-gray-700"
                on:click={() => removePendingSelection(user.value)}
              >
                ×
              </button>
            </div>
          {/each}
        </div>
      {/if}

      {#if $errors.emails && Object.values($errors.emails).some(Boolean)}
        <div class="text-red-500 text-sm mt-1">
          {Object.values($errors.emails).filter(Boolean).join(", ")}
        </div>
      {/if}
    </div>
  </form>
</div>

<style>
  :global(.custom-avatar-size :where(div) > :first-child) {
    height: 50px !important;
    width: 50px !important;
    font-size: 1.5rem !important;
  }

  :global(.custom-avatar-size img),
  :global(.custom-avatar-size .rounded-sm) {
    height: 50px !important;
    width: 50px !important;
  }

  :global(.custom-avatar-size .text-sm) {
    font-size: 1rem !important;
  }
</style>
