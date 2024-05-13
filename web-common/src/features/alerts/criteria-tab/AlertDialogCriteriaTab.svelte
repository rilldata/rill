<script lang="ts">
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import AlertPreview from "@rilldata/web-common/features/alerts/criteria-tab/AlertPreview.svelte";
  import CriteriaGroup from "@rilldata/web-common/features/alerts/criteria-tab/CriteriaGroup.svelte";
  import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { createForm } from "svelte-forms-lib";

  export let formState: ReturnType<typeof createForm<AlertFormValues>>;

  const { form } = formState;
</script>

<div class="flex flex-col gap-y-3">
  <FormSection
    description="Trigger alert when these conditions are met"
    title="Criteria"
  >
    <CriteriaGroup {formState} />
  </FormSection>
  <FormSection title="Alert Preview">
    <AlertPreview formValues={$form} />
  </FormSection>
</div>
