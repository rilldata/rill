<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  export let orgName: string = "";
  export let orgDisplayName: string = "";
  export let onUpdate: (
    orgName: string,
    orgDisplayName: string | undefined,
  ) => void;

  const initialValues: {
    name: string;
    displayName: string;
  } = {
    name: orgName,
    displayName: orgDisplayName,
  };
  const schema = yup(
    object({
      name: string().required(),
      displayName: string(),
    }),
  );

  const { form, errors, enhance, submit } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        onUpdate(values.name, values.displayName);
      },
    },
  );
</script>

<div class="text-xl">Let’s create your first organization</div>
<div class="text-base text-gray-500">
  Create an organization to deploy this project to. <a
    href="https://docs.rilldata.com/reference/cli/org/create"
    target="_blank">See docs</a
  >
</div>

<form id="org-name-form" on:submit|preventDefault={submit} use:enhance>
  <Input
    bind:value={$form.displayName}
    errors={$errors?.displayName}
    id="displayName"
    label="Organization display name"
    textClass="text-sm"
  />
  <Input
    bind:value={$form.name}
    errors={$errors?.name}
    id="name"
    label="URL"
    textClass="text-sm"
    alwaysShowError
  >
    <div slot="icon">https://ui.rilldata.com/</div>
  </Input>
</form>
<Button wide type="primary" on:click={submit}>Continue</Button>
