export interface Envelope<T> {
  data?: T;
  meta?: PageMeta;
  error?: { code: string; message: string; details?: unknown };
}

export interface PageMeta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}

export interface User {
  id: number;
  name: string;
  email: string;
  phone?: string;
  role: string;
  status: string;
  created_at: string;
}

export interface AuthResponse {
  user: User;
  tokens: TokenPair;
}

// ── Media ──────────────────────────────────────────────
export interface MediaItem {
  id: number;
  url: string;
  thumb_url: string;
  original_name: string;
  mime: string;
  size_bytes: number;
  width: number;
  height: number;
  alt?: string;
  created_at: string;
}

// ── Catalog ────────────────────────────────────────────
export interface TranslationMap {
  [locale: string]: { name: string; description?: string };
}

export interface Category {
  id: number;
  slug: string;
  image_url?: string;
  sort_order: number;
  is_active: boolean;
  translations: TranslationMap;
  created_at: string;
}

export interface ProductVariant {
  id?: number;
  label: string;
  sku?: string;
  price_cents: number;
  is_default: boolean;
  sort_order: number;
}

export interface Modifier {
  id?: number;
  label: string;
  price_delta_cents: number;
  is_default: boolean;
  is_available: boolean;
}

export interface ModifierGroup {
  id?: number;
  label: string;
  min_select: number;
  max_select: number;
  is_required: boolean;
  sort_order: number;
  modifiers: Modifier[];
}

export interface Product {
  id: number;
  category_id: number;
  slug: string;
  image_url?: string;
  pricing_mode: "unit" | "weight";
  base_price_cents: number;
  is_preorder: boolean;
  preorder_lead_hours?: number;
  is_available: boolean;
  allergens?: string[];
  sort_order: number;
  translations: TranslationMap;
  variants: ProductVariant[];
  modifier_groups: ModifierGroup[];
  created_at: string;
}

// ── Coupons ────────────────────────────────────────────
export interface Coupon {
  id: number;
  code: string;
  type: "percent" | "fixed" | "free_delivery";
  value: number;
  min_order_cents: number;
  max_discount_cents?: number;
  usage_limit?: number;
  per_user_limit?: number;
  used_count: number;
  starts_at?: string;
  ends_at?: string;
  is_active: boolean;
  created_at: string;
}

// ── Banners ────────────────────────────────────────────
export interface Banner {
  id: number;
  title?: string;
  image_url: string;
  link_url?: string;
  sort_order: number;
  starts_at?: string;
  ends_at?: string;
  is_active: boolean;
  created_at: string;
}

// ── Delivery Zones ─────────────────────────────────────
export interface DeliveryZone {
  id: number;
  name: string;
  scope: "global" | "product";
  shape: "radius" | "polygon";
  center_lat?: number;
  center_lng?: number;
  radius_km?: number;
  polygon?: unknown;
  fee_cents: number;
  min_order_cents: number;
  is_active: boolean;
}

// ── Languages ──────────────────────────────────────────
export interface Language {
  id: number;
  code: string;
  name: string;
  native_name: string;
  is_rtl: boolean;
  is_active: boolean;
  is_default: boolean;
  sort_order: number;
}

// ── Settings ───────────────────────────────────────────
export interface Settings {
  cod_enabled: boolean;
  square_enabled: boolean;
  square_environment: string;
  square_application_id?: string;
  square_access_token?: string;
  square_location_id?: string;
  currency: string;
  tax_percent: number;
  meta_pixel_id?: string;
  meta_capi_token?: string;
  meta_test_event_code?: string;
}

// ── Orders (existing) ──────────────────────────────────
export interface AddressSnapshot {
  label?: string; line1: string; line2?: string;
  city: string; province_code: string; postal_code: string;
  country_code?: string; phone?: string; notes?: string;
}
export interface ModifierSnapshot { modifier_id: number; label: string; price_delta_cents: number; }
export interface OrderItem {
  id: number; product_id: number; variant_id?: number;
  name_snapshot: string; qty: number; weight_grams?: number;
  unit_price_cents: number; modifiers: ModifierSnapshot[];
  line_total_cents: number; created_at: string;
}
export interface OrderStatusHistory {
  id: number; order_id: number;
  from_status: string; to_status: string; note?: string;
  created_by_id?: number; created_at: string;
}
export type OrderStatus =
  | "pending_payment" | "paid" | "confirmed" | "preparing"
  | "ready" | "out_for_delivery" | "delivered" | "cancelled" | "refunded";
export interface Order {
  id: number; user_id: number; code: string; status: OrderStatus;
  fulfillment_type: "delivery" | "pickup";
  subtotal_cents: number; discount_cents: number;
  delivery_fee_cents: number; tax_cents: number;
  total_cents: number; currency: string;
  payment_method: string; payment_status: string;
  coupon_id?: number; address_snapshot?: AddressSnapshot;
  notes?: string; created_at: string;
  items?: OrderItem[]; history?: OrderStatusHistory[];
}
export const STATUS_TRANSITIONS: Record<OrderStatus, OrderStatus[]> = {
  pending_payment: ["paid", "confirmed", "cancelled"],
  paid: ["confirmed", "cancelled", "refunded"],
  confirmed: ["preparing", "cancelled", "refunded"],
  preparing: ["ready", "cancelled"],
  ready: ["out_for_delivery", "delivered", "cancelled"],
  out_for_delivery: ["delivered", "cancelled"],
  delivered: ["refunded"],
  cancelled: [], refunded: [],
};
export const STATUS_LABELS: Record<OrderStatus, string> = {
  pending_payment: "Pending Payment", paid: "Paid", confirmed: "Confirmed",
  preparing: "Preparing", ready: "Ready", out_for_delivery: "Out for Delivery",
  delivered: "Delivered", cancelled: "Cancelled", refunded: "Refunded",
};
export const STATUS_COLORS: Record<OrderStatus, string> = {
  pending_payment: "bg-gray-100 text-gray-700", paid: "bg-blue-100 text-blue-700",
  confirmed: "bg-indigo-100 text-indigo-700", preparing: "bg-amber-100 text-amber-700",
  ready: "bg-orange-100 text-orange-700", out_for_delivery: "bg-purple-100 text-purple-700",
  delivered: "bg-green-100 text-green-700", cancelled: "bg-red-100 text-red-700",
  refunded: "bg-gray-100 text-gray-500",
};
