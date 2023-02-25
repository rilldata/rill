<script lang="ts">
  import Dialog from "@rilldata/web-common/components/modal/dialog/Dialog.svelte";
  import { calendlyModalStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { behaviourEvent } from "@rilldata/web-local/lib/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
  import { onMount, tick } from "svelte";

  let calendlyContainer: HTMLElement;

  const Calendly = (window as any).Calendly;
  onMount(async () => {
    // This is needed because Dialog has Portal internally.
    // That removes the component DOM and re-adds to the end.
    // the await tick will make sure `Calendly.initInlineWidget` runs after the portal
    await tick();
    Calendly.initInlineWidget({
      url: "https://calendly.com/marissa-gorlick/rill-closed-beta-discovery",
      parentElement: calendlyContainer,
      prefill: {},
      utm: {},
    });

    function eventScheduledListener(e) {
      if (
        e.origin !== "https://calendly.com" ||
        !e.data?.event?.startsWith("calendly.")
      )
        return;
      if (e.data.event === "calendly.event_scheduled") {
        behaviourEvent.firePublishEvent(
          $calendlyModalStore,
          BehaviourEventMedium.Button,
          MetricsEventSpace.Workspace,
          MetricsEventScreenName.Dashboard,
          MetricsEventScreenName.Dashboard,
          false
        );
      }
    }

    window.addEventListener("message", eventScheduledListener);

    return () => {
      window.removeEventListener("message", eventScheduledListener);
    };
  });

  function closeCalendly() {
    calendlyModalStore.set("");
  }
</script>

<Dialog on:cancel={closeCalendly} size="full" yFixed={true}>
  <div bind:this={calendlyContainer} class="h-full" id="calendly" slot="body" />
</Dialog>
