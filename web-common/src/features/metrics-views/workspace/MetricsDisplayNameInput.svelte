<script lang="ts">
  import { notifications } from "@rilldata/web-common/components/notifications";
  import { createForm } from "svelte-forms-lib";
  import type { Readable } from "svelte/store";
  import Tooltip from "../../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../../components/tooltip/TooltipContent.svelte";
  import type { MetricsInternalRepresentation } from "../metrics-internal-store";

  export let metricsInternalRep: Readable<MetricsInternalRepresentation>;

  $: displayName = $metricsInternalRep.getMetricKey("display_name");

  const { form, handleSubmit } = createForm({
    initialValues: {
      newDisplayName: displayName || "",
    },
    onSubmit: async (values) => {
      try {
        $metricsInternalRep.updateMetricsParams({
          display_name: values.newDisplayName,
        });
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

<div class="w-80 flex items-center">
  <Tooltip alignment="middle" distance={8} location="bottom">
    <div class="text-gray-500 font-medium" style="width:10em; font-size:11px;">
      Display Name
    </div>

    <TooltipContent slot="tooltip-content">
      Add a title to your dashboard
    </TooltipContent>
  </Tooltip>
  <div>
    <form id="display-name-form" autocomplete="off">
      <input
        type="text"
        bind:value={$form["newDisplayName"]}
        on:keydown={handleKeydown}
        on:blur={handleSubmit}
        class="bg-white hover:bg-gray-100 rounded border border-6 border-transparent hover:border-gray-300 px-1"
      />
    </form>
  </div>
</div>
