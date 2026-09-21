"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { CartIcon, HeartIcon } from "@/components/GameActions";
import { useCart } from "@/hooks/useCart";

/**
 * NavShop is the wishlist and cart links in the navbar. Both show for
 * signed-out visitors too: the pages themselves send them to sign in.
 */
export default function NavShop() {
  const cart = useCart();
  const pathname = usePathname();

  // Hide the badge until the cart has loaded, rather than flashing a zero.
  const count = cart.loading ? 0 : cart.items.length;
  const cartLabel =
    count === 0 ? "Cart" : `Cart, ${count} ${count === 1 ? "item" : "items"}`;

  return (
    <div className="flex items-center gap-1">
      <NavIconLink
        href="/wishlist"
        label="Wishlist"
        active={pathname === "/wishlist"}
      >
        <HeartIcon className="h-5 w-5" />
      </NavIconLink>
      <NavIconLink href="/cart" label={cartLabel} active={pathname === "/cart"}>
        <CartIcon className="h-5 w-5" />
        {count > 0 && (
          <span
            aria-hidden
            className="absolute -right-0.5 -top-0.5 flex h-[18px] min-w-[18px] items-center justify-center rounded-full bg-neon-lime px-1 text-[10px] font-black leading-none text-background"
          >
            {count > 9 ? "9+" : count}
          </span>
        )}
      </NavIconLink>
    </div>
  );
}

function NavIconLink({
  href,
  label,
  active,
  children,
}: {
  href: string;
  label: string;
  active: boolean;
  children: React.ReactNode;
}) {
  return (
    <Link
      href={href}
      aria-label={label}
      title={label}
      aria-current={active ? "page" : undefined}
      className={`relative flex h-10 w-10 items-center justify-center rounded-xl transition-colors hover:bg-surface-container hover:text-on-surface ${
        active ? "text-on-surface" : "text-on-surface-variant"
      }`}
    >
      {children}
    </Link>
  );
}
