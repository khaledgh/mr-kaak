import { useRef, useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, qk } from "@/api/endpoints";
import { useToast } from "@/stores/toast";
import { Modal } from "./Modal";
import type { MediaItem } from "@/api/types";

interface Props {
  value?: string;           // current URL
  onChange: (url: string) => void;
  label?: string;
}

export function MediaPicker({ value, onChange, label = "Image" }: Props) {
  const [open, setOpen] = useState(false);
  const [q, setQ] = useState("");
  const { show } = useToast();
  const qc = useQueryClient();
  const inputRef = useRef<HTMLInputElement>(null);

  const { data, isLoading } = useQuery({
    queryKey: qk.media(q),
    queryFn: () => api.mediaList({ q, per_page: 40 }),
    enabled: open,
  });

  const upload = useMutation({
    mutationFn: (file: File) => api.uploadMedia(file),
    onSuccess: (item) => {
      qc.invalidateQueries({ queryKey: ["media"] });
      onChange(item.url);
      setOpen(false);
      show("Image uploaded");
    },
    onError: (e) => show(e instanceof Error ? e.message : "Upload failed", "error"),
  });

  function pick(item: MediaItem) {
    onChange(item.url);
    setOpen(false);
  }

  function onFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    if (file.size > 2 * 1024 * 1024) {
      show("File must be under 2 MB", "error");
      return;
    }
    upload.mutate(file);
  }

  return (
    <div>
      <label className="label">{label}</label>
      <div className="flex gap-2">
        <div
          className="h-20 w-20 shrink-0 cursor-pointer overflow-hidden rounded-lg border-2 border-dashed border-gray-300 bg-gray-50 hover:border-brand-400"
          onClick={() => setOpen(true)}
        >
          {value ? (
            <img src={value} alt="" className="h-full w-full object-cover" />
          ) : (
            <div className="flex h-full items-center justify-center text-2xl text-gray-300">+</div>
          )}
        </div>
        <div className="flex flex-col justify-center gap-1">
          <button type="button" className="btn-outline text-xs" onClick={() => setOpen(true)}>
            {value ? "Change image" : "Select image"}
          </button>
          {value && (
            <button type="button" className="btn-ghost text-xs text-red-500" onClick={() => onChange("")}>
              Remove
            </button>
          )}
        </div>
      </div>

      <Modal open={open} onClose={() => setOpen(false)} title="Media Library" size="xl">
        <div className="space-y-4">
          {/* Upload + search row */}
          <div className="flex gap-2">
            <input
              className="input"
              placeholder="Search by filename…"
              value={q}
              onChange={(e) => setQ(e.target.value)}
            />
            <button
              type="button"
              className="btn-primary shrink-0"
              onClick={() => inputRef.current?.click()}
              disabled={upload.isPending}
            >
              {upload.isPending ? "Uploading…" : "Upload"}
            </button>
            <input
              ref={inputRef}
              type="file"
              accept="image/jpeg,image/png,image/webp"
              className="hidden"
              onChange={onFileChange}
            />
          </div>

          {/* Grid */}
          {isLoading ? (
            <div className="grid grid-cols-4 gap-2">
              {Array.from({ length: 8 }).map((_, i) => (
                <div key={i} className="skeleton aspect-square rounded-lg" />
              ))}
            </div>
          ) : !data?.items.length ? (
            <p className="py-12 text-center text-sm text-gray-400">No images yet. Upload one above.</p>
          ) : (
            <div className="grid grid-cols-4 gap-2 max-h-96 overflow-y-auto">
              {data.items.map((item) => (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => pick(item)}
                  className={`group relative aspect-square overflow-hidden rounded-lg border-2 transition ${
                    value === item.url
                      ? "border-brand-500"
                      : "border-transparent hover:border-brand-300"
                  }`}
                >
                  <img
                    src={item.thumb_url}
                    alt={item.alt ?? item.original_name}
                    className="h-full w-full object-cover"
                    loading="lazy"
                  />
                  <div className="absolute inset-0 flex items-end bg-gradient-to-t from-black/40 opacity-0 transition group-hover:opacity-100">
                    <span className="truncate px-1 py-1 text-xs text-white">
                      {item.original_name}
                    </span>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>
      </Modal>
    </div>
  );
}
