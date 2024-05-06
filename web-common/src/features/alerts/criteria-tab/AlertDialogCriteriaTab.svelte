<script lang="ts">
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import AlertPreview from "@rilldata/web-common/features/alerts/criteria-tab/AlertPreview.svelte";
  import CriteriaGroup from "@rilldata/web-common/features/alerts/criteria-tab/CriteriaGroup.svelte";
  import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { createForm } from "svelte-forms-lib";

  export let formState: ReturnType<typeof createForm<AlertFormValues>>;

  const { form } = formState;
</script>

<div class="flex flex-col gap-y-3">
  <FormSection description="Trigger alert when these conditions are met">
    <CriteriaGroup {formState} />
    <div slot="title" class="flex flex-row w-full">
      <div class="w-full">Criteria</div>
      <div class="flex flex-row gap-2 pr-2 text-base items-center">
        Not <Switch bind:checked={$form["criteriaIsNot"]} />
      </div>
    </div>
  </FormSection>
  <FormSection title="Alert Preview">
    <AlertPreview formValues={$form} />
  </FormSection>
</div>
