import { Collapsible as CollapsiblePrimitive } from "bits-ui";
import Content from "web-common/src/components/collapsible/collapsible-content.svelte";
import Trigger from "web-common/src/components/collapsible/collapsible-trigger.svelte";

const Root = CollapsiblePrimitive.Root;

export {
  Root,
  Content,
  Trigger,
  //
  Root as Collapsible,
  Content as CollapsibleContent,
  Trigger as CollapsibleTrigger,
};
