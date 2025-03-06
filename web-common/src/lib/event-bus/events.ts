interface Link {
  text: string;
  href: string;
}

export interface NotificationMessage {
  type?: "default" | "success" | "loading" | "error";
  message?: string;
  detail?: string;
  link?: Link;
  options?: NotificationOptions;
}

interface NotificationOptions {
  persisted?: boolean;
  timeout?: number;
}

export enum BannerSlot {
  Billing,
  Dashboard,
  // Catch all slot when none is provided.
  // Make sure this is at the end
  Other,
}
export interface BannerEvent {
  // Slot for the banner to appear. A new banner here will clear the older ones.
  // A lower number here will mean it will show up higher.
  slot: BannerSlot;
  // null here means the slot is being cleared.
  message: BannerMessage | null;
}
export interface BannerMessage {
  type: "default" | "success" | "info" | "warning" | "error";
  message: string;
  includesHtml?: boolean;

  iconType: "none" | "alert" | "check" | "sleep" | "loading";

  // cta abstraction
  cta?: {
    text: string;
    type: "button" | "link";

    // if it is a direct link
    url?: string;
    target?: string;

    // callback when we need to take action like open pylon
    onClick?: () => void | Promise<void>;
  };
}
