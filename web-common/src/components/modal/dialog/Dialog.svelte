<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import IconButton from "../../button/IconButton.svelte";
  import Close from "../../icons/Close.svelte";
  import ModalContainer from "../ModalContainer.svelte";
  import DialogCTA from "./DialogCTA.svelte";
  import DialogFooter from "./DialogFooter.svelte";
  import DialogHeader from "./DialogHeader.svelte";

  export let dark = false;
  export let compact = false; // refers to padding
  export let size: "sm" | "md" | "lg" | "full" = "md";
  export let yFixed = false;
  export let showCancel = true;
  export let disabled = false;
  export let useContentForMinSize = false;

  const dispatch = createEventDispatcher();

  $: containerClasses = dark
    ? "text-white bg-gray-800"
    : "text-gray-800 bg-white";

  let xDimClasses = "";
  let yDimClasses = "";
  let width = "";
  let height = "";
  $: {
    switch (size) {
      case "sm":
        xDimClasses = "w-1/2 md:w-1/3 xl:w-1/4 2xl:w-1/5";
        yDimClasses = yFixed ? "h-1/3" : "";
        break;

      case "md":
        xDimClasses = "w-2/3 md:w-2/3 xl:w-1/3 2xl:w-1/3 max-w-2xl";
        yDimClasses = yFixed ? "h-1/2" : "";
        break;

      case "lg":
        xDimClasses = "w-4/5 md:w-3/5 xl:w-1/2 2xl:w-1/3";
        yDimClasses = yFixed ? "h-3/5" : "";
        break;

      case "full":
        xDimClasses = "w-4/5 md:w-3/5 xl:w-1/2 2xl:w-2/3";
        yDimClasses = yFixed ? "h-4/5" : "";
        break;
    }
  }
</script>

<ModalContainer on:cancel>
  <div
    class="grid place-items-center w-screen h-screen"
    on:click|self={() => {
      dispatch("cancel");
    }}
    on:keydown={() => {
      /** no op for now */
    }}
  >
    <div
      class="{containerClasses} {xDimClasses} {yDimClasses} rounded pointer-events-auto flex flex-col"
      class:min-w-max={useContentForMinSize}
      style:height
      style:transform={yFixed ? "" : "translateY(-120px)"}
      style:width
    >
      <DialogHeader {compact}>
        <svelte:fragment slot="title"><slot name="title" /></svelte:fragment>
        <div slot="right">
          {#if showCancel}
            <!-- FIXME: this should be replaced with the IconButton in an open PR -->
            <IconButton
              marginClasses="ml-3"
              on:click={() => {
                dispatch("cancel");
              }}
            >
              <Close size="16px" />
            </IconButton>
          {:else}
            <slot name="right" />
          {/if}
        </div>
      </DialogHeader>
      {#if $$slots.title}
        <hr />
      {/if}
      <slot />
      {#if $$slots.body}
        <div
          class="overflow-y-auto flex-grow
        {compact ? 'px-4 py-8' : 'px-7 pt-8 pb-16'}"
        >
          <slot name="body" />
        </div>
      {/if}
      {#if $$slots.footer}
        <DialogFooter>
          <slot name="footer" />
        </DialogFooter>
      {:else if $$slots["primary-action-body"]}
        <DialogFooter>
          <DialogCTA
            {disabled}
            on:cancel
            on:primary-action
            on:secondary-action
            showSecondary={$$slots["secondary-action-body"]}
          >
            <svelte:fragment slot="secondary-action-body"
              ><slot name="secondary-action-body" /></svelte:fragment
            >
            <svelte:fragment slot="primary-action-body"
              ><slot name="primary-action-body" /></svelte:fragment
            >
          </DialogCTA>
        </DialogFooter>
      {/if}
    </div>
  </div>
</ModalContainer>
