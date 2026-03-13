<script lang="ts">
  import { Collapsible as CollapsiblePrimitive } from "bits-ui";
  type $$Props = CollapsiblePrimitive.ContentProps;
</script>

<CollapsiblePrimitive.Content {...$$restProps}>
  <slot />
</CollapsiblePrimitive.Content>
