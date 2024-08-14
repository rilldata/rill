const appId = import.meta.env.RILL_UI_PUBLIC_PYLON_APP_ID as string;

export type UserLike = {
  email: string;
  displayName: string;
  photoUrl: string;
};

/**
 * Function implementation is copied from: https://docs.usepylon.com/chat/setup
 */
export function initPylonChat(user: UserLike) {
  window.pylon = {
    chat_settings: {
      app_id: appId,
      email: user.email,
      name: user.displayName,
      avatar_url: user.photoUrl,
    },
  };
}
