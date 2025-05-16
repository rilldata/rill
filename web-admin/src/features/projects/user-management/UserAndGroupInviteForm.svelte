<script lang="ts">
  import {
    createAdminServiceAddProjectMemberUser,
    createAdminServiceAddProjectMemberUsergroup,
    getAdminServiceListOrganizationMemberUsersQueryKey,
    getAdminServiceListProjectInvitesQueryKey,
    getAdminServiceListProjectMemberUsersQueryKey,
    getAdminServiceListProjectMemberUsergroupsQueryKey,
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

  export let organization: string;
  export let project: string;
  export let onInvite: () => void = () => {};

  const queryClient = useQueryClient();
  const userInvite = createAdminServiceAddProjectMemberUser();
  const addUsergroup = createAdminServiceAddProjectMemberUsergroup();

  const initialValues: {
    inputs: string[];
    role: string;
  } = {
    inputs: [""],
    role: "viewer",
  };
  const schema = yup(
    object({
      inputs: array(
        string().test({
          name: "emailOrGroupname",
          message: "Must be a valid email or group name",
          test: (value) => {
            if (!value) return true;
            // Either a valid email or a valid group name (must be at least 3 chars and alphanumeric with hyphens)
            return (
              RFC5322EmailRegex.test(value) ||
              /^[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*$/.test(value)
            );
          },
        }),
      ),
      role: string().required(),
    }),
  );

  const { form, errors, enhance, submit, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;
        const inputs = values.inputs.map((e) => e.trim()).filter(Boolean);
        if (inputs.length === 0) return;

        const succeededEmails = [];
        const succeededGroups = [];
        const failedEmails = [];
        const failedGroups = [];

        await Promise.all(
          inputs.map(async (input) => {
            // Check if input is an email or a group name
            if (RFC5322EmailRegex.test(input)) {
              // Handle as email
              try {
                await $userInvite.mutateAsync({
                  organization,
                  project,
                  data: {
                    email: input,
                    role: values.role,
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
                    role: values.role,
                  },
                });
                succeededGroups.push(input);
              } catch {
                failedGroups.push(input);
              }
            }
          }),
        );

        // Invalidate queries to refresh data
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
          queryKey: getAdminServiceListProjectMemberUsergroupsQueryKey(
            organization,
            project,
          ),
        });

        await queryClient.invalidateQueries({
          queryKey:
            getAdminServiceListOrganizationMemberUsersQueryKey(organization),
          type: "all", // Clear regular and inactive queries
        });

        // Generate success notification message
        let successMessage = "";
        if (succeededEmails.length > 0) {
          successMessage += `Invited ${succeededEmails.length} ${succeededEmails.length === 1 ? "person" : "people"}`;
        }
        if (succeededGroups.length > 0) {
          if (successMessage) successMessage += " and ";
          successMessage += `Added ${succeededGroups.length} ${succeededGroups.length === 1 ? "group" : "groups"}`;
        }
        if (successMessage) {
          successMessage += ` as ${values.role}`;
          eventBus.emit("notification", {
            type: "success",
            message: successMessage,
          });
        }

        // TODO: improve error message
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

        onInvite();
      },
      validationMethod: "oninput",
    },
  );

  $: hasInvalidInputs = $form.inputs.some(
    (e, i) => e.length > 0 && $errors.inputs?.[i] !== undefined,
  );
</script>

<form
  id="user-and-group-invite-form"
  on:submit|preventDefault={submit}
  class="w-full"
  use:enhance
>
  <MultiInput
    id="inputs"
    placeholder="Add emails and groups, separated by commas"
    contentClassName="relative"
    bind:values={$form.inputs}
    errors={$errors.inputs}
    singular="input"
    plural="inputs"
    preventFocus={true}
  >
    <div slot="within-input" class="h-full items-center flex">
      <UserRoleSelect bind:value={$form.role} />
    </div>
    <svelte:fragment slot="beside-input" let:hasSomeValue>
      <Button
        submitForm
        type="primary"
        form="user-invite-form"
        loading={$submitting}
        disabled={hasInvalidInputs || !hasSomeValue}
        forcedStyle="height: 32px !important; padding-left: 20px; padding-right: 20px;"
      >
        Invite
      </Button>
    </svelte:fragment>
  </MultiInput>
</form>
