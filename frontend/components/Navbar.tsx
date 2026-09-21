import Image from "next/image";
import Link from "next/link";
import NavAccount from "@/components/NavAccount";
import NavShop from "@/components/NavShop";

export default function Navbar() {
  return (
    <header className="sticky top-0 z-50 border-b border-white/10 bg-background/80 backdrop-blur-xl">
      <div className="mx-auto flex h-16 max-w-[1440px] items-center justify-between px-6 md:px-12">
        <Link href="/" className="flex items-center gap-3">
          <Image
            src="/brand_assets/reload-logo.png"
            alt="Reload"
            width={32}
            height={32}
            className="h-8 w-8 object-contain"
            preload
          />
          <span className="text-lg font-black uppercase tracking-tight text-on-surface">
            Reload
          </span>
        </Link>
        <nav className="flex items-center gap-3 sm:gap-6">
          {/* The logo already links home, so this gives way on phones to
              leave room for the cart and account. */}
          <Link
            href="/"
            className="hidden text-sm font-semibold text-on-surface-variant transition-colors hover:text-on-surface sm:inline"
          >
            All Games
          </Link>
          <NavShop />
          <NavAccount />
        </nav>
      </div>
    </header>
  );
}
