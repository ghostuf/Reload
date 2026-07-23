---
name: Kinetic Dark
colors:
  surface: '#131313'
  surface-dim: '#131313'
  surface-bright: '#3a3939'
  surface-container-lowest: '#0e0e0e'
  surface-container-low: '#1c1b1b'
  surface-container: '#201f1f'
  surface-container-high: '#2a2a2a'
  surface-container-highest: '#353534'
  on-surface: '#e5e2e1'
  on-surface-variant: '#c5c5da'
  inverse-surface: '#e5e2e1'
  inverse-on-surface: '#313030'
  outline: '#8f8fa3'
  outline-variant: '#444557'
  surface-tint: '#bdc2ff'
  primary: '#bdc2ff'
  on-primary: '#00149f'
  primary-container: '#253cff'
  on-primary-container: '#d0d3ff'
  inverse-primary: '#293fff'
  secondary: '#d4f332'
  on-secondary: '#2b3400'
  secondary-container: '#b8d601'
  on-secondary-container: '#4c5a00'
  tertiary: '#d1bcff'
  on-tertiary: '#3c0090'
  tertiary-container: '#7107ff'
  on-tertiary-container: '#deceff'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#dfe0ff'
  primary-fixed-dim: '#bdc2ff'
  on-primary-fixed: '#000965'
  on-primary-fixed-variant: '#001fdd'
  secondary-fixed: '#d1f02f'
  secondary-fixed-dim: '#b6d300'
  on-secondary-fixed: '#181e00'
  on-secondary-fixed-variant: '#404c00'
  tertiary-fixed: '#e9ddff'
  tertiary-fixed-dim: '#d1bcff'
  on-tertiary-fixed: '#23005b'
  on-tertiary-fixed-variant: '#5700c9'
  background: '#131313'
  on-background: '#e5e2e1'
  surface-variant: '#353534'
typography:
  display-lg:
    fontFamily: Inter
    fontSize: 64px
    fontWeight: '800'
    lineHeight: '1.1'
    letterSpacing: -0.04em
  display-md:
    fontFamily: Inter
    fontSize: 48px
    fontWeight: '800'
    lineHeight: '1.1'
    letterSpacing: -0.03em
  headline-lg:
    fontFamily: Inter
    fontSize: 32px
    fontWeight: '700'
    lineHeight: '1.2'
    letterSpacing: -0.02em
  headline-lg-mobile:
    fontFamily: Inter
    fontSize: 28px
    fontWeight: '700'
    lineHeight: '1.2'
    letterSpacing: -0.02em
  headline-md:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: '700'
    lineHeight: '1.3'
  body-lg:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: '400'
    lineHeight: '1.6'
  body-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: '1.6'
  label-md:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '600'
    lineHeight: '1'
    letterSpacing: 0.05em
  label-sm:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: '500'
    lineHeight: '1'
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  unit: 4px
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 40px
  xxl: 64px
  container-max: 1440px
  gutter: 24px
  margin-mobile: 16px
  margin-desktop: 48px
---

## Brand & Style

This design system establishes a high-octane, premium gaming aesthetic that balances the raw energy of competitive play with the sophisticated precision of modern SaaS tools. The visual language is defined by deep atmospheric backgrounds, "Electric Blue" core energy, and high-visibility "Neon Lime" accents.

The style is **Modern-Gaming-Minimalism**. It borrows the depth and immersion of gaming storefronts (Steam, Epic) while adopting the clean layouts and rigorous grid systems of developer-centric platforms (Vercel, Linear). Key characteristics include:
- **Atmospheric Depth:** Using glassmorphism and subtle background blurs to create a multi-layered interface.
- **High-Contrast Precision:** Crisp, bold typography paired with neon highlights to guide user focus.
- **Fluid Motion:** Surfaces should feel light and responsive, using generous whitespace to prevent the dark interface from feeling cramped.

## Colors

The palette is optimized for dark-mode environments, emphasizing depth through tonal layers rather than heavy borders.

- **Primary (Electric Blue):** The core brand color, used for primary actions, active states, and brand signatures.
- **Accent (Neon Lime):** High-visibility color reserved strictly for call-to-actions, conversion points, and critical highlights.
- **Surface Hierarchy:** 
  - `Base (#111111)` is the canvas.
  - `Surface (#1B1B1B)` is for primary content containers.
  - `Elevated (#252525)` is for cards, popovers, and interactive elements.
- **Functional Gradients:** Use the blue-to-purple gradient for immersive hero backgrounds and premium tier features.

## Typography

This design system utilizes **Inter** for its neutral, geometric qualities that perform exceptionally well in high-density data environments and large-scale marketing headlines.

- **Display & Headlines:** Use heavy weights (700-800) with tight letter spacing for a forceful, cinematic presence.
- **Body:** Standardized at 16px for optimal legibility against dark backgrounds.
- **Labels:** Small caps and increased tracking are used for metadata, categories, and utility text to distinguish them from narrative body copy.

## Layout & Spacing

The layout follows a **Fluid Grid** philosophy with fixed maximum widths for content readability.

- **Grid System:** A 12-column grid for desktop with 24px gutters. Use 4-column layouts for mobile.
- **Whitespace:** Emphasize "Generous Whitespace." Hero sections should use `xxl` (64px) padding to create a premium, unhurried feel. 
- **Rhythm:** All spacing must be multiples of 4px. Use `md` (16px) as the standard padding for cards and `lg` (24px) for section vertical spacing.
- **Safe Zones:** Content is center-aligned on desktop with a maximum width of 1440px to ensure accessibility on ultra-wide monitors.

## Elevation & Depth

Hierarchy is established through "Surface Stacking" and "Luminous Depth."

- **Glassmorphism:** Navigation bars and modal overlays use a 20px backdrop blur with a 10% white tint. This maintains context while separating the active layer.
- **Tonal Tiers:** Higher elevation elements use lighter shades of Gray (moving from #111111 to #252525).
- **Soft Shadows:** Interactive elements (cards on hover) use an extra-diffused shadow: `0 20px 40px rgba(0, 0, 0, 0.4)`. 
- **Luminous Glow:** Primary buttons and active states may use a subtle outer glow of Electric Blue (`0 0 15px rgba(37, 60, 255, 0.3)`) to simulate a digital screen effect.

## Shapes

The shape language is consistently "Soft-Geometric." 

- **Containers:** Large cards and main content areas utilize a **20px radius** to feel modern and friendly.
- **Interactive Elements:** Buttons and inputs use a slightly sharper **12px radius** to denote clickability and precision.
- **Minimal Borders:** Avoid high-contrast borders. Use 1px borders with 10% opacity white or 10% opacity primary color only when necessary for structural definition on identical background colors.

## Components

- **Primary Button:** Background: Electric Blue (#253CFF); Text: White; Radius: 12px. High-priority actions only.
- **CTA Button (Accent):** Background: Neon Lime (#DFFF3F); Text: Black (#111111); Weight: Bold. Use for "Buy Now" or "Play" actions.
- **Secondary Button:** Background: Transparent; Border: 1px Solid #FFFFFF (20% Opacity); Text: White.
- **Input Fields:** Background: #1B1B1B; Border: 1px Solid #252525; Focus State: Border Color changes to Electric Blue with a subtle 2px outer glow.
- **Game Cards:** 20px corner radius. Image fills the container. On hover, the card scales 2% and the Electric Blue shadow is applied.
- **Chips/Badges:** Small, 100px radius (pill). Dark gray background with Electric Blue text for categories, or Neon Lime background for "Sale" or "New" tags.
- **Lists:** Clean rows with 1px #252525 bottom border. Subtle #1B1B1B background fill on hover.