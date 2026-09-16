import { Children, Fragment, isValidElement, type ElementType, type ReactNode } from "react"

/**
 * A key a menu prints beside what it does.
 *
 * Read from the keycap itself rather than declared twice. A menu that prints "⌘D" and
 * binds something else is a menu that lies, and the two drift the first time one of
 * them is edited.
 */
export interface Keycap {
  /** ⌘ on a Mac and Ctrl everywhere else, which a Member means as the same key. */
  command: boolean
  shift: boolean
  /** What KeyboardEvent.key calls it, lowercased. */
  key: string
}

/** The keycaps that stand for a key with no letter: what is drawn, and what it is. */
const DRAWN: Record<string, string> = {
  "↵": "enter",
  "⌫": "backspace",
  "⎋": "escape",
}

/**
 * readKeycap turns a printed keycap into the key it stands for.
 *
 * Null for anything that is not a key at all. Menus print other things in the same
 * place — a tick against the order a table is in — and those are not to be pressed.
 */
export function readKeycap(drawn: string): Keycap | null {
  let rest = drawn
  const command = rest.startsWith("⌘")
  if (command) {
    rest = rest.slice(1)
  }
  const shift = rest.startsWith("⇧")
  if (shift) {
    rest = rest.slice(1)
  }

  if (DRAWN[rest]) {
    return { command, shift, key: DRAWN[rest] }
  }
  // A letter, or a function key. Anything else is a symbol that means something other
  // than "press this".
  if (/^([a-z]|f\d{1,2})$/i.test(rest)) {
    return { command, shift, key: rest.toLowerCase() }
  }
  return null
}

/** Whether a key press is the one a keycap named. */
export function pressed(cap: Keycap, event: KeyboardEvent): boolean {
  return (
    event.key.toLowerCase() === cap.key &&
    // Either modifier: the same keycap is ⌘ on a Mac and Ctrl on everything else.
    command(event) === cap.command &&
    event.shiftKey === cap.shift &&
    !event.altKey
  )
}

function command(event: KeyboardEvent): boolean {
  return event.metaKey || event.ctrlKey
}

/** One entry's keycap and what pressing it does. */
export interface BoundKey {
  cap: Keycap
  run: () => void
}

/** What an entry in a menu has to say for its key to be found. */
interface KeyedEntry {
  shortcut?: string
  disabled?: boolean
  onSelect: () => void
}

/*
keysIn reads the keys the entries of a menu print.

From the elements themselves rather than from a second list the caller keeps in step:
the keycap a Member reads and the key the menu answers are then one declaration, and
there is nothing to forget to update. Fragments are walked into, because that is how a
menu renders a section only some Lists have.
*/
export function keysIn(children: ReactNode, entry: ElementType): BoundKey[] {
  return Children.toArray(children).flatMap((child) => {
    if (!isValidElement(child)) {
      return []
    }
    if (child.type === Fragment) {
      return keysIn((child.props as { children?: ReactNode }).children, entry)
    }
    if (child.type !== entry) {
      return []
    }

    const { shortcut, disabled, onSelect } = child.props as KeyedEntry
    const cap = shortcut ? readKeycap(shortcut) : null
    return cap && !disabled ? [{ cap, run: onSelect }] : []
  })
}

/**
 * answers runs whatever the pressed key was the keycap for.
 *
 * On the container rather than on each entry, because only the highlighted one has
 * focus: a handler on the entries would answer ⌘D while Pin is highlighted and ignore
 * it the rest of the time, which is worse than not answering at all.
 */
export function answers(keys: BoundKey[], event: KeyboardEvent): boolean {
  const bound = keys.find(({ cap }) => pressed(cap, event))
  if (!bound) {
    return false
  }
  bound.run()
  return true
}
