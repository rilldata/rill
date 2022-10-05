<script lang="ts">
  import IconButton from "../../button/IconButton.svelte";

  import { createEventDispatcher } from "svelte";
  import Close from "../../icons/Close.svelte";
  import ModalContainer from "../ModalContainer.svelte";
  import DialogCTA from "./DialogCTA.svelte";
  import DialogFooter from "./DialogFooter.svelte";
  import DialogHeader from "./DialogHeader.svelte";
  export let dark = false;
  export let compact = false;
  export let showCancel = true;
  export let disabled = false;
  export let location: "top" | "center" = "center";
  export let minHeight: string = undefined;

  const dispatch = createEventDispatcher();

  export let minWidth: string = undefined;

  $: containerClasses = dark
    ? "text-white bg-gray-800"
    : "text-gray-800 bg-white";
</script>

<ModalContainer on:cancel>
  <div
    class:content-center={location === "center"}
    class:content-start={location === "top"}
    class="justify-center grid w-screen h-screen relative"
  >
    <div
      style:min-height={minHeight}
      class:min-width={minWidth}
      class="{minWidth ? '' : 'min-w-[400px]'} {containerClasses} rounded"
      style:transform={location === "center" && "translateY(-120px)"}
      class:m-24={location === "top"}
    >
      <DialogHeader {compact}>
        <svelte:fragment slot="title"><slot name="title" /></svelte:fragment>
        <div slot="right">
          {#if showCancel}
            <!-- FIXME: this should be replaced with the IconButton in an open PR -->
            <IconButton
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
      <hr />
      <div class={compact ? "px-4 py-8" : "px-7 pt-8 pb-16"}>
        <slot name="body" />
      </div>
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
