import { lazy, type ComponentType } from "react"

/**
 * Fetching a component the first time it is actually rendered.
 *
 * React.lazy wants a module with a default export, and nothing in Nooks has one — a
 * default export is a name the importer gets to choose, which is how two files end up
 * calling the same component two things. This adapts a named export instead, so the
 * split costs nothing at the call site and no component has to change shape to be
 * split.
 */
export function lazyNamed<Props>(
  load: () => Promise<Record<string, unknown>>,
  name: string,
): ComponentType<Props> {
  return lazy(async () => {
    const module = await load()
    return { default: module[name] as ComponentType<Props> }
  })
}
