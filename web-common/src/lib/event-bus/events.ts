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

export interface PageContentResized {
  width: number;
  height: number;
}
