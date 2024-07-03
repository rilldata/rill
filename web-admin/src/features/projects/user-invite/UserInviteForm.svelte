<script lang="ts">
  import { createAdminServiceAddProjectMember } from "@rilldata/web-admin/client";
  import UserRoleSelect from "@rilldata/web-admin/features/projects/user-invite/UserRoleSelect.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";

  export let organization: string;
  export let project: string;
  export let onInvite: () => void = () => {};

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
      onInvite();
    },
  });

  const { form, handleSubmit } = formState;
</script>

<form
  id="user-invite-form"
  on:submit|preventDefault={handleSubmit}
  class="flex flex-row items-center gap-1.5"
>
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
</form>
