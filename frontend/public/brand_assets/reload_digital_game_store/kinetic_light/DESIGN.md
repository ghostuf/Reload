---
name: Kinetic Light
colors:
  surface: '#fcf9f8'
  surface-dim: '#dcd9d9'
  surface-bright: '#fcf9f8'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#f6f3f2'
  surface-container: '#f0edec'
  surface-container-high: '#ebe7e7'
  surface-container-highest: '#e5e2e1'
  on-surface: '#1c1b1b'
  on-surface-variant: '#444557'
  inverse-surface: '#313030'
  inverse-on-surface: '#f3f0ef'
  outline: '#757689'
  outline-variant: '#c5c5da'
  surface-tint: '#293fff'
  primary: '#001fd9'
  on-primary: '#ffffff'
  primary-container: '#253cff'
  on-primary-container: '#d0d3ff'
  inverse-primary: '#bdc2ff'
  secondary: '#566500'
  on-secondary: '#ffffff'
  secondary-container: '#ceed2c'
  on-secondary-container: '#5a6900'
  tertiary: '#881b00'
  on-tertiary: '#ffffff'
  tertiary-container: '#b22600'
  on-tertiary-container: '#ffcabd'
  error: '#ba1a1a'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#dfe0ff'
  primary-fixed-dim: '#bdc2ff'
  on-primary-fixed: '#000965'
  on-primary-fixed-variant: '#001fdd'
  secondary-fixed: '#d1f02f'
  secondary-fixed-dim: '#b6d300'
  on-secondary-fixed: '#181e00'
  on-secondary-fixed-variant: '#404c00'
  tertiary-fixed: '#ffdad2'
  tertiary-fixed-dim: '#ffb4a3'
  on-tertiary-fixed: '#3d0700'
  on-tertiary-fixed-variant: '#8a1b00'
  background: '#fcf9f8'
  on-background: '#1c1b1b'
  surface-variant: '#e5e2e1'
typography:
  headline-xl:
    fontFamily: Anybody
    fontSize: 64px
    fontWeight: '800'
    lineHeight: '1.1'
    letterSpacing: -0.04em
  headline-lg:
    fontFamily: Anybody
    fontSize: 40px
    fontWeight: '700'
    lineHeight: '1.2'
    letterSpacing: -0.02em
  headline-lg-mobile:
    fontFamily: Anybody
    fontSize: 32px
    fontWeight: '700'
    lineHeight: '1.2'
    letterSpacing: -0.02em
  body-lg:
    fontFamily: Hanken Grotesk
    fontSize: 18px
    fontWeight: '400'
    lineHeight: '1.6'
    letterSpacing: '0'
  body-md:
    fontFamily: Hanken Grotesk
    fontSize: 16px
    fontWeight: '400'
    lineHeight: '1.5'
    letterSpacing: '0'
  label-caps:
    fontFamily: Geist
    fontSize: 12px
    fontWeight: '600'
    lineHeight: '1'
    letterSpacing: 0.1em
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  base: 4px
  xs: 8px
  sm: 16px
  md: 24px
  lg: 48px
  xl: 80px
  gutter: 24px
  margin: 32px
---

## Brand & Style
This design system is built on "Kinetic Energy"—a high-impact, high-motion aesthetic that merges technical precision with aggressive marketing visuals. It is designed for high-growth tech, sports, or creative platforms where speed and clarity are paramount.

The style is a hybrid of **Modern Minimalism** and **High-Contrast Bold**. It utilizes a "White Room" foundation—vast expanses of pure white space—to allow the high-saturation brand colors to pop with maximum intensity. The emotional response is one of urgency, confidence, and cutting-edge reliability. All interactions should feel snappy, with sharp transitions and substantial visual weight.

## Colors
The palette is dominated by a stark, high-contrast relationship between the pure white background and deep black typography. 

- **Primary (Electric Blue):** Used for primary actions, active states, and structural highlights. It provides a "tech-forward" anchor to the layout.
- **Secondary (Neon Lime):** Reserved exclusively for high-priority Call-to-Actions (CTAs) and success states. It creates a vibrating contrast against both white and blue.
- **Surface & Container:** The base is pure `#FFFFFF`. To create hierarchy without shadows, use `#F5F5F5` for secondary content containers and section backgrounds. 
- **Neutral:** Use `#0F0F0F` for headings and primary body text to ensure maximum legibility and a "ink-on-paper" crispness.

## Typography
The typography is the primary driver of the "Kinetic" personality. 

- **Display & Headlines:** Use **Anybody**. Its variable-width nature and aggressive weight create a sense of forward motion. Headlines should be tightly leaded and slightly tracked in to create a dense, powerful block of text.
- **Body:** **Hanken Grotesk** provides a clean, contemporary feel that balances the intensity of the headlines with high readability.
- **Labels/UI Elements:** **Geist** is used for technical data, labels, and small UI details to reinforce the precision and "developer-friendly" aspect of the system.

## Layout & Spacing
The layout follows a **Fluid Grid** model with generous margins to preserve the minimalist aesthetic. 

- **Grid:** A 12-column system for desktop, 8-column for tablet, and 4-column for mobile.
- **Rhythm:** Use an 8px base unit for internal component spacing and a 24px (md) unit for layout gutters. 
- **White Space:** Use the `xl` (80px) spacing between major sections to emphasize the "Pure White" canvas. 
- **Alignment:** Lean into hard left-alignment for all text blocks to maintain a structured, editorial feel.

## Elevation & Depth
In this light-mode iteration, depth is achieved through **Tonal Layers** and **Bold Outlines** rather than traditional shadows.

1. **Level 0 (Base):** `#FFFFFF` - The main canvas.
2. **Level 1 (Surface):** `#F5F5F5` - Used for inset cards, background sections, and form fields.
3. **Level 2 (Interaction):** No shadows. Instead, use a 2px solid border of `#0F0F0F` or the primary color to indicate focus or elevation.
4. **Dividers:** Use 1px solid lines in `#EDEDED` for subtle separation, or 4px solid `#0F0F0F` for high-impact structural breaks.

## Shapes
The shape language is "Hyper-Rounded," creating a friendly yet high-tech tension against the aggressive typography.

- **Base Radius:** 20px (set via the `rounded-lg` equivalent of 1rem/1.25rem logic).
- **Buttons & Inputs:** Must strictly follow the 20px rounding to maintain the signature look.
- **Containers:** Large layout containers (cards) should use 24px or 32px rounding to emphasize the "organic-tech" blend.

## Components
- **Buttons:** 
    - *Primary:* Electric Blue background, White text. 20px radius. Heavy 600 weight.
    - *CTA:* Neon Lime background, Black text. Use for "Buy," "Sign Up," or "Get Started."
    - *Ghost:* 2px Black border, no fill, Black text.
- **Inputs:** Pure white fill with a 1px border of `#D1D1D1`. On focus, the border thickens to 2px Electric Blue.
- **Cards:** Use the `#F5F5F5` background with 20px rounding. No shadows. If high emphasis is needed, use a 2px Black border.
- **Chips/Badges:** Use Geist (Label-caps) with high-contrast fills (Black background/White text) for status indicators.
- **Lists:** Separated by 1px horizontal lines (`#EDEDED`). Use large `body-lg` typography for list items to maintain the bold feel.
- **Interaction States:** Hovering over a button should produce a slight "lift" (shift -4px Y) or a solid black offset "drop shadow" (blocky, not soft) to mimic physical movement.