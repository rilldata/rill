import { Tooltip as TooltipPrimitive } from "bits-ui";
import Content from "web-common/src/components/tooltip-v2/tooltip-content.svelte";
import Trigger from "web-common/src/components/tooltip-v2/tooltip-trigger.svelte";

const Root = TooltipPrimitive.Root;
const Provider = TooltipPrimitive.Provider;

export {
  Root,
  Trigger,
  Content,
  Provider,
  //
  Root as Tooltip,
  Content as TooltipContent,
  Trigger as TooltipTrigger,
  Provider as TooltipProvider,
};
