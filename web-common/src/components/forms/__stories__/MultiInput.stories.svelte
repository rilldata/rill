<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import { Meta, Story } from "@storybook/addon-svelte-csf";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";

  const formState = createForm({
    initialValues: {
      emails: [],
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
    },
  });

  const { handleSubmit } = formState;
</script>

<Meta title="Multiple Input" />

<Story name="Form with Multiple Input">
  <form
    autocomplete="off"
    class="flex flex-col gap-y-6 w-[300px]"
    id="multi-input-form"
    on:submit|preventDefault={handleSubmit}
  >
    <MultiInput id="emails" label="Emails" accessorKey="email" {formState} />
    <Button form="multi-input-form" submitForm type="primary">Invite</Button>
  </form>
</Story>
