interface Link {
  text: string;
  href: string;
}

export interface NotificationMessage {
  type?: "default" | "success" | "error";
  message: string;
  detail?: string;
  link?: Link;
  options?: NotificationOptions;
}

interface NotificationOptions {
  persisted?: boolean;
}

export interface BannerMessage {
  type: "default" | "success" | "info" | "warning" | "error";
  message: string;
  includesHtml?: boolean;

  iconType: "none" | "alert" | "check" | "sleep" | "loading";

  // cta abstraction
  ctaText?: string;
  // if it is a direct link
  ctaUrl?: string;
  ctaTarget?: string;
  // callback when we need to take action like open pylon
  ctaCallback?: () => void;
}
