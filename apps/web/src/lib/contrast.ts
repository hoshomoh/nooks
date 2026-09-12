/**
 * Reading a colour out of the palette and asking whether it can be read.
 *
 * The tokens are OKLCH because OKLCH is the right space to pick colours in — it is
 * perceptually even, so two colours a step apart look a step apart. It is the wrong
 * space to measure contrast in: WCAG is defined on sRGB relative luminance, and
 * Oklab's L is not that. So this converts, once, in one place, and the conversion is
 * here rather than in a test because a number nobody can point at is a number nobody
 * can argue with.
 */

/** One colour, as OKLCH declares it. */
export interface Oklch {
  /** Perceptual lightness, 0 to 1. */
  l: number
  /** Chroma. */
  c: number
  /** Hue in degrees. */
  h: number
}

/**
 * parseOklch reads `oklch(0.4836 0.0883 252.04)`, or nothing.
 *
 * Answers null rather than throwing for anything else: the stylesheet holds sizes and
 * references as well as colours, and a caller asking about a token that is not a colour
 * wants to be told so.
 */
export function parseOklch(value: string): Oklch | null {
  const found = /oklch\(\s*([\d.]+%?)\s+([\d.]+)\s+([\d.]+)/i.exec(value)
  if (!found) {
    return null
  }
  const [, lightness = "", chroma = "", hue = ""] = found
  return {
    l: lightness.endsWith("%") ? Number.parseFloat(lightness) / 100 : Number.parseFloat(lightness),
    c: Number.parseFloat(chroma),
    h: Number.parseFloat(hue),
  }
}

/** One colour as sRGB, each channel 0 to 1. */
export interface Rgb {
  r: number
  g: number
  b: number
}

/** toRgb converts OKLCH to sRGB, clamped to what a screen can show. */
export function toRgb({ l, c, h }: Oklch): Rgb {
  const radians = (h * Math.PI) / 180
  const a = c * Math.cos(radians)
  const b = c * Math.sin(radians)

  // Oklab to the cone responses it is defined against, then cubed back out of them.
  const lCone = (l + 0.3963377774 * a + 0.2158037573 * b) ** 3
  const mCone = (l - 0.1055613458 * a - 0.0638541728 * b) ** 3
  const sCone = (l - 0.0894841775 * a - 1.291485548 * b) ** 3

  return {
    r: encode(4.0767416621 * lCone - 3.3077115913 * mCone + 0.2309699292 * sCone),
    g: encode(-1.2684380046 * lCone + 2.6097574011 * mCone - 0.3413193965 * sCone),
    b: encode(-0.0041960863 * lCone - 0.7034186147 * mCone + 1.707614701 * sCone),
  }
}

/** encode applies the sRGB transfer function and clamps to the displayable range. */
function encode(channel: number): number {
  const clamped = Math.min(1, Math.max(0, channel))
  return clamped <= 0.0031308 ? clamped * 12.92 : 1.055 * clamped ** (1 / 2.4) - 0.055
}

/** luminance is WCAG relative luminance, which is defined on linear sRGB. */
export function luminance({ r, g, b }: Rgb): number {
  const linear = (channel: number) =>
    channel <= 0.04045 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4
  return 0.2126 * linear(r) + 0.7152 * linear(g) + 0.0722 * linear(b)
}

/**
 * contrastRatio is how far apart two colours are to read, 1 to 21.
 *
 * WCAG AA wants 4.5 for body text, 3 for large text and for the parts of a control that
 * say where it is.
 */
export function contrastRatio(a: Rgb, b: Rgb): number {
  const [brighter, darker] = [luminance(a), luminance(b)].sort((x, y) => y - x)
  return ((brighter ?? 0) + 0.05) / ((darker ?? 0) + 0.05)
}
