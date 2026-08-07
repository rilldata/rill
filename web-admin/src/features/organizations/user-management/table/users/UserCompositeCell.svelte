<script lang="ts">
  import AvatarListItem from "@rilldata/web-common/components/avatar/AvatarListItem.svelte";
  import { Chip } from "@rilldata/web-common/components/chip";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let name: string;
  export let email: string;
  export let isCurrentUser: boolean;
  export let pendingAcceptance: boolean;
  export let photoUrl: string | null;
  export let role: string;
  export let attributes: Record<string, unknown> | undefined = undefined;

  $: attributeEntries = Object.entries(attributes ?? {});

  function formatAttributeValue(value: unknown): string {
    return typeof value === "string" ? value : JSON.stringify(value);
  }
</script>

<div class="flex items-center gap-x-2">
  <AvatarListItem
    {name}
    {email}
    {isCurrentUser}
    {pendingAcceptance}
    {photoUrl}
    {role}
  />
  {#if attributeEntries.length > 0}
    <Tooltip location="right" alignment="start" distance={8}>
      <Chip
        type="dimension"
        label={m.users_attributes_count({ count: attributeEntries.length })}
        compact
        readOnly
      >
        <svelte:fragment slot="body">
          {m.users_attributes_count({ count: attributeEntries.length })}
        </svelte:fragment>
      </Chip>
      <TooltipContent slot="tooltip-content">
        <ul>
          {#each attributeEntries as [key, value] (key)}
            <li class="text-xs py-0.5">{key}: {formatAttributeValue(value)}</li>
          {/each}
        </ul>
      </TooltipContent>
    </Tooltip>
  {/if}
</div>
