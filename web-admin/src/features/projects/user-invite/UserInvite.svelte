<script lang="ts">
  import { createAdminServiceAddProjectMember } from "@rilldata/web-admin/client";
  import UserInviteAllowlist from "@rilldata/web-admin/features/projects/user-invite/UserInviteAllowlist.svelte";
  import UserRoleSelect from "@rilldata/web-admin/features/projects/user-invite/UserRoleSelect.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";

  export let organization: string;
  export let project: string;
  export let onInvited: (users: string[]) => void;

  const userInvite = createAdminServiceAddProjectMember();

  const formState = createForm<{
    emails: Array<{ email: "" }>;
    role: string;
  }>({
    initialValues: {
      emails: [],
      role: "viewer",
    },
    validationSchema: yup.object({
      emails: yup.array().of(
        yup.object().shape({
          email: yup.string().email("Invalid email"),
        }),
      ),
    }),
    onSubmit: async (values) => {
      await Promise.all(
        values.emails.map(({ email }) => {
          return $userInvite.mutateAsync({
            organization,
            project,
            data: {
              email,
              role: values.role,
            },
          });
        }),
      );
      onInvited(values.emails.map(({ email }) => email));
    },
  });

  const { form, handleSubmit } = formState;
</script>

<form id="user-invite-form" on:submit|preventDefault={handleSubmit}>
  <div class="flex flex-col gap-1.5">
    <div class="text-base font-medium">Share this project</div>
    <div class="flex flex-row items-center gap-1.5">
      <div class="w-full">
        <MultiInput
          id="emails"
          label=""
          description=""
          accessorKey="email"
          {formState}
          contentClassName="relative"
        >
          <div
            slot="adjacent-content"
            class="absolute right-0 top-0 h-full items-center flex"
          >
            <UserRoleSelect bind:value={$form.role} />
          </div>
        </MultiInput>
      </div>
      <Button submitForm type="primary" form="user-invite-form">Invite</Button>
    </div>
    <UserInviteAllowlist {organization} {project} />
  </div>
</form>
