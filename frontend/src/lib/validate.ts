// Field validators — return an i18n key on failure, null on success.

export function vRequired(v: string): string | null {
  return v.trim() ? null : "validation.required";
}

export function vEmail(v: string): string | null {
  if (!v.trim()) return "validation.required";
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v) ? null : "validation.emailInvalid";
}

export function vPassword(v: string): string | null {
  if (!v) return "validation.required";
  return v.length >= 8 ? null : "validation.passwordMin";
}

export function vName(v: string): string | null {
  if (!v.trim()) return "validation.required";
  return v.trim().length >= 2 ? null : "validation.nameMin";
}

// Phone is optional — only validates format when a value is present.
export function vPhone(v: string): string | null {
  if (!v.trim()) return null;
  return /^[+\d][\d\s\-().]{5,}$/.test(v.trim()) ? null : "validation.phoneInvalid";
}

// Canadian postal code: A1A 1A1 or A1A1A1
export function vPostalCA(v: string): string | null {
  if (!v.trim()) return "validation.required";
  return /^[A-Za-z]\d[A-Za-z][\s-]?\d[A-Za-z]\d$/.test(v.trim())
    ? null
    : "validation.postalCodeCA";
}
