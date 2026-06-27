import { useRef, useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, qk } from "@/api/endpoints";
import { useToast } from "@/stores/toast";
import { PageHeader } from "@/components/PageHeader";
import { EmptyState } from "@/components/EmptyState";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { Pagination } from "@/components/Pagination";
import type { MediaItem } from "@/api/types";

export function MediaPage() {
  const [page, setPage] = useState(1);
  const [q, setQ] = useState("");
  const [selected, setSelected] = useState<MediaItem | null>(null);
  const [toDelete, setToDelete] = useState<MediaItem | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);
  const { show } = useToast();
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: [...qk.media(q), page],
    queryFn: () => api.mediaList({ q, page, per_page: 24 }),
  });

  const upload = useMutation({
    mutationFn: (file: File) => api.uploadMedia(file),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["media"] }); show("Uploaded successfully"); },
    onError: (e) => show(e instanceof Error ? e.message : "Upload failed", "error"),
  });

  const del = useMutation({
    mutationFn: (id: number) => api.deleteMedia(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["media"] });
      setToDelete(null);
      if (selected?.id === toDelete?.id) setSelected(null);
      show("Deleted");
    },
    onError: () => show("Delete failed", "error"),
  });

  function onFiles(e: React.ChangeEvent<HTMLInputElement>) {
    const files = Array.from(e.target.files ?? []);
    for (const f of files) {
      if (f.size > 2 * 1024 * 1024) { show(`${f.name} exceeds 2 MB`, "error"); continue; }
      upload.mutate(f);
    }
    e.target.value = "";
  }

  function fmt(bytes: number) {
    return bytes < 1024 ? `${bytes} B` : `${(bytes / 1024).toFixed(0)} KB`;
  }

  return (
    <div>
      <PageHeader
        title="Media Center"
        description="Upload and manage images used across the admin."
        action={
          <button className="btn-primary" onClick={() => fileRef.current?.click()} disabled={upload.isPending}>
            {upload.isPending ? "Uploading…" : "Upload images"}
          </button>
        }
      />
      <input ref={fileRef} type="file" accept="image/jpeg,image/png,image/webp" multiple className="hidden" onChange={onFiles} />

      {/* Search */}
      <div className="mb-4">
        <input className="input max-w-xs" placeholder="Search…" value={q} onChange={(e) => { setQ(e.target.value); setPage(1); }} />
      </div>

      {isLoading ? (
        <div className="grid grid-cols-3 gap-3 sm:grid-cols-4 lg:grid-cols-6">
          {Array.from({ length: 12 }).map((_, i) => <div key={i} className="skeleton aspect-square rounded-xl" />)}
        </div>
      ) : !data?.items.length ? (
        <EmptyState icon="🖼️" title="No images yet" description="Upload your first image to get started." />
      ) : (
        <>
          <div className="grid grid-cols-3 gap-3 sm:grid-cols-4 lg:grid-cols-6">
            {data.items.map((item) => (
              <button
                key={item.id}
                className={`group relative aspect-square overflow-hidden rounded-xl border-2 transition ${
                  selected?.id === item.id ? "border-brand-500" : "border-transparent hover:border-brand-300"
                }`}
                onClick={() => setSelected(selected?.id === item.id ? null : item)}
              >
                <img src={item.thumb_url} alt={item.alt ?? item.original_name} className="h-full w-full object-cover" loading="lazy" />
              </button>
            ))}
          </div>
          <Pagination meta={data.meta} onPage={setPage} />
        </>
      )}

      {/* Detail panel */}
      {selected && (
        <div className="fixed bottom-6 right-6 w-64 rounded-2xl bg-white p-4 shadow-2xl ring-1 ring-gray-200">
          <img src={selected.url} alt={selected.alt ?? ""} className="mb-3 w-full rounded-lg object-cover" style={{ maxHeight: 160 }} />
          <p className="truncate text-sm font-medium">{selected.original_name}</p>
          <p className="mt-0.5 text-xs text-gray-400">{selected.width}×{selected.height} · {fmt(selected.size_bytes)}</p>
          <div className="mt-3 flex gap-2">
            <button
              className="btn-outline flex-1 text-xs"
              onClick={() => navigator.clipboard.writeText(selected.url).then(() => show("URL copied"))}
            >
              Copy URL
            </button>
            <button
              className="btn-danger text-xs px-2"
              onClick={() => setToDelete(selected)}
            >
              Delete
            </button>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={!!toDelete}
        title="Delete image?"
        message={`"${toDelete?.original_name}" will be removed. Existing uses won't be broken, but the file will be gone.`}
        confirmLabel="Delete"
        danger
        loading={del.isPending}
        onConfirm={() => toDelete && del.mutate(toDelete.id)}
        onCancel={() => setToDelete(null)}
      />
    </div>
  );
}
