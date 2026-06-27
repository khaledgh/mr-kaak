import type { PageMeta } from "@/api/types";

interface Props {
  meta: PageMeta;
  onPage: (page: number) => void;
}

export function Pagination({ meta, onPage }: Props) {
  if (meta.total_pages <= 1) return null;

  return (
    <div className="flex items-center justify-between px-4 py-3 text-sm text-gray-600">
      <span>
        {(meta.page - 1) * meta.per_page + 1}–
        {Math.min(meta.page * meta.per_page, meta.total)} of {meta.total}
      </span>
      <div className="flex gap-1">
        <button
          className="btn-ghost px-2 py-1 text-xs"
          disabled={meta.page <= 1}
          onClick={() => onPage(meta.page - 1)}
        >
          ‹ Prev
        </button>
        <button
          className="btn-ghost px-2 py-1 text-xs"
          disabled={meta.page >= meta.total_pages}
          onClick={() => onPage(meta.page + 1)}
        >
          Next ›
        </button>
      </div>
    </div>
  );
}
