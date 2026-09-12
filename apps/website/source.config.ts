import { defineConfig, defineDocs } from "fumadocs-mdx/config"

/**
 * The docs live as MDX on disk, in the repository they document.
 *
 * A page about backing Nooks up should be reviewed in the same pull request as the
 * thing it describes, which is only possible when it is a file beside the code rather
 * than a row in somebody's CMS.
 */
export const docs = defineDocs({
  dir: "content/docs",
})

export default defineConfig()
