import type { Metadata } from "next";
import RequireAdmin from "@/components/admin/RequireAdmin";

export const metadata: Metadata = {
  title: "Admin · Reload",
  robots: { index: false },
};

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <RequireAdmin>
      <main className="mx-auto w-full max-w-[1440px] flex-1 px-4 py-10 sm:px-6 md:px-12">
        <p className="text-xs font-bold uppercase tracking-widest text-neon-lime">
          Admin
        </p>
        {children}
      </main>
    </RequireAdmin>
  );
}
