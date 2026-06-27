// Money formatting. The API stores integer cents; format for display only.
export function formatMoney(cents: number, currency = "CAD", locale = "en"): string {
  return new Intl.NumberFormat(locale === "ar" ? "ar" : locale, {
    style: "currency",
    currency,
  }).format(cents / 100);
}
