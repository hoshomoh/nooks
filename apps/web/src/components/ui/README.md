# Generated shadcn/ui components

**Do not edit anything in this directory.**

These files are produced by `pnpm dlx shadcn@latest add <component>` and are overwritten
whenever a component is added or updated. Editing them means the next update either
clobbers your change or has to be merged by hand — and both go wrong quietly.

Nooks' own components live in `../ds/`. They compose these primitives and apply
`DESIGN.md`. If a component needs to look or behave differently:

1. Wrap it in `../ds/`, passing `className` or props.
2. If that cannot express it, change a token in `src/index.css` — which fixes every
   instance rather than one.
3. Only if neither works, discuss it before touching this directory.
