<script context="module" lang="ts">
  export const CreateNewOrgFormId = "org-name-form";
</script>

<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    createLocalServiceCreateOrganization,
    getLocalServiceGetCurrentUserQueryKey,
  } from "@rilldata/web-common/runtime-client/local-service";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  export let onCreate: (orgName: string) => void;

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

  const orgCreator = createLocalServiceCreateOrganization();

  const { form, errors, enhance, submit } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      onUpdate: async ({ form }) => {
        if (!form.valid) return;
        const values = form.data;

        await $orgCreator.mutateAsync({
          name: values.name,
          displayName: values.displayName,
        });

        await queryClient.invalidateQueries({
          queryKey: getLocalServiceGetCurrentUserQueryKey(),
        });
        onCreate(values.name);
      },
      onError({ result }) {
        if (
          result.error.message.includes("an org with that name already exists")
        ) {
          $errors["name"] = [
            "Org already exists. Please choose a different name.",
          ];
        }
      },
    },
  );
</script>

<form
  id={CreateNewOrgFormId}
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
