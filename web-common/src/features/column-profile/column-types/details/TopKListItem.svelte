<script lang="ts">
  import BarAndLabel from "@rilldata/web-common/components/BarAndLabel.svelte";

  export let value: number; // should be between 0 and 1.
  export let color = "bg-primary-200";
  export let onFocus: (() => void) | undefined = undefined;
  export let onBlur: (() => void) | undefined = undefined;

  /** compact mode is used in e.g. profiles */

  const onHover = () => {
    onFocus?.();
  };
  const onLeave = () => {
    onBlur?.();
  };
</script>

<button
  class="block flex flex-row w-full text-left transition-color"
  onblur={onLeave}
  onfocus={onHover}
  onmouseleave={onLeave}
  onmouseover={onHover}
  transitionslide={{ duration: 200 }}
>
  <BarAndLabel
    {color}
    justify={false}
    showBackground={false}
    showHover
    tweenParameters={{ duration: 200 }}
    {value}
  >
    <div
      class="grid items-center gap-x-3"
      style="grid-template-columns: auto max-content; height: 18px;"
    >
      <div
        class="justify-self-start text-left w-full text-ellipsis overflow-hidden whitespace-nowrap"
      >
        <slot name="title" />
      </div>
      <div
        class="justify-self-end overflow-hidden ui-copy-number flex gap-x-4 items-baseline"
      >
        <slot name="right" />
      </div>
    </div>
  </BarAndLabel>
</button>
