import { create } from "zustand";

interface Toast {
  id: number;
  message: string;
  type: "success" | "error" | "info";
}

interface ToastState {
  toasts: Toast[];
  show: (message: string, type?: Toast["type"]) => void;
  dismiss: (id: number) => void;
}

let _id = 0;

export const useToast = create<ToastState>((set) => ({
  toasts: [],
  show: (message, type = "success") => {
    const id = ++_id;
    set((s) => ({ toasts: [...s.toasts, { id, message, type }] }));
    setTimeout(
      () => set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) })),
      3000,
    );
  },
  dismiss: (id) => set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) })),
}));
