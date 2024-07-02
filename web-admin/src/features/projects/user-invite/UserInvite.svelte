<script lang="ts">
  import { createAdminServiceAddProjectMember } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";

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
      console.log(values);
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

  const { form, errors, handleSubmit } = formState;
  $: console.log($errors);
</script>

<div class="flex flex-col">
  <form id="user-invite-form" on:submit|preventDefault={handleSubmit}>
    <h1>Invite teammates to your project</h1>
    <div>
      <div>Invite by email</div>
    </div>
    <div class="flex flex-row items-center gap-1.5">
      <div class="relative w-full">
        <MultiInput
          id="emails"
          label=""
          description=""
          accessorKey="email"
          {formState}
        />
        <div class="absolute right-0 top-0">
          <Select
            id="role"
            label=""
            bind:value={$form.role}
            options={[
              { value: "viewer", label: "can view" },
              { value: "admin", label: "can edit" },
            ]}
            className="w-24"
          />
        </div>
      </div>
      <Button submitForm type="primary" form="user-invite-form">Invite</Button>
    </div>
  </form>
</div>
