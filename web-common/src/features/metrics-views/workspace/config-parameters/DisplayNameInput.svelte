<script lang="ts">
  import { notifications } from "@rilldata/web-common/components/notifications";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createForm } from "svelte-forms-lib";
  import type { Readable } from "svelte/store";
  import type { MetricsInternalRepresentation } from "../../metrics-internal-store";
  import {
    CONFIG_SELECTOR,
    CONFIG_TOP_LEVEL_LABEL_CLASSES,
    INPUT_ELEMENT_CONTAINER,
    SELECTOR_BUTTON_TEXT_CLASSES,
    SELECTOR_CONTAINER,
  } from "../styles";

  export let metricsInternalRep: Readable<MetricsInternalRepresentation>;

  $: displayName = $metricsInternalRep.getMetricKey("display_name");

  $: innerDisplayName = displayName;

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
  $: updateFormWithNewDisplayName(innerDisplayName);
</script>

<div
  class={INPUT_ELEMENT_CONTAINER.classes}
  style={INPUT_ELEMENT_CONTAINER.style}
>
  <Tooltip alignment="middle" distance={8} location="bottom">
    <div class={CONFIG_TOP_LEVEL_LABEL_CLASSES}>Display Name</div>

    <TooltipContent slot="tooltip-content">
      Add a title to your dashboard
    </TooltipContent>
  </Tooltip>
  <div class={SELECTOR_CONTAINER.classes} style={SELECTOR_CONTAINER.style}>
    <form id="display-name-form" autocomplete="off">
      <input
        type="text"
        bind:value={$form["newDisplayName"]}
        on:keydown={handleKeydown}
        on:blur={handleSubmit}
        placeholder={"Inferred from model"}
        class="{SELECTOR_BUTTON_TEXT_CLASSES} placeholder:font-normal placeholder:text-gray-600 font-semibold bg-white w-full hover:bg-gray-200 rounded border border-6 border-gray-200 hover:border-gray-300  hover:text-gray-900 px-2 py-1 h-[34px] {CONFIG_SELECTOR.focus}"
      />
    </form>
  </div>
</div>
