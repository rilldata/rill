export async function load({ parent }) {
  const { onboardingState } = await parent();

  // Create the onboarding-state.json file, or fetch it
  // TODO: probably push this branching logic into the OnboardingState class
  if (!(await onboardingState.isOnboardingStateFilePresent())) {
    await onboardingState.initializeOnboardingState().catch(console.error);
  } else {
    await onboardingState.fetchAndParse().catch(console.error);
  }

  return { onboardingState };
}
