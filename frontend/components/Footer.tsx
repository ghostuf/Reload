import Image from "next/image";

export default function Footer() {
  return (
    <footer className="mt-auto border-t border-outline-variant bg-surface-container-lowest">
      <div className="mx-auto flex max-w-[1440px] flex-col items-center gap-4 px-6 py-12 text-center md:px-12">
        <div className="flex items-center gap-2">
          <Image
            src="/brand_assets/reload-logo.png"
            alt="Reload"
            width={24}
            height={24}
            className="h-6 w-6 object-contain"
          />
          <span className="font-black uppercase tracking-tight text-on-surface">
            Reload
          </span>
        </div>
        <p className="max-w-md text-sm text-on-surface-variant">
          A university e-commerce project. Browse the catalog — purchases
          aren&apos;t available yet.
        </p>
      </div>
    </footer>
  );
}
