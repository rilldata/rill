<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceGetOrganization,
    createAdminServiceUpdateOrganization,
    getAdminServiceGetOrganizationQueryKey,
  } from "@rilldata/web-admin/client";
  import SettingsItemContainer from "@rilldata/web-admin/features/organizations/settings/SettingsItemContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";

  export let organization: string;

  const initialValues: {
    name: string;
    description: string;
  } = {
    name: "",
    description: "",
  };
  const schema = yup(
    object({
      name: string()
        .required()
        .matches(/[_a-zA-Z0-9][-_a-zA-Z0-9 ]{2,39}/),
      description: string(),
    }),
  );

  function correctOrgName(org: string) {
    return org.replace(/ /g, "-");
  }

  const updateOrgMutation = createAdminServiceUpdateOrganization();

  const { form, errors, enhance, submit, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        const newOrg = correctOrgName(values.name);

        await $updateOrgMutation.mutateAsync({
          name: organization,
          data: {
            newName: correctOrgName(values.name),
            description: values.description,
          },
        });

        if (organization !== newOrg) {
          void goto(`/${newOrg}/-/settings`);
          queryClient.removeQueries(
            getAdminServiceGetOrganizationQueryKey(organization),
          );
        } else {
          void queryClient.refetchQueries(
            getAdminServiceGetOrganizationQueryKey(organization),
          );
        }
        eventBus.emit("notification", {
          message: "Updated organization",
        });
      },
      resetForm: false,
    },
  );

  $: orgResp = createAdminServiceGetOrganization(organization);
  $: if ($orgResp.data?.organization) {
    $form.name = $orgResp.data.organization.name;
    $form.description = $orgResp.data.organization.description;
  }

  $: changed =
    $orgResp.data?.organization?.name !== $form.name ||
    $orgResp.data?.organization?.description !== $form.description;
</script>

<SettingsItemContainer title="Organization">
  <form
    slot="description"
    id="org-update-form"
    on:submit|preventDefault={submit}
    class="w-full"
    use:enhance
  >
    <Input
      bind:value={$form.name}
      errors={$errors?.name}
      id="name"
      label="Name"
    />
    <div>
      Your org URL will be https://ui.rilldata.com/{correctOrgName($form.name)},
      to comply with our naming rules.
    </div>
    <Input
      bind:value={$form.description}
      errors={$errors?.description}
      id="description"
      label="Description"
      placeholder="Describe your organization"
    />
  </form>
  <Button
    on:click={submit}
    type="primary"
    loading={$submitting}
    disabled={!changed}
    slot="action"
  >
    Save
  </Button>
</SettingsItemContainer>
