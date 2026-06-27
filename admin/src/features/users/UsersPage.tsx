import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, qk } from "@/api/endpoints";
import { useToast } from "@/stores/toast";
import { PageHeader } from "@/components/PageHeader";
import { EmptyState } from "@/components/EmptyState";
import { Pagination } from "@/components/Pagination";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import type { User } from "@/api/types";

const ROLES = ["customer", "kitchen", "staff", "admin", "super_admin"];

export function UsersPage() {
  const [page, setPage] = useState(1);
  const [q, setQ] = useState("");
  const [roleFilter, setRoleFilter] = useState("");
  const [toSuspend, setToSuspend] = useState<User | null>(null);
  const { show } = useToast();
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: [...qk.users(q), page, roleFilter],
    queryFn: () => api.users({ q, page, per_page: 20, role: roleFilter || undefined }),
  });

  const changeRole = useMutation({
    mutationFn: ({ id, role }: { id: number; role: string }) => api.updateUserRole(id, role),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ["users"] }); show("Role updated"); },
    onError: () => show("Update failed", "error"),
  });

  const suspend = useMutation({
    mutationFn: (u: User) => u.status === "suspended" ? api.activateUser(u.id) : api.suspendUser(u.id),
    onSuccess: (_, u) => { qc.invalidateQueries({ queryKey: ["users"] }); setToSuspend(null); show(u.status === "suspended" ? "Activated" : "Suspended"); },
    onError: () => show("Action failed", "error"),
  });

  return (
    <div>
      <PageHeader title="Users" description="Manage roles and account status." />

      <div className="mb-4 flex flex-wrap gap-2">
        <input className="input max-w-xs" placeholder="Search name or email…" value={q} onChange={(e) => { setQ(e.target.value); setPage(1); }} />
        <select className="input max-w-[150px]" value={roleFilter} onChange={(e) => { setRoleFilter(e.target.value); setPage(1); }}>
          <option value="">All roles</option>
          {ROLES.map((r) => <option key={r} value={r}>{r}</option>)}
        </select>
      </div>

      {isLoading ? (
        <div className="space-y-2">{Array.from({ length: 6 }).map((_, i) => <div key={i} className="skeleton h-12 rounded-xl" />)}</div>
      ) : !data?.items.length ? (
        <EmptyState icon="👥" title="No users found" />
      ) : (
        <>
          <div className="table-wrap">
            <table className="table">
              <thead><tr><th>User</th><th>Email</th><th>Role</th><th>Status</th><th></th></tr></thead>
              <tbody>
                {data.items.map((u) => (
                  <tr key={u.id}>
                    <td className="font-medium">{u.name}</td>
                    <td className="text-xs text-gray-500">{u.email}</td>
                    <td>
                      <select
                        className="rounded-lg border border-gray-200 bg-transparent px-1.5 py-1 text-xs focus:outline-none"
                        value={u.role}
                        onChange={(e) => changeRole.mutate({ id: u.id, role: e.target.value })}
                      >
                        {ROLES.map((r) => <option key={r} value={r}>{r}</option>)}
                      </select>
                    </td>
                    <td>
                      <span className={`badge ${u.status === "active" ? "badge-green" : u.status === "suspended" ? "badge-red" : "badge-gray"}`}>
                        {u.status}
                      </span>
                    </td>
                    <td>
                      <button
                        className={`btn-ghost py-1 px-2 text-xs ${u.status === "suspended" ? "text-green-600" : "text-red-500"}`}
                        onClick={() => setToSuspend(u)}
                      >
                        {u.status === "suspended" ? "Activate" : "Suspend"}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {data.meta && <Pagination meta={data.meta} onPage={setPage} />}
        </>
      )}

      <ConfirmDialog
        open={!!toSuspend}
        title={toSuspend?.status === "suspended" ? "Activate user?" : "Suspend user?"}
        message={toSuspend?.status === "suspended"
          ? `Re-activate "${toSuspend?.name}"?`
          : `Suspend "${toSuspend?.name}"? They will not be able to log in.`}
        confirmLabel={toSuspend?.status === "suspended" ? "Activate" : "Suspend"}
        danger={toSuspend?.status !== "suspended"}
        loading={suspend.isPending}
        onConfirm={() => toSuspend && suspend.mutate(toSuspend)}
        onCancel={() => setToSuspend(null)}
      />
    </div>
  );
}
