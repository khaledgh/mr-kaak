import { useEffect, useRef, useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams, Link } from "react-router-dom";
import L from "leaflet";
import { api, qk } from "@/api/endpoints";
import { useToast } from "@/stores/toast";
import { FormField } from "@/components/FormField";
import type { DeliveryZone } from "@/api/types";

function blankForm(): Partial<DeliveryZone> & { fee_str: string; min_str: string } {
  return { name: "", scope: "global", shape: "radius", center_lat: 49.2827, center_lng: -123.1207, radius_km: 5, fee_cents: 0, fee_str: "0", min_order_cents: 0, min_str: "0", is_active: true };
}

function c2d(c?: number) { return c != null ? (c / 100).toFixed(2) : "0"; }
function d2c(s: string) { return Math.round(parseFloat(s || "0") * 100); }

const customMarkerIcon = L.divIcon({
  className: "custom-map-pin",
  html: `<div style="background-color: #ea580c; width: 16px; height: 16px; border-radius: 50%; border: 3px solid white; box-shadow: 0 0 0 2px #ea580c, 0 3px 6px rgba(0,0,0,0.3); pointer-events: none;"></div>`,
  iconSize: [16, 16],
  iconAnchor: [8, 8]
});

interface ZoneMapProps {
  lat: number;
  lng: number;
  radiusKm: number;
  onChange: (lat: number, lng: number) => void;
}

function ZoneMap({ lat, lng, radiusKm, onChange }: ZoneMapProps) {
  const mapRef = useRef<HTMLDivElement>(null);
  const mapInstance = useRef<L.Map | null>(null);
  const markerInstance = useRef<L.Marker | null>(null);
  const circleInstance = useRef<L.Circle | null>(null);

  useEffect(() => {
    if (!mapRef.current) return;

    const initialLat = lat || 49.2827;
    const initialLng = lng || -123.1207;

    const map = L.map(mapRef.current).setView([initialLat, initialLng], 12);
    mapInstance.current = map;

    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
    }).addTo(map);

    const marker = L.marker([initialLat, initialLng], {
      draggable: true,
      icon: customMarkerIcon
    }).addTo(map);
    markerInstance.current = marker;

    const circle = L.circle([initialLat, initialLng], {
      color: "#ea580c",
      fillColor: "#f97316",
      fillOpacity: 0.15,
      weight: 2,
      radius: radiusKm * 1000
    }).addTo(map);
    circleInstance.current = circle;

    // Sync state on drag
    marker.on("drag", () => {
      const position = marker.getLatLng();
      circle.setLatLng(position);
    });

    marker.on("dragend", () => {
      const position = marker.getLatLng();
      onChange(position.lat, position.lng);
    });

    // Map click selection
    map.on("click", (e) => {
      const { lat: clickLat, lng: clickLng } = e.latlng;
      marker.setLatLng([clickLat, clickLng]);
      circle.setLatLng([clickLat, clickLng]);
      onChange(clickLat, clickLng);
    });

    // Wait a brief moment to invalidate size so that Leaflet renders full bounds on page load
    const timer = setTimeout(() => {
      if (mapInstance.current) {
        mapInstance.current.invalidateSize();
      }
    }, 100);

    return () => {
      clearTimeout(timer);
      map.remove();
      mapInstance.current = null;
      markerInstance.current = null;
      circleInstance.current = null;
    };
  }, []);

  // Sync state changes from inputs back to map
  useEffect(() => {
    const map = mapInstance.current;
    const marker = markerInstance.current;
    const circle = circleInstance.current;

    if (!map || !marker || !circle) return;

    const currentLat = lat || 49.2827;
    const currentLng = lng || -123.1207;

    const currentMarkerLatLng = marker.getLatLng();
    if (Math.abs(currentMarkerLatLng.lat - currentLat) > 0.000001 || Math.abs(currentMarkerLatLng.lng - currentLng) > 0.000001) {
      const newPos = L.latLng(currentLat, currentLng);
      marker.setLatLng(newPos);
      circle.setLatLng(newPos);
      map.panTo(newPos);
    }

    circle.setRadius((radiusKm || 0.1) * 1000);
  }, [lat, lng, radiusKm]);

  return (
    <div style={{ height: "450px" }} className="w-full rounded-2xl overflow-hidden border border-gray-200 relative z-10 shadow-inner">
      <div ref={mapRef} style={{ height: "100%", width: "100%" }} />
    </div>
  );
}

export function DeliveryFormPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const { show } = useToast();
  
  const isEdit = !!id;
  const [form, setForm] = useState(blankForm());

  const { data: zones = [], isLoading: loadingZones } = useQuery({
    queryKey: qk.zones(),
    queryFn: api.deliveryZones,
    enabled: isEdit,
  });

  useEffect(() => {
    if (isEdit && zones.length > 0) {
      const zone = zones.find((z) => z.id === Number(id));
      if (zone) {
        setForm({
          ...zone,
          fee_str: c2d(zone.fee_cents),
          min_str: c2d(zone.min_order_cents),
        });
      } else {
        show("Zone not found", "error");
        navigate("/delivery");
      }
    }
  }, [isEdit, id, zones, navigate, show]);

  const save = useMutation({
    mutationFn: (body: Partial<DeliveryZone>) =>
      isEdit ? api.updateZone(Number(id), body) : api.createZone(body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.zones() });
      show(isEdit ? "Updated" : "Created");
      navigate("/delivery");
    },
    onError: (e) => show(e instanceof Error ? e.message : "Save failed", "error"),
  });

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!form.name?.trim()) {
      show("Zone name is required", "error");
      return;
    }
    save.mutate({
      name: form.name,
      scope: form.scope,
      shape: form.shape,
      center_lat: form.center_lat,
      center_lng: form.center_lng,
      radius_km: form.radius_km,
      fee_cents: d2c(form.fee_str ?? "0"),
      min_order_cents: d2c(form.min_str ?? "0"),
      is_active: form.is_active,
    });
  }

  if (isEdit && loadingZones) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="text-gray-500">Loading delivery zone details...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Link to="/delivery" className="flex items-center justify-center h-8 w-8 rounded-lg border border-gray-300 bg-white text-gray-600 hover:bg-gray-50 hover:text-gray-800 transition shadow-sm">
          <svg className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 19.5L8.25 12l7.5-7.5" />
          </svg>
        </Link>
        <h1 className="text-xl font-bold text-gray-900">{isEdit ? "Edit Delivery Zone" : "New Delivery Zone"}</h1>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
        {/* Form Column */}
        <div className="lg:col-span-2">
          <form onSubmit={submit} className="card p-6 space-y-4" noValidate>
            <FormField label="Zone name" required>
              <input
                className="input"
                placeholder="e.g. Downtown Area"
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              />
            </FormField>

            <div className="grid grid-cols-2 gap-4">
              <FormField label="Shape">
                <select
                  className="input"
                  value={form.shape}
                  onChange={(e) => setForm((f) => ({ ...f, shape: e.target.value as DeliveryZone["shape"] }))}
                >
                  <option value="radius">Radius</option>
                  <option value="polygon">Polygon (Static Coordinates)</option>
                </select>
              </FormField>

              <FormField label="Scope">
                <select
                  className="input"
                  value={form.scope}
                  onChange={(e) => setForm((f) => ({ ...f, scope: e.target.value as DeliveryZone["scope"] }))}
                >
                  <option value="global">Global</option>
                  <option value="product">Product-specific</option>
                </select>
              </FormField>
            </div>

            {form.shape === "radius" && (
              <div className="grid grid-cols-3 gap-3">
                <FormField label="Center lat">
                  <input
                    type="number"
                    step="any"
                    className="input"
                    value={isNaN(form.center_lat ?? 0) ? "" : form.center_lat}
                    onChange={(e) => {
                      const v = parseFloat(e.target.value);
                      setForm((f) => ({ ...f, center_lat: isNaN(v) ? 0 : v }));
                    }}
                  />
                </FormField>
                <FormField label="Center lng">
                  <input
                    type="number"
                    step="any"
                    className="input"
                    value={isNaN(form.center_lng ?? 0) ? "" : form.center_lng}
                    onChange={(e) => {
                      const v = parseFloat(e.target.value);
                      setForm((f) => ({ ...f, center_lng: isNaN(v) ? 0 : v }));
                    }}
                  />
                </FormField>
                <FormField label="Radius (km)">
                  <input
                    type="number"
                    min={0.1}
                    step={0.1}
                    className="input"
                    value={isNaN(form.radius_km ?? 0) ? "" : form.radius_km}
                    onChange={(e) => {
                      const v = parseFloat(e.target.value);
                      setForm((f) => ({ ...f, radius_km: isNaN(v) ? 0.1 : v }));
                    }}
                  />
                </FormField>
              </div>
            )}

            <div className="grid grid-cols-2 gap-4">
              <FormField label="Delivery fee ($)">
                <input
                  type="number"
                  min={0}
                  step={0.01}
                  className="input"
                  value={form.fee_str}
                  onChange={(e) => setForm((f) => ({ ...f, fee_str: e.target.value }))}
                />
              </FormField>
              <FormField label="Min order ($)">
                <input
                  type="number"
                  min={0}
                  step={0.01}
                  className="input"
                  value={form.min_str}
                  onChange={(e) => setForm((f) => ({ ...f, min_str: e.target.value }))}
                />
              </FormField>
            </div>

            <label className="flex cursor-pointer items-center gap-2 text-sm pt-1">
              <button
                type="button"
                className={`switch ${form.is_active ? "bg-brand-500" : "bg-gray-200"}`}
                onClick={() => setForm((f) => ({ ...f, is_active: !f.is_active }))}
              >
                <span className={`switch-thumb ${form.is_active ? "translate-x-4" : "translate-x-0"}`} />
              </button>
              Active
            </label>

            <div className="flex justify-end gap-2 pt-4 border-t border-gray-100">
              <button
                type="button"
                className="btn-ghost"
                onClick={() => navigate("/delivery")}
              >
                Cancel
              </button>
              <button
                type="submit"
                className="btn-primary"
                disabled={save.isPending}
              >
                {save.isPending ? "Saving…" : "Save Zone"}
              </button>
            </div>
          </form>
        </div>

        {/* Map Column */}
        <div className="lg:col-span-3">
          {form.shape === "radius" ? (
            <div className="card p-6 space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold uppercase tracking-wider text-gray-500">Interactive Location Selection</span>
                <span className="badge badge-amber font-mono text-xs">
                  {form.center_lat?.toFixed(4)}, {form.center_lng?.toFixed(4)}
                </span>
              </div>
              <div className="relative">
                <ZoneMap
                  lat={form.center_lat ?? 49.2827}
                  lng={form.center_lng ?? -123.1207}
                  radiusKm={form.radius_km ?? 5}
                  onChange={(lat, lng) => setForm((f) => ({ ...f, center_lat: parseFloat(lat.toFixed(6)), center_lng: parseFloat(lng.toFixed(6)) }))}
                />
              </div>
            </div>
          ) : (
            <div className="card p-8 flex items-center justify-center text-center text-gray-500 h-64">
              <div>
                <p className="text-2xl mb-2">🗺️</p>
                <p className="text-sm font-medium">Polygon mode does not require center selection.</p>
                <p className="text-xs text-gray-400 mt-1">Specify coordinate points or boundaries in coordinates array.</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
