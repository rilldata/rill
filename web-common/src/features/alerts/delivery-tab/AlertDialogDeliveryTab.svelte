<script lang="ts">
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { SnoozeOptions } from "@rilldata/web-common/features/alerts/delivery-tab/snooze";
  import RecipientsInputArray from "@rilldata/web-common/features/scheduled-reports/RecipientsInputArray.svelte";

  export let formState: any; // svelte-forms-lib's FormState

  const { form } = formState;
</script>

<div class="flex flex-col gap-y-3">
  <FormSection
    description="We'll check for this alert whenever the data refreshes"
    title="Schedule"
  />
  <FormSection
    description="Set a snooze period to silence repeat notifications for the same alert."
    title="Snooze"
  >
    <Select
      bind:value={$form["snooze"]}
      id="snooze"
      label=""
      options={SnoozeOptions}
    />
  </FormSection>
  <FormSection
    description="Choose who will get notified by email for this alert. Make sure they have access to your project."
    title="Recipients"
  >
    <RecipientsInputArray {formState} showLabel={false} />
  </FormSection>
</div>
