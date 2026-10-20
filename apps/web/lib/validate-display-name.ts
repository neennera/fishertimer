// Shared display name rule for the styleguide demo and the auth first-time setup flow.
// Mirrors the account service (services/account/internal/usecase/service.go,
// domain.MaxDisplayNameLength): 30 characters as in wireframes 02c / 04e,
// counted as code points like the backend's runes, and no control or
// invisible-only names.
export const DISPLAY_NAME_MAX_LENGTH = 30;

const CONTROL_CHARACTER = /\p{Cc}/u;
const INVISIBLE_OR_SPACE = /^[\s​‌‍⁠﻿]*$/u;

export function validateDisplayName(name: string): string | undefined {
  const trimmed = name.trim();
  if (trimmed.length === 0 || INVISIBLE_OR_SPACE.test(trimmed)) {
    return "Display name can't be empty.";
  }
  if ([...trimmed].length > DISPLAY_NAME_MAX_LENGTH) {
    return `Display name is too long (max ${DISPLAY_NAME_MAX_LENGTH} characters).`;
  }
  if (CONTROL_CHARACTER.test(trimmed)) {
    return "Display name can't contain control characters.";
  }
  return undefined;
}
