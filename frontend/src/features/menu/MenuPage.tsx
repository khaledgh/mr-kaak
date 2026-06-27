import { useEffect, useMemo, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { api, qk } from "@/api/endpoints";
import { ProductCard } from "./ProductCard";

export function MenuPage() {
  const { t, i18n } = useTranslation();
  const lang = i18n.language;
  const [q, setQ] = useState("");

  const menu = useQuery({ queryKey: qk.menu(lang), queryFn: () => api.menu(lang) });
  const banners = useQuery({ queryKey: qk.banners(), queryFn: api.banners });

  const displayBanners = useMemo(() => {
    const list = banners.data ?? [];
    if (list.length === 0) {
      return [
        {
          id: 0,
          title: "Delicious Sweets Are Waiting For You",
          image_url: "",
          link_url: "",
        },
      ];
    }
    return list;
  }, [banners.data]);

  const navigate = useNavigate();
  const scrollRef = useRef<HTMLDivElement>(null);
  const menuSectionRef = useRef<HTMLDivElement>(null);
  const [activeSlide, setActiveSlide] = useState(0);
  const [isHovering, setIsHovering] = useState(false);

  const handleScrollToMenu = () => {
    menuSectionRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  const handleOrderNow = (linkUrl?: string) => {
    if (linkUrl && linkUrl !== "#" && linkUrl !== "") {
      if (linkUrl.startsWith("http")) {
        window.location.href = linkUrl;
      } else {
        navigate(linkUrl);
      }
    } else {
      handleScrollToMenu();
    }
  };

  const handleScroll = () => {
    if (!scrollRef.current) return;
    const { scrollLeft, clientWidth } = scrollRef.current;
    const isRtl = document.documentElement.dir === "rtl";
    const scrollPos = isRtl ? Math.abs(scrollLeft) : scrollLeft;
    
    const index = Math.round(scrollPos / clientWidth);
    if (index >= 0 && index < displayBanners.length) {
      setActiveSlide(index);
    }
  };

  const scrollToSlide = (index: number) => {
    if (!scrollRef.current) return;
    const isRtl = document.documentElement.dir === "rtl";
    const { clientWidth } = scrollRef.current;
    const targetScrollLeft = index * clientWidth;
    
    scrollRef.current.scrollTo({
      left: isRtl ? -targetScrollLeft : targetScrollLeft,
      behavior: "smooth",
    });
    setActiveSlide(index);
  };

  useEffect(() => {
    if (displayBanners.length <= 1 || isHovering) return;
    const interval = setInterval(() => {
      setActiveSlide((prev) => {
        const next = (prev + 1) % displayBanners.length;
        scrollToSlide(next);
        return next;
      });
    }, 5000);
    return () => clearInterval(interval);
  }, [displayBanners.length, isHovering]);

  const handleMouseEnter = () => setIsHovering(true);
  const handleMouseLeave = () => setIsHovering(false);
  const handleTouchStart = () => setIsHovering(true);
  const handleTouchEnd = () => {
    setTimeout(() => {
      setIsHovering(false);
    }, 2000);
  };

  const filtered = useMemo(() => {
    const cats = menu.data ?? [];
    if (!q.trim()) return cats;
    const needle = q.toLowerCase();
    return cats
      .map((c) => ({
        ...c,
        products: (c.products ?? []).filter((p) => p.name.toLowerCase().includes(needle)),
      }))
      .filter((c) => (c.products ?? []).length > 0);
  }, [menu.data, q]);

  return (
    <div className="space-y-8">
      {/* ── Hero Carousel ── */}
      <div 
        className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-amber-50 via-brand-100 to-brand-200 group/hero min-h-[180px]"
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        onTouchStart={handleTouchStart}
        onTouchEnd={handleTouchEnd}
      >
        {/* decorative blobs */}
        <div className="pointer-events-none absolute -right-10 -top-10 h-52 w-52 rounded-full bg-brand-300/25 blur-3xl" />
        <div className="pointer-events-none absolute -bottom-8 -left-8 h-40 w-40 rounded-full bg-amber-200/40 blur-2xl" />

        {/* Scroll Container */}
        <div
          ref={scrollRef}
          onScroll={handleScroll}
          className="flex overflow-x-auto snap-x snap-mandatory scrollbar-hide scroll-smooth"
        >
          {displayBanners.map((b, idx) => (
            <div
              key={b.id || idx}
              className="w-full shrink-0 snap-start flex items-center gap-4 px-6 py-8 z-10"
            >
              <div className="flex-1 min-w-0">
                <p className="mb-1 text-xs font-semibold uppercase tracking-widest text-brand-500">
                  {t("app.title")}
                </p>
                <h1 className="mb-4 text-xl sm:text-2xl font-bold leading-tight text-brand-900 line-clamp-2">
                  {b.title || "Delicious Sweets Are Waiting For You"}
                </h1>
                <div className="flex flex-wrap gap-2">
                  <button 
                    onClick={handleScrollToMenu}
                    className="rounded-full bg-brand-600 px-5 py-2 text-sm font-medium text-white shadow-md hover:bg-brand-700 transition"
                  >
                    {t("nav.menu")}
                  </button>
                  <button 
                    onClick={() => handleOrderNow(b.link_url)}
                    className="rounded-full border border-white/80 bg-white/70 px-5 py-2 text-sm font-medium text-brand-700 shadow backdrop-blur hover:bg-white transition"
                  >
                    Order Now
                  </button>
                </div>
              </div>

              {b.image_url ? (
                <img
                  src={b.image_url}
                  alt={b.title || "Banner Image"}
                  className="h-28 w-28 sm:h-32 sm:w-32 shrink-0 rounded-full object-cover shadow-2xl ring-4 ring-white/60"
                />
              ) : (
                <div className="shrink-0 text-6xl sm:text-7xl leading-none drop-shadow-lg">🍯</div>
              )}
            </div>
          ))}
        </div>

        {/* Navigation Arrows (Desktop only, visible on hover) */}
        {displayBanners.length > 1 && (
          <>
            <button
              onClick={() => {
                const prev = (activeSlide - 1 + displayBanners.length) % displayBanners.length;
                scrollToSlide(prev);
              }}
              className="absolute left-4 top-1/2 -translate-y-1/2 z-20 hidden group-hover/hero:flex items-center justify-center h-8 w-8 rounded-full bg-white/90 hover:bg-white shadow-md border border-brand-100 text-brand-700 hover:text-brand-900 transition active:scale-95"
              aria-label="Previous slide"
            >
              <svg className="h-4.5 w-4.5 rtl:rotate-180" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 19.5L8.25 12l7.5-7.5" />
              </svg>
            </button>
            <button
              onClick={() => {
                const next = (activeSlide + 1) % displayBanners.length;
                scrollToSlide(next);
              }}
              className="absolute right-4 top-1/2 -translate-y-1/2 z-20 hidden group-hover/hero:flex items-center justify-center h-8 w-8 rounded-full bg-white/90 hover:bg-white shadow-md border border-brand-100 text-brand-700 hover:text-brand-900 transition active:scale-95"
              aria-label="Next slide"
            >
              <svg className="h-4.5 w-4.5 rtl:rotate-180" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
              </svg>
            </button>
          </>
        )}

        {/* Dots Indicator */}
        {displayBanners.length > 1 && (
          <div className="absolute bottom-3 left-1/2 -translate-x-1/2 flex justify-center gap-1.5 z-20">
            {displayBanners.map((_, idx) => (
              <button
                key={idx}
                onClick={() => scrollToSlide(idx)}
                className={`h-1.5 rounded-full transition-all duration-300 ${
                  activeSlide === idx
                    ? "w-4 bg-brand-600"
                    : "w-1.5 bg-brand-400/50 hover:bg-brand-500"
                }`}
                aria-label={`Go to slide ${idx + 1}`}
              />
            ))}
          </div>
        )}
      </div>

      {/* ── Search ── */}
      <div ref={menuSectionRef} className="relative">
        <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-brand-400">
          <svg className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
            <circle cx="11" cy="11" r="8" />
            <path strokeLinecap="round" d="m21 21-4.35-4.35" />
          </svg>
        </span>
        <input
          className="input ps-9 rounded-2xl"
          placeholder={t("menu.search")}
          value={q}
          onChange={(e) => setQ(e.target.value)}
        />
      </div>

      {menu.isLoading && (
        <p className="py-8 text-center text-brand-700/60">{t("common.loading")}</p>
      )}
      {menu.isError && (
        <p className="text-center text-red-600">{t("common.error")}</p>
      )}

      {/* ── Category sections ── */}
      {filtered.map((cat) => (
        <section key={cat.id}>
          <div className="mb-4 text-center">
            <h2 className="text-2xl font-bold text-brand-900">{cat.name}</h2>
            <p className="mt-0.5 text-sm text-brand-600/70">
              {cat.description ?? "Our specialty dishes"}
            </p>
          </div>

          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
            {(cat.products ?? []).map((p) => (
              <ProductCard key={p.id} product={p} />
            ))}
          </div>
        </section>
      ))}

      {!menu.isLoading && filtered.length === 0 && (
        <p className="py-12 text-center text-brand-700/60">{t("menu.empty")}</p>
      )}
    </div>
  );
}
