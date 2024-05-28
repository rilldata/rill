import { Dialog as DialogPrimitive } from "bits-ui";

import Title from "web-common/src/components/dialog-v2/dialog-title.svelte";
import Portal from "web-common/src/components/dialog-v2/dialog-portal.svelte";
import Footer from "web-common/src/components/dialog-v2/dialog-footer.svelte";
import Header from "web-common/src/components/dialog-v2/dialog-header.svelte";
import Overlay from "web-common/src/components/dialog-v2/dialog-overlay.svelte";
import Content from "web-common/src/components/dialog-v2/dialog-content.svelte";
import Description from "web-common/src/components/dialog-v2/dialog-description.svelte";

const Root = DialogPrimitive.Root;
const Trigger = DialogPrimitive.Trigger;
const Close = DialogPrimitive.Close;

export {
  Root,
  Title,
  Portal,
  Footer,
  Header,
  Trigger,
  Overlay,
  Content,
  Description,
  Close,
  //
  Root as Dialog,
  Title as DialogTitle,
  Portal as DialogPortal,
  Footer as DialogFooter,
  Header as DialogHeader,
  Trigger as DialogTrigger,
  Overlay as DialogOverlay,
  Content as DialogContent,
  Description as DialogDescription,
  Close as DialogClose,
};
