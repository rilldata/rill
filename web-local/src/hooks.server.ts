import type { Handle } from "@sveltejs/kit";
import { paraglideMiddleware } from "@rilldata/web-common/features/i18n/gen/server";

// creating a handle to use the paraglide middleware
const paraglideHandle: Handle = ({ event, resolve }) =>
  paraglideMiddleware(
    event.request,
    ({ request: localizedRequest, locale }) => {
      event.request = localizedRequest;
      return resolve(event, {
        transformPageChunk: ({ html }) => {
          return html.replace("%lang%", locale);
        },
      });
    },
  );

export const handle: Handle = paraglideHandle;
