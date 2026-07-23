import Image from "next/image";
import Link from "next/link";

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
        <nav className="flex items-center gap-6">
          <Link
            href="/"
            className="text-sm font-semibold text-on-surface-variant transition-colors hover:text-on-surface"
          >
            All Games
          </Link>
        </nav>
      </div>
    </header>
  );
}
