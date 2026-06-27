import { useState } from "react";
import { Link, NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "@/stores/auth";
import { ToastContainer } from "./Toast";

const ADMIN_ROLES = ["admin", "super_admin"];

interface NavItem {
  to: string;
  label: string;
  icon: string;
  adminOnly?: boolean;
}

const NAV: NavItem[] = [
  { to: "/orders",     label: "Orders",      icon: "🧾" },
  { to: "/media",      label: "Media",        icon: "🖼️", adminOnly: true },
  { to: "/catalog",    label: "Catalog",      icon: "🍬", adminOnly: true },
  { to: "/coupons",    label: "Coupons",      icon: "🏷️", adminOnly: true },
  { to: "/banners",    label: "Banners",      icon: "📢", adminOnly: true },
  { to: "/delivery",   label: "Delivery",     icon: "🚚", adminOnly: true },
  { to: "/languages",  label: "Localization", icon: "🌐", adminOnly: true },
  { to: "/users",      label: "Users",        icon: "👥", adminOnly: true },
  { to: "/settings",   label: "Settings",     icon: "⚙️", adminOnly: true },
];

export function Layout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const isAdmin = ADMIN_ROLES.includes(user?.role ?? "");

  const visibleNav = NAV.filter((n) => !n.adminOnly || isAdmin);

  const sidebar = (
    <nav className="flex h-full flex-col">
      <div className="flex h-14 items-center px-5 border-b border-gray-100">
        <Link to="/orders" className="text-base font-bold text-brand-700">
          mrkaak<span className="font-normal text-gray-400"> admin</span>
        </Link>
      </div>
      <div className="flex-1 overflow-y-auto py-3 px-2 space-y-0.5">
        {visibleNav.map((n) => (
          <NavLink
            key={n.to}
            to={n.to}
            onClick={() => setOpen(false)}
            className={({ isActive }) =>
              `flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition ${
                isActive
                  ? "bg-brand-50 text-brand-700"
                  : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
              }`
            }
          >
            <span className="text-base">{n.icon}</span>
            {n.label}
          </NavLink>
        ))}
      </div>
      <div className="border-t border-gray-100 px-3 py-3">
        <div className="mb-2 px-2 text-xs text-gray-400">
          {user?.name}{" "}
          <span className="rounded bg-brand-100 px-1 py-0.5 font-medium text-brand-600">
            {user?.role}
          </span>
        </div>
        <button
          className="btn-ghost w-full justify-start text-xs text-gray-500"
          onClick={() => { logout(); navigate("/login"); }}
        >
          Sign out
        </button>
      </div>
    </nav>
  );

  return (
    <div className="flex min-h-screen">
      {/* Desktop sidebar */}
      <aside className="hidden w-56 shrink-0 border-r border-gray-200 bg-white lg:block">
        {sidebar}
      </aside>

      {/* Mobile overlay sidebar */}
      {open && (
        <>
          <div
            className="fixed inset-0 z-40 bg-black/30 lg:hidden"
            onClick={() => setOpen(false)}
          />
          <aside className="animate-slide-in fixed inset-y-0 left-0 z-50 w-56 border-r border-gray-200 bg-white lg:hidden">
            {sidebar}
          </aside>
        </>
      )}

      {/* Main content */}
      <div className="flex flex-1 flex-col min-w-0">
        {/* Topbar (mobile) */}
        <header className="sticky top-0 z-30 flex h-14 items-center gap-3 border-b border-gray-200 bg-white px-4 lg:hidden">
          <button
            className="btn-ghost p-1.5"
            onClick={() => setOpen(true)}
            aria-label="Open menu"
          >
            ☰
          </button>
          <span className="font-bold text-brand-700">mrkaak admin</span>
        </header>

        <main className="flex-1 p-5 sm:p-7">
          <Outlet />
        </main>
      </div>

      <ToastContainer />
    </div>
  );
}
