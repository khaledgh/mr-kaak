// Shared TS types mirroring the API DTOs (response envelope, catalog, orders).

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

export type PricingMode = "unit" | "weight";

export interface Variant {
  id: number;
  label: string;
  price_cents: number;
  is_default: boolean;
}

export interface Modifier {
  id: number;
  label: string;
  price_delta_cents: number;
  is_default: boolean;
}

export interface ModifierGroup {
  id: number;
  label: string;
  min_select: number;
  max_select: number;
  is_required: boolean;
  modifiers: Modifier[];
}

export interface Product {
  id: number;
  slug: string;
  pricing_mode: PricingMode;
  base_price_cents: number;
  is_preorder: boolean;
  is_available: boolean;
  image_url?: string;
  allergens: string[];
  name: string;
  description?: string;
  variants?: Variant[];
  modifier_groups?: ModifierGroup[];
}

export interface Category {
  id: number;
  slug: string;
  image_url?: string;
  name: string;
  description?: string;
  products?: Product[];
}

export interface Language {
  code: string;
  name: string;
  native_name: string;
  is_default: boolean;
  is_rtl: boolean;
}

export interface Banner {
  id: number;
  title: string;
  image_url: string;
  link_url?: string;
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
}

export interface AuthResponse {
  user: User;
  tokens: TokenPair;
}

export interface Address {
  id: number;
  label?: string;
  line1: string;
  line2?: string;
  city: string;
  province_code: string;
  postal_code: string;
  lat?: number;
  lng?: number;
  phone?: string;
  is_default?: boolean;
}

export interface OrderItem {
  id: number;
  product_id: number;
  name_snapshot: string;
  qty: number;
  weight_grams?: number;
  unit_price_cents: number;
  line_total_cents: number;
}

export interface Order {
  id: number;
  code: string;
  status: string;
  fulfillment_type: "delivery" | "pickup";
  subtotal_cents: number;
  discount_cents: number;
  delivery_fee_cents: number;
  tax_cents: number;
  total_cents: number;
  currency: string;
  payment_method: string;
  payment_status: string;
  items: OrderItem[];
}

export interface DeliveryQuote {
  deliverable: boolean;
  fee_cents: number;
  min_order_cents: number;
  reason?: string;
}
