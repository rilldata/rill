<script lang="ts">
  import Dialog from "@rilldata/web-common/components/modal/dialog/Dialog.svelte";
  import { showCalendlyModal } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { onMount, tick } from "svelte";

  let calendlyContainer: HTMLElement;

  const Calendly = (window as any).Calendly;
  onMount(async () => {
    // This is needed because Dialog has Portal internally.
    // That removes the component DOM and re-adds to the end.
    // the await tick will make sure `Calendly.initInlineWidget` runs after the portal
    await tick();
    Calendly.initInlineWidget({
      url: "https://calendly.com/marissa-gorlick/rill-closed-beta-discovery?month=2023-02",
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
      switch (e.data.event) {
        case "calendly.event_scheduled":
          break;
      }
    }

    window.addEventListener("message", eventScheduledListener);

    return () => {
      window.removeEventListener("message", eventScheduledListener);
    };
  });

  function closeCalendly() {
    showCalendlyModal.set(false);
  }
</script>

<Dialog on:cancel={closeCalendly} size="full" yFixed={true}>
  <div bind:this={calendlyContainer} class="h-full" id="calendly" slot="body" />
</Dialog>
