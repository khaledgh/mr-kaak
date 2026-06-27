import { createPortal } from "react-dom";
import { useToast } from "@/stores/toast";

export function ToastContainer() {
  const { toasts, dismiss } = useToast();
  if (!toasts.length) return null;

  return createPortal(
    <div className="pointer-events-none fixed bottom-20 left-1/2 z-50 flex -translate-x-1/2 flex-col items-center gap-2 md:bottom-6">
      {toasts.map((t) => (
        <button
          key={t.id}
          onClick={() => dismiss(t.id)}
          className="pointer-events-auto flex items-center gap-2 rounded-full bg-brand-800 px-5 py-3 text-sm font-medium text-white shadow-xl animate-toast"
        >
          <svg className="h-4 w-4 shrink-0" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
          </svg>
          {t.message}
        </button>
      ))}
    </div>,
    document.body,
  );
}
