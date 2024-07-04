<script lang="ts">
  import { createAdminServiceAddProjectMember } from "@rilldata/web-admin/client";
  import UserRoleSelect from "@rilldata/web-admin/features/projects/user-invite/UserRoleSelect.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { array, object, string } from "yup";

  export let organization: string;
  export let project: string;
  export let onInvite: () => void = () => {};

  const userInvite = createAdminServiceAddProjectMember();

  const initialValues: {
    emails: string[];
    role: string;
  } = {
    emails: [],
    role: "viewer",
  };
  const schema = yup(
    object({
      emails: array(string().email("Invalid email")).min(1).max(2),
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

        await Promise.all(
          values.emails.map((email) => {
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
    },
  );
</script>

<form
  id="user-invite-form"
  on:submit|preventDefault={submit}
  class="w-full"
  use:enhance
>
  <MultiInput
    id="emails"
    placeholder="Invite by email, separated by commas"
    contentClassName="relative"
    bind:values={$form.emails}
    errors={$errors.emails}
  >
    <div
      slot="within-input"
      class="absolute right-0 top-0 h-full items-center flex"
    >
      <UserRoleSelect bind:value={$form.role} />
    </div>
    <Button
      submitForm
      type="primary"
      form="user-invite-form"
      slot="beside-input"
      disabled={$submitting}>Invite</Button
    >
  </MultiInput>
</form>
