import { redirect } from "next/navigation";

/** Games are the only thing to manage so far, so /admin opens straight onto them. */
export default function AdminIndex() {
  redirect("/admin/games");
}
