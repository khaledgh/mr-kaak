import { http, unwrap } from "./client";
import type {
  Address,
  AuthResponse,
  Banner,
  Category,
  DeliveryQuote,
  Envelope,
  Language,
  Order,
  Product,
} from "./types";

// Typed API surface. Query keys live alongside so TanStack Query usage stays
// consistent across the app.
export const api = {
  // Catalog
  async menu(lang: string): Promise<Category[]> {
    const { data } = await http.get<Envelope<Category[]>>("/menu", { params: { lang } });
    return unwrap(data);
  },
  async product(slug: string, lang: string): Promise<Product> {
    const { data } = await http.get<Envelope<Product>>(`/products/${slug}`, { params: { lang } });
    return unwrap(data);
  },
  async search(q: string, lang: string): Promise<Product[]> {
    const { data } = await http.get<Envelope<Product[]>>("/search", { params: { q, lang } });
    return unwrap(data) ?? [];
  },
  async banners(): Promise<Banner[]> {
    const { data } = await http.get<Envelope<Banner[]>>("/banners");
    return unwrap(data) ?? [];
  },
  async languages(): Promise<Language[]> {
    const { data } = await http.get<Envelope<Language[]>>("/languages");
    return unwrap(data) ?? [];
  },

  // Auth
  async register(body: { name: string; email: string; password: string; phone?: string }) {
    const { data } = await http.post<Envelope<AuthResponse>>("/auth/register", body);
    return unwrap(data);
  },
  async login(body: { email: string; password: string }) {
    const { data } = await http.post<Envelope<AuthResponse>>("/auth/login", body);
    return unwrap(data);
  },

  // Addresses
  async addresses(): Promise<Address[]> {
    const { data } = await http.get<Envelope<Address[]>>("/me/addresses");
    return unwrap(data) ?? [];
  },
  async addAddress(body: Partial<Address>): Promise<Address> {
    const { data } = await http.post<Envelope<Address>>("/me/addresses", body);
    return unwrap(data);
  },
  async deleteAddress(id: number): Promise<void> {
    await http.delete(`/me/addresses/${id}`);
  },
  async setDefaultAddress(id: number): Promise<Address> {
    const { data } = await http.post<Envelope<Address>>(`/me/addresses/${id}/default`);
    return unwrap(data);
  },

  // Delivery + payment
  async deliveryQuote(body: { lat: number; lng: number; product_ids: number[] }): Promise<DeliveryQuote> {
    const { data } = await http.post<Envelope<DeliveryQuote>>("/delivery/quote", body);
    return unwrap(data);
  },
  async paymentMethods(): Promise<string[]> {
    const { data } = await http.get<Envelope<{ methods: string[] }>>("/payment/methods");
    return unwrap(data).methods ?? [];
  },

  // Orders
  async checkout(body: unknown, idempotencyKey: string): Promise<Order> {
    const { data } = await http.post<Envelope<Order>>("/orders", body, {
      headers: { "Idempotency-Key": idempotencyKey },
    });
    return unwrap(data);
  },
  async myOrders(): Promise<Order[]> {
    const { data } = await http.get<Envelope<Order[]>>("/orders");
    return unwrap(data) ?? [];
  },
  async order(code: string): Promise<Order> {
    const { data } = await http.get<Envelope<Order>>(`/orders/${code}`);
    return unwrap(data);
  },
};

export const qk = {
  menu: (lang: string) => ["menu", lang] as const,
  product: (slug: string, lang: string) => ["product", slug, lang] as const,
  banners: () => ["banners"] as const,
  languages: () => ["languages"] as const,
  addresses: () => ["addresses"] as const,
  myOrders: () => ["orders", "mine"] as const,
  order: (code: string) => ["order", code] as const,
};
