import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import en from "./locales/en.json";
import ar from "./locales/ar.json";
import fr from "./locales/fr.json";

// i18next with bundled fallbacks. Remote bundles from GET /i18n/:locale can be
// merged in at runtime so adding a language server-side works without a deploy.
export const SUPPORTED = ["en", "ar", "fr"] as const;
export const RTL_LOCALES = new Set(["ar"]);

export function initialLocale(): string {
  const saved = localStorage.getItem("lang");
  if (saved && SUPPORTED.includes(saved as (typeof SUPPORTED)[number])) return saved;
  return "en";
}

void i18n.use(initReactI18next).init({
  resources: {
    en: { translation: en },
    ar: { translation: ar },
    fr: { translation: fr },
  },
  lng: initialLocale(),
  fallbackLng: "en",
  interpolation: { escapeValue: false },
});

// applyDir sets <html dir/lang> so RTL layout + Arabic font kick in.
export function applyDir(locale: string) {
  const dir = RTL_LOCALES.has(locale) ? "rtl" : "ltr";
  document.documentElement.setAttribute("dir", dir);
  document.documentElement.setAttribute("lang", locale);
}

export async function changeLanguage(locale: string) {
  localStorage.setItem("lang", locale);
  await i18n.changeLanguage(locale);
  applyDir(locale);
}

export default i18n;
