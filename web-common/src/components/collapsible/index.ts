import { Collapsible as CollapsiblePrimitive } from "bits-ui";
import Content from "web-common/src/components/collapsible/collapsible-content.svelte";

const Root = CollapsiblePrimitive.Root;
const Trigger = CollapsiblePrimitive.Trigger;

export {
  Root,
  Content,
  Trigger,
  //
  Root as Collapsible,
  Content as CollapsibleContent,
  Trigger as CollapsibleTrigger,
};
