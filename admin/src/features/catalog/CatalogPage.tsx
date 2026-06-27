import { Link } from "react-router-dom";
import { PageHeader } from "@/components/PageHeader";

export function CatalogPage() {
  return (
    <div>
      <PageHeader title="Catalog" description="Manage your menu categories and products." />
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
        {[
          { to: "/catalog/categories", label: "Categories", icon: "🗂️", desc: "Organize your menu sections" },
          { to: "/catalog/products",   label: "Products",   icon: "🍬", desc: "Items, variants & modifier groups" },
        ].map((item) => (
          <Link
            key={item.to}
            to={item.to}
            className="card flex flex-col gap-2 p-6 hover:ring-brand-300 transition"
          >
            <span className="text-3xl">{item.icon}</span>
            <p className="font-semibold">{item.label}</p>
            <p className="text-xs text-gray-400">{item.desc}</p>
          </Link>
        ))}
      </div>
    </div>
  );
}
