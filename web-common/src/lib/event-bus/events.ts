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
