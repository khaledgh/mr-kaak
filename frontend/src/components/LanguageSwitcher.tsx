import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { changeLanguage } from "@/i18n";
import { api } from "@/api/endpoints";
import type { Language } from "@/api/types";

export function LanguageSwitcher() {
  const { i18n } = useTranslation();
  const [langs, setLangs] = useState<Language[]>([]);

  useEffect(() => {
    api.languages().then(setLangs).catch(() => setLangs([]));
  }, []);

  const options: Language[] = langs.length
    ? langs
    : [
        { code: "en", native_name: "English",  name: "English", is_default: true,  is_rtl: false },
        { code: "ar", native_name: "العربية",  name: "Arabic",  is_default: false, is_rtl: true  },
        { code: "fr", native_name: "Français", name: "French",  is_default: false, is_rtl: false },
      ];

  const currentIdx = options.findIndex((l) => l.code === i18n.language);
  const next = options[(currentIdx + 1) % options.length];

  return (
    <button
      aria-label={`Switch to ${next.native_name}`}
      title={`Switch to ${next.native_name}`}
      className="flex h-8 w-10 items-center justify-center rounded-full border border-brand-200 bg-white text-xs font-bold tracking-wide text-brand-700 transition hover:bg-brand-50"
      onClick={() => void changeLanguage(next.code)}
    >
      {i18n.language.toUpperCase().slice(0, 2)}
    </button>
  );
}
