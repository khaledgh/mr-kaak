import { createBrowserRouter, Navigate } from "react-router-dom";
import { Layout } from "@/components/Layout";
import { LoginPage } from "@/features/auth/LoginPage";
import { OrdersPage } from "@/features/orders/OrdersPage";
import { OrderDetailPage } from "@/features/orders/OrderDetailPage";
import { MediaPage } from "@/features/media/MediaPage";
import { CatalogPage } from "@/features/catalog/CatalogPage";
import { CategoriesPage } from "@/features/catalog/CategoriesPage";
import { ProductsPage } from "@/features/catalog/ProductsPage";
import { CouponsPage } from "@/features/coupons/CouponsPage";
import { BannersPage } from "@/features/banners/BannersPage";
import { DeliveryPage } from "@/features/delivery/DeliveryPage";
import { DeliveryFormPage } from "@/features/delivery/DeliveryFormPage";
import { LanguagesPage } from "@/features/languages/LanguagesPage";
import { UsersPage } from "@/features/users/UsersPage";
import { SettingsPage } from "@/features/settings/SettingsPage";
import { useAuth } from "@/stores/auth";

const STAFF_ROLES = ["kitchen", "staff", "admin", "super_admin"];

function RequireStaff({ children }: { children: React.ReactNode }) {
  const { accessToken, user } = useAuth();
  if (!accessToken) return <Navigate to="/login" replace />;
  if (!user || !STAFF_ROLES.includes(user.role)) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="card max-w-sm space-y-3 p-8 text-center">
          <p className="text-lg font-semibold text-red-600">Access Denied</p>
          <p className="text-sm text-gray-500">Staff accounts only.</p>
          <button className="btn-ghost text-sm" onClick={() => useAuth.getState().logout()}>Sign out</button>
        </div>
      </div>
    );
  }
  return <>{children}</>;
}

export const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  {
    path: "/",
    element: <RequireStaff><Layout /></RequireStaff>,
    children: [
      { index: true,                      element: <Navigate to="/orders" replace /> },
      { path: "orders",                   element: <OrdersPage /> },
      { path: "orders/:code",             element: <OrderDetailPage /> },
      { path: "media",                    element: <MediaPage /> },
      { path: "catalog",                  element: <CatalogPage /> },
      { path: "catalog/categories",       element: <CategoriesPage /> },
      { path: "catalog/products",         element: <ProductsPage /> },
      { path: "coupons",                  element: <CouponsPage /> },
      { path: "banners",                  element: <BannersPage /> },
      { path: "delivery",                 element: <DeliveryPage /> },
      { path: "delivery/new",             element: <DeliveryFormPage /> },
      { path: "delivery/:id/edit",        element: <DeliveryFormPage /> },
      { path: "languages",                element: <LanguagesPage /> },
      { path: "users",                    element: <UsersPage /> },
      { path: "settings",                 element: <SettingsPage /> },
      { path: "*",                        element: <Navigate to="/orders" replace /> },
    ],
  },
]);
