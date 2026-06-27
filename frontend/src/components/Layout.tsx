import { useState } from "react";
import { Link, NavLink, Outlet, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useCart } from "@/stores/cart";
import { useAuth } from "@/stores/auth";
import { LanguageSwitcher } from "./LanguageSwitcher";
import { CartDrawer } from "@/features/cart/CartDrawer";

export function Layout() {
  const { t } = useTranslation();
  const count = useCart((s) => s.count());
  const { user, logout } = useAuth();
  const [cartOpen, setCartOpen] = useState(false);
  const navigate = useNavigate();

  function handleLogout() {
    logout();
    navigate("/");
  }

  const desktopLink = ({ isActive }: { isActive: boolean }) =>
    isActive ? "btn-ghost bg-brand-100" : "btn-ghost";

  return (
    // pb-14 on mobile to clear the fixed bottom nav; removed on md+
    <div className="min-h-screen pb-14 md:pb-0">
      <header className="sticky top-0 z-20 border-b border-brand-100 bg-brand-50/90 backdrop-blur">
        <div className="mx-auto flex max-w-3xl items-center gap-3 px-4 py-3">
          <Link to="/" className="flex items-center gap-2 text-lg font-bold text-brand-700 shrink-0 hover:opacity-90 transition-opacity">
            <img src="/logo.png" className="h-14 w-auto object-contain" alt={t("app.title")} />
            {/* <span className="hidden sm:inline">{t("app.title")}</span> */}
          </Link>

          {/* Desktop nav — hidden on small screens */}
          <nav className="ms-auto hidden items-center gap-2 text-sm md:flex">
            <NavLink to="/" end className={desktopLink}>{t("nav.menu")}</NavLink>
            {user && <NavLink to="/orders" className={desktopLink}>{t("nav.orders")}</NavLink>}
            {user && <NavLink to="/addresses" className={desktopLink}>{t("nav.addresses")}</NavLink>}
            {user ? (
              <button className="btn-ghost" onClick={handleLogout}>{t("nav.logout")}</button>
            ) : (
              <NavLink to="/login" className={desktopLink}>{t("nav.login")}</NavLink>
            )}
            <LanguageSwitcher />
            <button
              className="relative flex h-10 w-10 items-center justify-center rounded-full bg-brand-600 text-white shadow hover:bg-brand-700 transition"
              onClick={() => setCartOpen(true)}
              aria-label={t("nav.cart")}
            >
              <BagIcon />
              {count > 0 && (
                <span className="absolute -right-1 -top-1 flex h-5 w-5 items-center justify-center rounded-full bg-white text-xs font-bold text-brand-700 shadow">
                  {count}
                </span>
              )}
            </button>
          </nav>

          {/* Mobile: language pill + cart bag icon */}
          <div className="ms-auto flex items-center gap-2 md:hidden">
            <LanguageSwitcher />
            <button
              className="relative flex h-10 w-10 items-center justify-center rounded-full bg-brand-600 text-white shadow hover:bg-brand-700 transition"
              onClick={() => setCartOpen(true)}
              aria-label={t("nav.cart")}
            >
              <BagIcon />
              {count > 0 && (
                <span className="absolute -right-1 -top-1 flex h-5 w-5 items-center justify-center rounded-full bg-white text-xs font-bold text-brand-700 shadow">
                  {count}
                </span>
              )}
            </button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-3xl px-4 py-6">
        <Outlet />
      </main>

      {/* Mobile bottom nav — hidden on desktop */}
      <nav className="fixed inset-x-0 bottom-0 z-20 border-t border-brand-100 bg-white md:hidden">
        <div className="mx-auto flex max-w-3xl items-stretch justify-around">
          <BottomTab to="/" end label={t("nav.menu")} icon={<HomeIcon />} />
          <BottomTab to="/orders" label={t("nav.orders")} icon={<OrdersIcon />} />
          <BottomTab to="/addresses" label={t("nav.addresses")} icon={<MapPinIcon />} />
          {user ? (
            <button
              className="flex flex-1 flex-col items-center justify-center gap-0.5 py-2 text-xs text-brand-400"
              onClick={handleLogout}
            >
              <LogoutIcon />
              <span>{t("nav.logout")}</span>
            </button>
          ) : (
            <BottomTab to="/login" label={t("nav.login")} icon={<PersonIcon />} />
          )}
        </div>
      </nav>

      <CartDrawer open={cartOpen} onClose={() => setCartOpen(false)} />
    </div>
  );
}

function BottomTab({
  to, label, icon, end,
}: {
  to: string;
  label: string;
  icon: React.ReactNode;
  end?: boolean;
}) {
  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        `flex flex-1 flex-col items-center justify-center gap-0.5 py-2 text-xs transition-colors ${isActive ? "text-brand-700" : "text-brand-400"
        }`
      }
    >
      {icon}
      <span>{label}</span>
    </NavLink>
  );
}

function HomeIcon() {
  return (
    <svg className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth={1.8} viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" d="M3 9.5L12 3l9 6.5V20a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V9.5z" />
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 21V12h6v9" />
    </svg>
  );
}

function OrdersIcon() {
  return (
    <svg className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth={1.8} viewBox="0 0 24 24">
      <rect x="5" y="3" width="14" height="18" rx="2" strokeLinecap="round" strokeLinejoin="round" />
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 7h6M9 11h6M9 15h4" />
    </svg>
  );
}

function MapPinIcon() {
  return (
    <svg className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth={1.8} viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 21s-7-7.5-7-12.5a7 7 0 0 1 14 0c0 5-7 12.5-7 12.5z" />
      <circle cx="12" cy="8.5" r="2.5" />
    </svg>
  );
}

function PersonIcon() {
  return (
    <svg className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth={1.8} viewBox="0 0 24 24">
      <circle cx="12" cy="8" r="4" strokeLinecap="round" strokeLinejoin="round" />
      <path strokeLinecap="round" strokeLinejoin="round" d="M4 20c0-4 3.6-7 8-7s8 3 8 7" />
    </svg>
  );
}

function LogoutIcon() {
  return (
    <svg className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth={1.8} viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2h6a2 2 0 0 1 2 2v1" />
    </svg>
  );
}

function BagIcon() {
  return (
    <svg className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth={1.8} viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4z" />
      <line x1="3" y1="6" x2="21" y2="6" strokeLinecap="round" />
      <path strokeLinecap="round" strokeLinejoin="round" d="M16 10a4 4 0 0 1-8 0" />
    </svg>
  );
}
