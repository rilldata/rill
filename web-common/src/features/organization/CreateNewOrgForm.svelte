<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { sanitizeOrgName } from "@rilldata/web-common/features/organization/sanitizeOrgName";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  export let isFirstOrg: boolean;
  export let onUpdate: (
    orgName: string,
    orgDisplayName: string | undefined,
  ) => void;

  const initialValues: {
    name: string;
    displayName: string;
  } = {
    name: "",
    displayName: "",
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

<div class="text-xl">
  {#if isFirstOrg}
    Let’s create your first organization
  {:else}
    Create a new organization
  {/if}
</div>
<div class="text-base text-gray-500">
  Create an organization to deploy this project to. <a
    href="https://docs.rilldata.com/reference/cli/org/create"
    target="_blank">See docs</a
  >
</div>

<form
  id="org-name-form"
  on:submit|preventDefault={submit}
  use:enhance
  class="flex flex-col gap-y-3"
>
  <Input
    bind:value={$form.displayName}
    errors={$errors?.displayName}
    id="displayName"
    label="Organization display name"
    textClass="text-sm"
    width="500px"
  />
  <Input
    bind:value={$form.name}
    errors={$errors?.name}
    id="name"
    label="URL"
    textClass="text-sm"
    alwaysShowError
    width="500px"
  >
    <div
      slot="extra-content"
      class="bg-neutral-100 text-neutral-400 border border-r-0 border-gray-300 text-base px-2 py-1 max-h-8"
    >
      https://ui.rilldata.com/
    </div>
  </Input>
  <div class="text-xs">
    Must comply with <a
      href="https://docs.rilldata.com/reference/cli/org/create#TODO"
      target="_blank">our naming rules.</a
    >
  </div>
</form>
<Button
  wide
  forcedStyle="min-width:500px !important;"
  type="primary"
  on:click={submit}
>
  Continue
</Button>
