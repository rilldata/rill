<script lang="ts">
  import { onDestroy, onMount, createEventDispatcher } from "svelte";
  import { easepick } from "@easepick/core";
  import { RangePlugin } from "./range-plugin";

  let datepicker;
  let picker;
  export let startEl, endEl, defaultStart, defaultEnd;
  let editingDate = 0;

  const dispatch = createEventDispatcher();

  const handleStartFocus = () => {
    if (picker) {
      picker.show();
      picker.setEditingDate(0);
    }
  };

  const handleEndFocus = () => {
    if (picker) {
      picker.show();
      picker.setEditingDate(1);
    }
  };

  onMount(() => {
    picker = new easepick.create({
      element: datepicker,
      calendars: 2,
      grid: 2,
      css: [
        "https://cdn.jsdelivr.net/npm/@easepick/bundle@1.2.1/dist/index.css",
        "https://cdn.jsdelivr.net/npm/@easepick/lock-plugin@1.2.1/dist/index.css",
        //Set custom css
        //'/css/calendar.css
      ],
      zIndex: 10,
      plugins: [RangePlugin],
      inline: false,
      autoApply: true,
      format: "MM/DD/YYYY",
      RangePlugin: {
        startEl,
        endEl,
        startDate: defaultStart,
        endDate: defaultEnd,
      },
    });

    picker.on("editingDate", (v) => {
      editingDate = v.detail;
      dispatch("editing", v.detail);
    });

    picker.on("select", (d) => {
      dispatch("change", d.detail);
    });

    picker.on("hide", (d) => {
      dispatch("toggle", false);
    });

    picker.on("show", (d) => {
      dispatch("toggle", true);
    });

    startEl.addEventListener("focus", handleStartFocus);
    endEl.addEventListener("focus", handleEndFocus);
  });

  onDestroy(() => {
    startEl.removeEventListener("focus", handleStartFocus);
    endEl.removeEventListener("focus", handleEndFocus);
  });
</script>

<div bind:this={datepicker} />
