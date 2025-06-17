/**
 * Constructs the appropriate signup URL based on the environment
 *
 * AUTHENTICATION STRATEGY:
 *
 * LOCAL DEVELOPMENT (isDev = true):
 * - Uses runtime server's auth endpoint via loginUrl (localhost:9009/auth)
 * - Runtime server has simple PKCE flow that redirects to admin server
 * - Auth0 is NOT configured with local runtime, so /auth/signup is not available
 * - Generic /auth endpoint provides basic authentication flow
 *
 * CLOUD ENVIRONMENTS (isDev = false):
 * - Uses admin server's dedicated signup endpoint via adminUrl (admin.rilldata.com/auth/signup)
 * - Admin server has full Auth0 integration with signup flow
 * - Provides complete signup experience with proper UI and flow
 * - Handles cookies, tokens, and redirects properly
 *
 * @param metadata - The metadata object containing isDev, loginUrl, and adminUrl
 * @param redirectUrl - The URL to redirect to after successful authentication
 * @returns The complete signup URL with redirect parameter
 */
export function constructSignupUrl(
  metadata: { isDev: boolean; loginUrl: string; adminUrl: string },
  redirectUrl: string,
): string {
  const signupUrl = metadata.isDev
    ? metadata.loginUrl
    : `${metadata.adminUrl}/auth/signup`;

  return `${signupUrl}?redirect=${redirectUrl}`;
}
