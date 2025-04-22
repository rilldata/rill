import { Dialog as DialogPrimitive } from "bits-ui";

import Content from "web-common/src/components/dialog/dialog-content.svelte";
import Description from "web-common/src/components/dialog/dialog-description.svelte";
import Footer from "web-common/src/components/dialog/dialog-footer.svelte";
import Header from "web-common/src/components/dialog/dialog-header.svelte";
import Overlay from "web-common/src/components/dialog/dialog-overlay.svelte";
import Portal from "web-common/src/components/dialog/dialog-portal.svelte";
import Title from "web-common/src/components/dialog/dialog-title.svelte";

const Root = DialogPrimitive.Root;
const Trigger = DialogPrimitive.Trigger;
const Close = DialogPrimitive.Close;

export {
  Close,
  Content,
  Description,
  //
  Root as Dialog,
  Close as DialogClose,
  Content as DialogContent,
  Description as DialogDescription,
  Footer as DialogFooter,
  Header as DialogHeader,
  Overlay as DialogOverlay,
  Portal as DialogPortal,
  Title as DialogTitle,
  Trigger as DialogTrigger,
  Footer,
  Header,
  Overlay,
  Portal,
  Root,
  Title,
  Trigger,
};
