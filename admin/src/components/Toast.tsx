import { createPortal } from "react-dom";
import { useToast } from "@/stores/toast";

export function ToastContainer() {
  const { toasts, dismiss } = useToast();
  if (toasts.length === 0) return null;

  return createPortal(
    <div className="fixed bottom-6 right-6 z-[9999] flex flex-col gap-2">
      {toasts.map((t) => (
        <div
          key={t.id}
          className={`animate-toast flex items-center gap-3 rounded-xl px-4 py-3 text-sm font-medium shadow-lg ${
            t.type === "error"
              ? "bg-red-600 text-white"
              : t.type === "info"
              ? "bg-gray-800 text-white"
              : "bg-gray-900 text-white"
          }`}
          onClick={() => dismiss(t.id)}
        >
          <span className="text-base">
            {t.type === "error" ? "✕" : t.type === "info" ? "ℹ" : "✓"}
          </span>
          {t.message}
        </div>
      ))}
    </div>,
    document.body,
  );
}
