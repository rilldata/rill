import { getOnboardingState } from "@rilldata/web-common/features/welcome/wizard/onboarding-state";

export async function load() {
  const onboardingState = getOnboardingState();
  await onboardingState.fetch().catch(console.error);

  return { onboardingState };
}
