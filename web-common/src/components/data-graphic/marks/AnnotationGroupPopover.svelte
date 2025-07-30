<script lang="ts">
  import type { AnnotationGroup } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
  import * as Popover from "@rilldata/web-common/components/popover";
  import { builderActions, getAttrs } from "bits-ui";

  export let annotationGroup: AnnotationGroup;

  $: popoverOffset = 8 + annotationGroup.right - annotationGroup.left;
</script>

<div class="relative">
  <Popover.Root open>
    <Popover.Trigger asChild let:builder>
      <button
        class="absolute bottom-2 w-0 h-0"
        style="left: {annotationGroup.left}px;"
        {...getAttrs([builder])}
        use:builderActions={{ builders: [builder] }}
      ></button>
    </Popover.Trigger>
    <Popover.Content side="right" sideOffset={popoverOffset}>
      <div class="flex flex-col gap-y-2">
        {#each annotationGroup.items as annotation, i (i)}
          <div class="flex flex-col gap-y-1">
            <div>{annotation.description}</div>
            <div>{annotation.startTime.toISOString()}</div>
          </div>
        {/each}
      </div>
    </Popover.Content>
  </Popover.Root>
</div>
