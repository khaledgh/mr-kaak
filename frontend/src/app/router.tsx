import { createBrowserRouter, Navigate } from "react-router-dom";
import { Layout } from "@/components/Layout";
import { MenuPage } from "@/features/menu/MenuPage";
import { CheckoutPage } from "@/features/checkout/CheckoutPage";
import { OrdersPage } from "@/features/orders/OrdersPage";
import { OrderTrackPage } from "@/features/orders/OrderTrackPage";
import { LoginPage } from "@/features/auth/LoginPage";
import { RegisterPage } from "@/features/auth/RegisterPage";
import { AddressesPage } from "@/features/account/AddressesPage";
import { useAuth } from "@/stores/auth";

function RequireAuth({ children }: { children: React.ReactNode }) {
  const token = useAuth((s) => s.accessToken);
  return token ? <>{children}</> : <Navigate to="/login" replace />;
}

export const router = createBrowserRouter([
  {
    path: "/",
    element: <Layout />,
    children: [
      { index: true, element: <MenuPage /> },
      { path: "login", element: <LoginPage /> },
      { path: "register", element: <RegisterPage /> },
      { path: "checkout", element: <RequireAuth><CheckoutPage /></RequireAuth> },
      { path: "orders", element: <RequireAuth><OrdersPage /></RequireAuth> },
      { path: "orders/:code", element: <RequireAuth><OrderTrackPage /></RequireAuth> },
      { path: "addresses", element: <RequireAuth><AddressesPage /></RequireAuth> },
      { path: "*", element: <Navigate to="/" replace /> },
    ],
  },
]);
