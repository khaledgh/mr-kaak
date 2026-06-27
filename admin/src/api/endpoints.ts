import { http, unwrap } from "./client";
import type {
  AuthResponse, Banner, Category, Coupon, DeliveryZone,
  Envelope, Language, MediaItem, Order, PageMeta,
  Product, Settings, User,
} from "./types";

// ── helper for paginated list responses ──
function paged<T>(env: { data?: T; meta?: PageMeta }) {
  return {
    items: (env.data as T) ?? ([] as unknown as T),
    meta: env.meta ?? { page: 1, per_page: 20, total: 0, total_pages: 1 },
  };
}

export const api = {
  // Auth
  async login(body: { email: string; password: string }): Promise<AuthResponse> {
    const { data } = await http.post<Envelope<AuthResponse>>("/auth/login", body);
    return unwrap(data);
  },

  // Media
  async uploadMedia(file: File): Promise<MediaItem> {
    const form = new FormData();
    form.append("file", file);
    const { data } = await http.post<Envelope<MediaItem>>("/admin/media", form, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return unwrap(data);
  },
  async mediaList(params?: { page?: number; per_page?: number; q?: string }) {
    const { data } = await http.get<Envelope<MediaItem[]>>("/admin/media", { params });
    return paged<MediaItem[]>(data);
  },
  async deleteMedia(id: number) {
    await http.delete(`/admin/media/${id}`);
  },

  // Categories
  async categories(params?: { page?: number; per_page?: number }) {
    const { data } = await http.get<Envelope<Category[]>>("/admin/categories", { params });
    return paged<Category[]>(data);
  },
  async createCategory(body: Partial<Category>): Promise<Category> {
    const { data } = await http.post<Envelope<Category>>("/admin/categories", body);
    return unwrap(data);
  },
  async updateCategory(id: number, body: Partial<Category>): Promise<Category> {
    const { data } = await http.put<Envelope<Category>>(`/admin/categories/${id}`, body);
    return unwrap(data);
  },
  async deleteCategory(id: number) {
    await http.delete(`/admin/categories/${id}`);
  },
  async toggleCategory(id: number, is_active: boolean): Promise<Category> {
    const { data } = await http.patch<Envelope<Category>>(`/admin/categories/${id}`, { is_active });
    return unwrap(data);
  },

  // Products
  async products(params?: { page?: number; per_page?: number; category_id?: number; q?: string }) {
    const { data } = await http.get<Envelope<Product[]>>("/admin/products", { params });
    return paged<Product[]>(data);
  },
  async product(id: number): Promise<Product> {
    const { data } = await http.get<Envelope<Product>>(`/admin/products/${id}`);
    return unwrap(data);
  },
  async createProduct(body: Partial<Product>): Promise<Product> {
    const { data } = await http.post<Envelope<Product>>("/admin/products", body);
    return unwrap(data);
  },
  async updateProduct(id: number, body: Partial<Product>): Promise<Product> {
    const { data } = await http.put<Envelope<Product>>(`/admin/products/${id}`, body);
    return unwrap(data);
  },
  async deleteProduct(id: number) {
    await http.delete(`/admin/products/${id}`);
  },
  async toggleProduct(id: number, is_available: boolean): Promise<Product> {
    const { data } = await http.patch<Envelope<Product>>(`/admin/products/${id}/availability`, { is_available });
    return unwrap(data);
  },
  async flushCache() {
    await http.post("/admin/cache/flush");
  },

  // Coupons
  async coupons(params?: { page?: number; per_page?: number }) {
    const { data } = await http.get<Envelope<Coupon[]>>("/admin/coupons", { params });
    return paged<Coupon[]>(data);
  },
  async createCoupon(body: Partial<Coupon>): Promise<Coupon> {
    const { data } = await http.post<Envelope<Coupon>>("/admin/coupons", body);
    return unwrap(data);
  },
  async updateCoupon(id: number, body: Partial<Coupon>): Promise<Coupon> {
    const { data } = await http.put<Envelope<Coupon>>(`/admin/coupons/${id}`, body);
    return unwrap(data);
  },
  async deleteCoupon(id: number) {
    await http.delete(`/admin/coupons/${id}`);
  },

  // Banners
  async banners(params?: { page?: number; per_page?: number }) {
    const { data } = await http.get<Envelope<Banner[]>>("/admin/banners", { params });
    return paged<Banner[]>(data);
  },
  async createBanner(body: Partial<Banner>): Promise<Banner> {
    const { data } = await http.post<Envelope<Banner>>("/admin/banners", body);
    return unwrap(data);
  },
  async updateBanner(id: number, body: Partial<Banner>): Promise<Banner> {
    const { data } = await http.put<Envelope<Banner>>(`/admin/banners/${id}`, body);
    return unwrap(data);
  },
  async deleteBanner(id: number) {
    await http.delete(`/admin/banners/${id}`);
  },

  async deliveryZones() {
    const { data } = await http.get<Envelope<DeliveryZone[]>>("/admin/delivery-zones");
    return (unwrap(data) ?? []) as DeliveryZone[];
  },
  async getZone(id: number): Promise<DeliveryZone> {
    const zones = await api.deliveryZones();
    const zone = zones.find((z) => z.id === id);
    if (!zone) throw new Error("Zone not found");
    return zone;
  },
  async createZone(body: Partial<DeliveryZone>): Promise<DeliveryZone> {
    const { data } = await http.post<Envelope<DeliveryZone>>("/admin/delivery-zones", body);
    return unwrap(data);
  },
  async updateZone(id: number, body: Partial<DeliveryZone>): Promise<DeliveryZone> {
    const { data } = await http.put<Envelope<DeliveryZone>>(`/admin/delivery-zones/${id}`, body);
    return unwrap(data);
  },
  async deleteZone(id: number) {
    await http.delete(`/admin/delivery-zones/${id}`);
  },

  // Languages
  async languages() {
    const { data } = await http.get<Envelope<Language[]>>("/admin/languages");
    return (unwrap(data) ?? []) as Language[];
  },
  async createLanguage(body: Partial<Language>): Promise<Language> {
    const { data } = await http.post<Envelope<Language>>("/admin/languages", body);
    return unwrap(data);
  },
  async updateLanguage(id: number, body: Partial<Language>): Promise<Language> {
    const { data } = await http.put<Envelope<Language>>(`/admin/languages/${id}`, body);
    return unwrap(data);
  },
  async deleteLanguage(id: number) {
    await http.delete(`/admin/languages/${id}`);
  },
  async uiBundle(locale: string): Promise<Record<string, string>> {
    const { data } = await http.get<Envelope<Record<string, string>>>(`/i18n/${locale}`);
    return unwrap(data) ?? {};
  },
  async saveUiBundle(locale: string, bundle: Record<string, string>) {
    await http.put(`/admin/languages/${locale}/bundle`, bundle);
  },

  // Users
  async users(params?: { page?: number; per_page?: number; q?: string; role?: string }) {
    const { data } = await http.get<Envelope<User[]>>("/admin/users", { params });
    return paged<User[]>(data);
  },
  async updateUserRole(id: number, role: string): Promise<User> {
    const { data } = await http.patch<Envelope<User>>(`/admin/users/${id}/role`, { role });
    return unwrap(data);
  },
  async suspendUser(id: number): Promise<User> {
    const { data } = await http.post<Envelope<User>>(`/admin/users/${id}/suspend`);
    return unwrap(data);
  },
  async activateUser(id: number): Promise<User> {
    const { data } = await http.post<Envelope<User>>(`/admin/users/${id}/activate`);
    return unwrap(data);
  },

  // Settings
  async getSettings(): Promise<Settings> {
    const { data } = await http.get<Envelope<Settings>>("/admin/settings");
    return unwrap(data);
  },
  async saveSettings(body: Partial<Settings>): Promise<Settings> {
    const { data } = await http.put<Envelope<Settings>>("/admin/settings", body);
    return unwrap(data);
  },

  // Orders (existing)
  async adminOrders(params: { status?: string; page?: number; per_page?: number }) {
    const { data } = await http.get<Envelope<Order[]>>("/admin/orders", { params });
    return {
      orders: unwrap(data) ?? [],
      meta: data.meta ?? { page: 1, per_page: 50, total: 0, total_pages: 1 },
    };
  },
  async adminOrder(code: string): Promise<Order> {
    const { data } = await http.get<Envelope<Order>>(`/admin/orders/${code}`);
    return unwrap(data);
  },
  async advanceStatus(id: number, status: string, note?: string): Promise<Order> {
    const { data } = await http.post<Envelope<Order>>(`/admin/orders/${id}/status`, {
      status, note: note || undefined,
    });
    return unwrap(data);
  },
};

export const qk = {
  adminOrders:  (status?: string)  => ["admin-orders", status ?? "all"] as const,
  adminOrder:   (code: string)     => ["admin-order", code] as const,
  media:        (q?: string)       => ["media", q ?? ""] as const,
  categories:   ()                 => ["categories"] as const,
  products:     (catId?: number)   => ["products", catId ?? 0] as const,
  product:      (id: number)       => ["product", id] as const,
  coupons:      ()                 => ["coupons"] as const,
  banners:      ()                 => ["banners"] as const,
  zones:        ()                 => ["delivery-zones"] as const,
  languages:    ()                 => ["languages"] as const,
  uiBundle:     (locale: string)   => ["ui-bundle", locale] as const,
  users:        (q?: string)       => ["users", q ?? ""] as const,
  settings:     ()                 => ["settings"] as const,
};
