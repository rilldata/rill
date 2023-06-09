<script lang="ts">
  import { onDestroy, onMount, createEventDispatcher } from "svelte";
  import Custompicker from "./custom-picker";

  export let startEl, endEl, defaultStart, defaultEnd;
  let datepicker;
  let picker;

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
    picker = new Custompicker({
      element: datepicker,
      autoApply: false,
      autoRefresh: true,
      numberOfMonths: 2,
      numberOfColumns: 2,
      position: "bottom left",
      singleMode: false,
      startDate: defaultStart,
      endDate: defaultEnd,
      startEl,
      endEl,
    });

    picker.ui.addEventListener("click", (evt) => {
      evt.preventDefault();
      evt.stopPropagation();
    });
    picker.on("change", (dates) => {
      dispatch("change", dates);
    });

    picker.on("show", () => {
      dispatch("toggle", true);
    });

    picker.on("hide", () => {
      dispatch("toggle", false);
    });

    picker.on("editingDate", (v) => {
      editingDate = v;
      dispatch("editing", v);
    });

    startEl.addEventListener("focus", handleStartFocus);
    endEl.addEventListener("focus", handleEndFocus);
  });

  onDestroy(() => {
    startEl.removeEventListener("focus", handleStartFocus);
    endEl.removeEventListener("focus", handleEndFocus);
  });
</script>

<div bind:this={datepicker} class="w-0 h-0 absolute top-0 left-full" />

<style>
  :global(.litepicker) {
    --day-width: 42px;
    --day-height: 37px;
    --litepicker-tooltip-color-bg: #fff;
    --litepicker-month-header-color: #333;
    --litepicker-button-prev-month-color: #9e9e9e;
    --litepicker-button-next-month-color: #9e9e9e;
    --litepicker-button-prev-month-color-hover: #2a57e1;
    --litepicker-button-next-month-color-hover: #2a57e1;
    --litepicker-month-width: calc(var(--litepicker-day-width) * 7);
    --litepicker-day-width: 42px;
    --litepicker-day-color: #333;
    --litepicker-day-color-hover: #2a57e1;
    --litepicker-day-color-bg-hover: #b9d6fb;
    --litepicker-is-today-color: #333;
    --litepicker-is-in-range-color: #b9d6fb;
    --litepicker-is-start-color: #fff;
    --litepicker-is-start-color-bg: #3b76ee;
    --litepicker-is-end-color: #fff;
    --litepicker-is-end-color-bg: #3b76ee;
    --litepicker-start-border-radius: 5px 0px 0px 5px;
    --litepicker-end-border-radius: 0px 5px 5px 0px;
  }

  :global(.litepicker) {
    font-family: "Inter";
  }

  :global(.litepicker .container__footer) {
    display: none;
  }

  :global(.litepicker .month-item-weekdays-row) {
    font-size: 12px;
  }

  :global(.litepicker .month-item) {
    font-size: 15px;
  }

  :global(.litepicker .container__days .day-item) {
    height: var(--day-height);
    font-size: 12px;
    display: flex;
    justify-content: center;
    align-items: center;
    margin-bottom: 2px;
    position: relative;
    -webkit-transition: none;
    transition: none;
    cursor: default;
  }

  /* Triangles for current range boundaries */
  :global(.litepicker .day-item.is-start-date:not(.is-end-date):after) {
    border: 8px solid transparent;
    border-left: 8px solid var(--litepicker-is-start-color-bg);
    content: "";
    pointer-events: none;
    position: absolute;
    right: -14px;
    z-index: 1;
  }

  :global(.litepicker .day-item.is-end-date:not(.is-start-date):after) {
    border: 8px solid transparent;
    border-right: 8px solid var(--litepicker-is-end-color-bg);
    content: "";
    left: -14px;
    pointer-events: none;
    position: absolute;
    z-index: 1;
  }

  /* Triangles for proposed range boundaries */
  :global(
      .litepicker.editing-end .day-item.is-proposed-end:not(.is-end-date):after
    ) {
    position: absolute;
    width: 8px;
    height: 8px;
    border-top: 0px solid var(--litepicker-day-color-hover);
    border-right: 0px solid var(--litepicker-day-color-hover);
    border-bottom: 1.5px solid var(--litepicker-day-color-hover);
    border-left: 1.5px solid var(--litepicker-day-color-hover);
    top: 50%;
    right: 100%;
    margin-top: -4px;
    content: "";
    transform: rotate(45deg);
    margin-right: -4px;
    background: inherit;
  }

  :global(
      .litepicker.editing-start
        .day-item.is-proposed-start:not(.is-start-date):after
    ) {
    position: absolute;
    width: 8px;
    height: 8px;
    border-top: 1.5px solid var(--litepicker-day-color-hover);
    border-right: 1.5px solid var(--litepicker-day-color-hover);
    border-bottom: 0px solid var(--litepicker-day-color-hover);
    border-left: 0px solid var(--litepicker-day-color-hover);
    top: 50%;
    left: 100%;
    margin-top: -4px;
    content: "";
    transform: rotate(45deg);
    margin-left: -4px;
    background: inherit;
  }

  /* Use !important to override the Litepicker :hover styles */
  :global(
      .litepicker .day-item.is-proposed-start:not(.is-start-date),
      .litepicker .day-item.is-proposed-end:not(.is-end-date)
    ) {
    background: white;
    -webkit-box-shadow: inset 0 0 0 1.5px var(--litepicker-day-color-hover) !important;
    box-shadow: inset 0 0 0 1.5px var(--litepicker-day-color-hover) !important;
    z-index: 3;
    color: inherit !important;
    font-weight: bold;
  }

  /* Disable Litepicker hover styles */
  :global(
      .litepicker .day-item.is-start-date:hover,
      .litepicker .day-item.is-end-date:hover
    ) {
    box-shadow: none;
    -webkit-box-shadow: none;
  }

  :global(
      .litepicker
        .container__days
        .day-item.is-proposed-start:not(.is-start-date)
    ) {
    border-radius: var(--litepicker-start-border-radius) !important;
  }

  :global(
      .litepicker .container__days .day-item.is-proposed-end:not(.is-end-date)
    ) {
    border-radius: var(--litepicker-end-border-radius) !important;
  }

  :global(.litepicker .day-item.is-in-proposed-range) {
    border-radius: 0px;
    background: #f0f7fe;
    /* Custom dashed line with svg */
    background-image: url("data:image/svg+xml,<svg width='100%' height='100%' xmlns='http://www.w3.org/2000/svg'><rect width='200%' x='-20px' height='100%' fill='none' stroke='%232A57E1' stroke-width='2' stroke-dasharray='2%2c6' stroke-dashoffset='0' stroke-linecap='square'/></svg>");
  }

  :global(.litepicker .day-item.is-in-range.is-in-proposed-range) {
    /* Custom dashed line with svg, darker */
    background-image: url("data:image/svg+xml,<svg width='100%' height='100%' xmlns='http://www.w3.org/2000/svg'><rect width='200%' x='-20px' height='100%' fill='none' stroke='%233A6AE5' stroke-width='2' stroke-dasharray='2%2c6' stroke-dashoffset='0' stroke-linecap='square'/></svg>");
  }

  :global(.litepicker .container__tooltip) {
    z-index: 1000;
  }
</style>
