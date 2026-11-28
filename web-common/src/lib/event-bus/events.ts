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

export interface BannerEvent {
  // Unique identifier used to set/unset for particular feature.
  id: string;
  // Determines the order in which the banner is shown.
  // Lower values means the banner is shown higher in the stack.
  priority: number;
  message: BannerMessage;
}
export interface BannerMessage {
  type: "default" | "success" | "info" | "warning" | "error";
  // Shown as plain text, unless one of the flags below is set.
  message: string;
  // The message is HTML built by the app.
  includesHtml?: boolean;
  // The message is inline markdown written by a project developer, such as a dashboard's `banner`.
  includesMarkdown?: boolean;

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

  dismissible?: {
    // TTL scope for the banner's dismissal tracking.
    key: string;
    // ID used to calculate ttl for.
    // The same key can switch IDs. This means we need to start a new counter but save under the same `key`
    id: string;
    // How long the banner stays dismissed, in milliseconds.
    // If value is 0 then banner is dismissed forever.
    ttl: number;
  };
}

export interface PageContentResized {
  width: number;
  height: number;
}
