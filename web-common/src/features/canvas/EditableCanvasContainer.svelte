<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import BlockGrip from "./BlockGrip.svelte";
  import BlockMenuItems from "./BlockMenuItems.svelte";
  import EditableBlockRegion, {
    type BlockBuilderProps,
  } from "./EditableBlockRegion.svelte";
  import { containerRowsPath } from "./layout-util";
  import { isGeneratedBlockName } from "./stores/block-edit";
  import type { CanvasEntity } from "./stores/canvas-entity";
  import type { RowsBlock } from "./stores/layout-blocks";
  import VisibilityBadge from "./VisibilityBadge.svelte";
  import LevelControls from "./controls/LevelControls.svelte";

  // A container in the editor: the shared block region in its dashed variant, with a chip
  // in the header band that names the container and holds its grip, visibility badge and
  // menu. Viewers never see the frame or the chip.
  // svelte-ignore custom_element_props_identifier
  let {
    container,
    canvas,
    selected,
    onUngroup,
    onDeleteBlock,
    onSelectBlock,
    onBlockMouseDown,
    onDuplicateBlock,
    ...builder
  }: BlockBuilderProps & {
    container: RowsBlock;
    canvas: CanvasEntity;
    selected: boolean;
    // Replace the container with its rows.
    onUngroup: (blockIndex: number) => void;
    onDeleteBlock: (blockIndex: number) => void;
  } = $props();

  const blockIndex = $derived(builder.blockIndex);
  // The chip shows a name the author chose, not a generated one such as "container-2".
  // Renaming happens in the inspector, which a click on the chip opens.
  const name = $derived(
    isGeneratedBlockName("container", container.name) ? "" : container.name,
  );

  let grip = $state<HTMLDivElement>();

  function select() {
    onSelectBlock(container.name);
  }

  // Pressing the grip or the chip's own background starts a drag of the whole container;
  // the menu keeps its own behavior.
  function handleChipMouseDown(event: MouseEvent) {
    const target = event.target;
    if (!(target instanceof Element)) return;
    if (target.closest("button")) return;
    if (target === event.currentTarget || grip?.contains(target)) {
      onBlockMouseDown(blockIndex, event);
    }
  }
</script>

<EditableBlockRegion
  {...builder}
  variant="dashed"
  grid={container.grid}
  target={{ rowsPath: containerRowsPath(blockIndex) }}
  idPrefix={`c${blockIndex}-`}
  zoneScope={`container:${container.name}`}
  rowKey={container.name}
  emptyLabel={m.canvas_add_widget_to_container()}
  afterLabel={m.canvas_add_widget_below_container()}
  {selected}
  onSelect={select}
>
  {#snippet header()}
    <div class="header">
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <div
        class={["chip", { selected }]}
        role="presentation"
        onmousedown={handleChipMouseDown}
        onclick={select}
      >
        <BlockGrip kind="container" bind:element={grip} />
        <span class="label" title={m.canvas_container_viewer_note()}>
          {m.canvas_container()}
        </span>
        {#if name}
          <span class="name">{name}</span>
        {/if}
        <VisibilityBadge {canvas} path={["rows", blockIndex]} inline />
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            {#snippet child({ props })}
              <button
                {...props}
                class="menu-trigger"
                title={m.canvas_container_options()}
                aria-label={m.canvas_container_options()}
              >
                <ThreeDot size="14px" />
              </button>
            {/snippet}
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="start" class="min-w-[180px]">
            <BlockMenuItems
              kind="container"
              onDuplicate={() => onDuplicateBlock(blockIndex)}
              onUngroup={() => onUngroup(blockIndex)}
              onDelete={() => onDeleteBlock(blockIndex)}
            />
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </div>
      <!-- The container's viewer controls, which viewers see above its rows. -->
      <LevelControls canvasEntity={canvas} scopes={[[container.name]]} />
    </div>
  {/snippet}
</EditableBlockRegion>

<style lang="postcss">
  /* The band above the rows that holds the chip; px-2.5 lines the chip up with the cards. */
  .header {
    @apply flex min-h-7 w-full flex-none items-center px-2.5;
  }

  .chip {
    @apply flex h-5 items-center gap-x-1 rounded-sm border border-gray-200 bg-surface-card pl-0.5 pr-0.5;
    @apply cursor-grab text-xs font-medium text-fg-secondary;
  }

  .chip:hover {
    @apply border-gray-300 text-fg-primary;
  }

  .chip.selected {
    @apply border-primary-400 text-fg-primary;
  }

  .name {
    @apply max-w-60 truncate font-normal text-fg-primary;
  }

  .menu-trigger {
    @apply grid size-4 place-content-center rounded-sm text-fg-secondary hover:text-fg-primary;
    @apply cursor-pointer;
  }
</style>
