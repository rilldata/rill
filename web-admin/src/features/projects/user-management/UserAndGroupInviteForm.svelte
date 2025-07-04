<script lang="ts">
  import {
    createAdminServiceAddProjectMemberUser,
    createAdminServiceAddProjectMemberUsergroup,
    getAdminServiceListOrganizationMemberUsersQueryKey,
    getAdminServiceListProjectInvitesQueryKey,
    getAdminServiceListProjectMemberUsersQueryKey,
    getAdminServiceListProjectMemberUsergroupsQueryKey,
  } from "@rilldata/web-admin/client";
  import { RFC5322EmailRegex } from "@rilldata/web-common/components/forms/validation";
  import { ProjectUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import SearchAndInviteInput from "@rilldata/web-admin/features/projects/user-management/SearchAndInviteInput.svelte";

  export let organization: string;
  export let project: string;
  export let searchList: any[] = [];
  export let onInvite: () => void = () => {};

  const queryClient = useQueryClient();
  const userInvite = createAdminServiceAddProjectMemberUser();
  const addUsergroup = createAdminServiceAddProjectMemberUsergroup();

  async function processInvitations(emailsAndGroups: string[], role: string) {
    const succeededEmails = [];
    const succeededGroups = [];
    const failedEmails = [];
    const failedGroups = [];

    await Promise.all(
      emailsAndGroups.map(async (input) => {
        // Check if input is an email or a group name
        if (RFC5322EmailRegex.test(input)) {
          // Handle as email
          try {
            await $userInvite.mutateAsync({
              organization,
              project,
              data: {
                email: input,
                role: role,
              },
            });
            succeededEmails.push(input);
          } catch {
            failedEmails.push(input);
          }
        } else {
          // Handle as group name
          try {
            await $addUsergroup.mutateAsync({
              organization,
              project,
              usergroup: input,
              data: {
                role: role,
              },
            });
            succeededGroups.push(input);
          } catch {
            failedGroups.push(input);
          }
        }
      }),
    );

    // Batch invalidate queries in parallel for better performance
    await Promise.all([
      queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectMemberUsersQueryKey(
          organization,
          project,
        ),
      }),
      queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectInvitesQueryKey(
          organization,
          project,
        ),
      }),
      queryClient.invalidateQueries({
        queryKey: getAdminServiceListProjectMemberUsergroupsQueryKey(
          organization,
          project,
        ),
      }),
      queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationMemberUsersQueryKey(organization),
        type: "all", // Clear regular and inactive queries
      }),
    ]);

    // Generate success notification message
    let successMessage = "";
    if (succeededEmails.length > 0) {
      successMessage += `Invited ${succeededEmails.length} ${succeededEmails.length === 1 ? "person" : "people"}`;
    }
    if (succeededGroups.length > 0) {
      if (successMessage) successMessage += " and ";
      successMessage += `${successMessage ? "added" : "Added"} ${succeededGroups.length} ${succeededGroups.length === 1 ? "group" : "groups"}`;
    }
    if (successMessage) {
      successMessage += ` as ${role}`;
      eventBus.emit("notification", {
        type: "success",
        message: successMessage,
      });
    }

    // Handle error notifications
    if (failedGroups.length > 0) {
      const groupsText = failedGroups.join(", ");
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to add group${failedGroups.length > 1 ? "s" : ""}: ${groupsText}`,
      });
    }

    if (failedEmails.length > 0) {
      const emailsText = failedEmails.join(", ");
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to invite user${failedEmails.length > 1 ? "s" : ""}: ${emailsText}`,
      });
    }

    return { succeededEmails, succeededGroups, failedEmails, failedGroups };
  }

  function emailOrGroupValidator(value: string) {
    if (!value) return true;
    return (
      RFC5322EmailRegex.test(value) ||
      /^[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*$/.test(value) ||
      "Must be a valid email or group name"
    );
  }

  async function handleSearch(query: string) {
    if (!query) return [];

    const lower = query.toLowerCase();
    return searchList
      .filter((user) => user.identifier.toLowerCase().includes(lower))
      .slice(0, 5); // Limit to 5 results to match previous behavior
  }

  async function onInviteHandler(emailsAndGroups: string[], role: string) {
    await processInvitations(emailsAndGroups, role);
    onInvite();
  }
</script>

<SearchAndInviteInput
  onSearch={handleSearch}
  onInvite={onInviteHandler}
  placeholder="Email or group, separated by commas"
  validators={[emailOrGroupValidator]}
  roleSelect={true}
  initialRole={ProjectUserRoles.Viewer}
  searchKeys={["identifier"]}
  autoFocusInput={-1}
  multiSelect={true}
  {searchList}
/>
