export async function load({ parent }) {
  const { onboardingState } = await parent();
  await onboardingState.fetch().catch(console.error);

  return { onboardingState };
}
