// Shared display name rule for the styleguide demo and the auth first-time setup flow.
export const DISPLAY_NAME_MAX_LENGTH = 30;

export function validateDisplayName(name: string): string | undefined {
  const trimmed = name.trim();
  return trimmed.length === 0
    ? "Display name can't be empty."
    : trimmed.length > DISPLAY_NAME_MAX_LENGTH
      ? `Display name is too long (max ${DISPLAY_NAME_MAX_LENGTH} characters).`
      : undefined;
}
