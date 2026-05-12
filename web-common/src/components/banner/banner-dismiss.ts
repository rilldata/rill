import type { BannerMessage } from "@rilldata/web-common/lib/event-bus/events.ts";

type BannerDismissState = {
  id: string;
  dismissedTill?: number;
};

export function isBannerDismissed(dismiss: BannerMessage["dismissible"]) {
  if (!dismiss) return false;

  const key = `rill:banner:dismiss:${dismiss.key}`;
  try {
    const rawValue = localStorage.getItem(key);
    if (!rawValue) return false;
    const value = JSON.parse(rawValue) as BannerDismissState;
    if (value.id !== dismiss?.id) return false;
    return !value.dismissedTill || value.dismissedTill > Date.now();
  } catch {
    return false;
  }
}

export function dismissBanner(dismiss: BannerMessage["dismissible"]) {
  if (!dismiss) return;
  const key = `rill:banner:dismiss:${dismiss.key}`;
  const value: BannerDismissState = {
    id: dismiss.id,
  };
  if (dismiss.ttl) {
    value.dismissedTill = Date.now() + dismiss.ttl;
  }
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch {
    // no-op
  }
}
