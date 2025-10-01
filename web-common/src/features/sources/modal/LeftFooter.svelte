<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";

  export let onBack: () => void;
  export let disabled: boolean = false;
  export let loading: boolean = false;
  export let loadingCopy: string = "Testing connection...";
  export let formId: string;
  export let submitButtonText: string;
</script>

<div
  class="w-full bg-white border-t border-gray-200 p-6 flex justify-between gap-2"
>
  <Button onClick={onBack} type="secondary">Back</Button>

  <Button
    {disabled}
    {loading}
    {loadingCopy}
    form={formId}
    submitForm
    type="primary"
  >
    {submitButtonText}
  </Button>
</div>
