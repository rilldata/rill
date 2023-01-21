<script lang="ts">
  import { notifications } from "@rilldata/web-common/components/notifications";
  import { createForm } from "svelte-forms-lib";
  import type { Readable } from "svelte/store";
  import type { MetricsInternalRepresentation } from "../../../application-state-stores/metrics-internal-store";

  export let metricsInternalRep: Readable<MetricsInternalRepresentation>;

  $: displayName = $metricsInternalRep.getMetricKey("display_name");

  const { form, handleSubmit } = createForm({
    initialValues: {
      newDisplayName: displayName || "",
    },
    onSubmit: async (values) => {
      try {
        $metricsInternalRep.updateMetricKey(
          "display_name",
          values.newDisplayName
        );
      } catch (err) {
        console.error(err);
        notifications.send({ message: err.response.data.message });
      }
    },
  });

  function handleKeydown(event: KeyboardEvent) {
    if (event.code == "Enter") {
      event.preventDefault();
      handleSubmit(event);
      (event.target as HTMLInputElement).blur();
    }
  }

  function updateFormWithNewDisplayName(displayName: string) {
    $form.newDisplayName = displayName;
  }

  // This kicks in when the user changes the display name via code artifact
  $: updateFormWithNewDisplayName(displayName);
</script>

<div class="flex items-center mb-3">
  <div class="text-gray-500 font-medium" style="width:10em; font-size:11px;">
    Display name
  </div>
  <div>
    <form id="display-name-form" autocomplete="off">
      <input
        type="text"
        bind:value={$form["newDisplayName"]}
        on:keydown={handleKeydown}
        on:blur={handleSubmit}
        class="bg-white hover:bg-gray-100 rounded border border-6 border-transparent hover:border-gray-300 px-1"
        style="width:18em"
      />
    </form>
  </div>
</div>
