# Changelog

## [1.4.0](https://github.com/hoshomoh/nooks/compare/v1.3.0...v1.4.0) (2026-09-16)


### New

* **api:** say how much of a list is done, and when it changed ([548877a](https://github.com/hoshomoh/nooks/commit/548877a7c383cc496f915ed254193e7ceb359933))
* archive a list without losing it ([c42631c](https://github.com/hoshomoh/nooks/commit/c42631cd5a6cd7a18c052668e37d70ed036c6f48))
* **web:** give finished lists a group of their own ([e595e47](https://github.com/hoshomoh/nooks/commit/e595e47979fb46a746e14e457ac2659755d7e36b))
* **web:** keep a record of what was ticked, and when ([7fe09fd](https://github.com/hoshomoh/nooks/commit/7fe09fd01319a8f8726032f45c5a0fab73350638))
* **web:** make All lists somewhere you can look things up ([fbc056d](https://github.com/hoshomoh/nooks/commit/fbc056dd7d24db6885a454c6392ddc79a277ba4c))
* **web:** mark a note on the row instead of quoting it ([e8530f8](https://github.com/hoshomoh/nooks/commit/e8530f8e31d2a0e4cd13beb50b3d199932fa5781))


### Fixed

* **mcp:** tell an assistant when there are more lists than it was shown ([1a57ab3](https://github.com/hoshomoh/nooks/commit/1a57ab3d34beb33e6cd4f0e1d9aa3182042adfe5))
* **store:** order names the way a person reads them ([16b8fb8](https://github.com/hoshomoh/nooks/commit/16b8fb8d76477f4c4a5422d316c5a8743538e7db))
* **web:** leave room under what scrolls, and put the count at the end ([a7ac9e8](https://github.com/hoshomoh/nooks/commit/a7ac9e8c04c476d105076ad52b346b1f5eef68e9))
* **web:** let Escape stop at the control that took it ([84ff998](https://github.com/hoshomoh/nooks/commit/84ff99885f557d70c6418178f43a3e04f97dd0d3))


### Changed

* **ci:** let preflight skip groups a commit cannot affect ([925aaaa](https://github.com/hoshomoh/nooks/commit/925aaaaac7377ed2144a545355d2342708bf7c8c))
* read Lists a page at a time ([d7e4e2a](https://github.com/hoshomoh/nooks/commit/d7e4e2ab28b4b720a4f5c7671b88d58f56bbd62d))
* **store:** keep a List's counts on the List, and index what a page is ordered by ([b2578d0](https://github.com/hoshomoh/nooks/commit/b2578d0f1f285cb1b06bd8816612ea10ea729941))
* **web:** let the typecheck reuse what it already knows ([4ae67f4](https://github.com/hoshomoh/nooks/commit/4ae67f4436498f1b31501e150cad6885e4199779))

## [1.3.0](https://github.com/hoshomoh/nooks/compare/v1.2.0...v1.3.0) (2026-09-15)


### New

* **web:** move the list with the sheet, and answer a tick ([4d6ddd1](https://github.com/hoshomoh/nooks/commit/4d6ddd1a88db27884e6d05c1eb18e0a510d888ff))


### Fixed

* **web:** space the assistants page into sections ([8284e10](https://github.com/hoshomoh/nooks/commit/8284e109b5dd62c34c460dfe7b44736d8cc90077))

## [1.2.0](https://github.com/hoshomoh/nooks/compare/v1.1.1...v1.2.0) (2026-09-15)


### New

* **web:** keep the app on the machine that opened it ([77b85ee](https://github.com/hoshomoh/nooks/commit/77b85ee7fd992e028ad5a73f067924ee1a1dbe17))
* **web:** set an assistant up from settings ([c8e1ddf](https://github.com/hoshomoh/nooks/commit/c8e1ddf750fe02fd9203d8e64cce12633487df4a))


### Fixed

* one search result per item, and one identity per row ([1e31473](https://github.com/hoshomoh/nooks/commit/1e31473f0bda713a4b9fe87f4960d2b4582d64e0))
* **web:** keep the list clear of the side sheet ([d9207df](https://github.com/hoshomoh/nooks/commit/d9207df6cd580ae06d2f2d8e06d87f618b0d3f4c))


### Changed

* one source for the assistant setup blocks ([42f55a1](https://github.com/hoshomoh/nooks/commit/42f55a17389ca2a4d7b6726225214adb864f26d2))
* **web:** reuse test workers between files ([7939b4a](https://github.com/hoshomoh/nooks/commit/7939b4a70aeabf4c8d6174274d74840a1fd43a12))
* **web:** two test workers, not one per core ([5957b10](https://github.com/hoshomoh/nooks/commit/5957b10c5726c88514200d9daf759c2afaaec986))

## [1.1.1](https://github.com/hoshomoh/nooks/compare/v1.1.0...v1.1.1) (2026-09-14)


### Fixed

* **web:** print in one column, however long the list is ([75b8289](https://github.com/hoshomoh/nooks/commit/75b82894f22ba538a2391da06367523318ab074a))
* **web:** read a blank line as a separator, not an empty paragraph ([6e87517](https://github.com/hoshomoh/nooks/commit/6e87517b15ae263314d8778d20984e732f9d077e))
* **web:** scroll the content, not the window ([e6d7ee0](https://github.com/hoshomoh/nooks/commit/e6d7ee0e1e79cf280b821bfdca44bdcec9106d8e))
* **web:** stop the printed sheet calling every list a shopping list ([02ef127](https://github.com/hoshomoh/nooks/commit/02ef1272cc99897a2155e571bf1cd91254c98f70))
* **web:** truncate a long list name instead of squeezing the controls ([6fe075b](https://github.com/hoshomoh/nooks/commit/6fe075b5a0be2c820d6bcc88ae6267cd792f1d4b))
* **web:** wrap a long word instead of widening the note ([07b965d](https://github.com/hoshomoh/nooks/commit/07b965d23a07004b7c76ca7a8cab4f21403fc7fb))

## [1.1.0](https://github.com/hoshomoh/nooks/compare/v1.0.0...v1.1.0) (2026-09-14)


### New

* **web:** add dividers and tables to notes ([f73db88](https://github.com/hoshomoh/nooks/commit/f73db88142467e4be2069dcd641db2d819f59b0a))


### Fixed

* **build:** keep the embed marker that the app build deletes ([f179ec6](https://github.com/hoshomoh/nooks/commit/f179ec6fa5edb508e402feab9ff4cf28fe1a687b))
* **ci:** stop preflight tripping over its own build output ([0094463](https://github.com/hoshomoh/nooks/commit/0094463ddd25eec05d7ec33f4ee83e5a3a305d1d))
* make a clone runnable, and say where the instance is ([fbb194b](https://github.com/hoshomoh/nooks/commit/fbb194bf2585c4cf430a9cc329399475c14a6063))
* **web:** hold the sidebars still, and let a long name be read ([92a612a](https://github.com/hoshomoh/nooks/commit/92a612aab5f68d000d6bccafc06fabaacc440c1a))
* **web:** line up a search result with the list it came from ([df59cad](https://github.com/hoshomoh/nooks/commit/df59cad45ff2306e6df240665e0c0747f96b0428))


### Changed

* **web:** save a note the same way from both screens ([ea3252a](https://github.com/hoshomoh/nooks/commit/ea3252aa1f662f49b03e0d16922905a676e2cb42))

## 1.0.0 (2026-09-13)


### ⚠ BREAKING CHANGES

* rename the app from Nook to Nooks

### New

* **api:** add named sharing, groups and the activity panel ([9804b57](https://github.com/hoshomoh/nooks/commit/9804b57a46fdc0a18f7c671f4e2e3beee51f2bc0))
* **api:** cut and revoke access tokens ([e10117b](https://github.com/hoshomoh/nooks/commit/e10117b911c5563703dbc2da2ca70001ba4e6746))
* **api:** export the instance as its own database file ([2e20c2f](https://github.com/hoshomoh/nooks/commit/2e20c2f4ff24f845a062ca09b15899da842c4e64))
* **api:** give a token three abilities instead of two levels ([71b764c](https://github.com/hoshomoh/nooks/commit/71b764ca496fc4931e95ce268dfe17291167d24b))
* **api:** narrow a caller's access to what their token names ([c221e22](https://github.com/hoshomoh/nooks/commit/c221e22390eecea0e520d39e3f9bbfc3342fd281))
* **api:** publish one list to visitors without an account ([67cb187](https://github.com/hoshomoh/nooks/commit/67cb187b325866ae165a5e74dd95edf9d0730522))
* **api:** record and show what a token did ([1b44bc4](https://github.com/hoshomoh/nooks/commit/1b44bc47215513a5c88a24cf8be94d6f21f8207b))
* **api:** refuse a text change that would overwrite another ([1ee5fdd](https://github.com/hoshomoh/nooks/commit/1ee5fdd8924aeb0517f0cee01aefd6b0d49755d2))
* **api:** serve the mcp tools from the same services as the app ([747665d](https://github.com/hoshomoh/nooks/commit/747665d0dd7859a9cb9a56a3e4f76bf21f917737))
* **api:** serve the rest api from the same services as the app ([0aecde0](https://github.com/hoshomoh/nooks/commit/0aecde01b305acc429f1d2b14d7dd7984f99ebcf))
* **api:** split the session into a refresh cookie and an access token ([556b0d7](https://github.com/hoshomoh/nooks/commit/556b0d757a5bf03b40eacb9be1d47ed01a03c944))
* **auth:** add first run, sign in, sign out and password replacement ([f99a2f1](https://github.com/hoshomoh/nooks/commit/f99a2f1980f466d66fbef1b949af574ae65a4810))
* **auth:** add join and password reset request flows ([7f9986f](https://github.com/hoshomoh/nooks/commit/7f9986f7d4f699e445f4b43d6420b4c8c6f40914))
* choose which list an instance publishes ([bacf8e2](https://github.com/hoshomoh/nooks/commit/bacf8e287cce4916153e03304c22dea17b6b440a))
* **list:** add list and item rpcs with search ([a16f6ac](https://github.com/hoshomoh/nooks/commit/a16f6ac089da654950cb5c804d1fd699e5f2a6d5))
* **mcp:** reach everything an access token reaches ([a1e491c](https://github.com/hoshomoh/nooks/commit/a1e491c469f929cc5510fcf262c26f3445a783f4))
* **notes:** add the block editor with a slash menu ([268c93b](https://github.com/hoshomoh/nooks/commit/268c93ba1895359a57b3ce6611cd683b864e70b5))
* **notes:** add the full-screen note view ([d039ebf](https://github.com/hoshomoh/nooks/commit/d039ebf66803e89326bc6d59d50ae0917f2d9104))
* **notes:** store notes as markdown with a side sheet ([473a98d](https://github.com/hoshomoh/nooks/commit/473a98d417477e3613eef8b6283dd75f5ad6d031))
* **password:** add hashing and the twelve-character rule ([3fbe809](https://github.com/hoshomoh/nooks/commit/3fbe8091a0a578e607c2f7e622cd547ec56175db))
* **profile:** add instance configuration ([d2887ff](https://github.com/hoshomoh/nooks/commit/d2887ff2ef95ed978022c0602c35ab28bf812652))
* **requests:** add admin approval for join and reset requests ([94424c6](https://github.com/hoshomoh/nooks/commit/94424c6d45013d9814f422e0780cf7eab4031147))
* run nooks from one container ([91835d2](https://github.com/hoshomoh/nooks/commit/91835d21c3e646bc3917cc367db928d6a768e4ce))
* **search:** add full-text search over lists and items ([f66f1b4](https://github.com/hoshomoh/nooks/commit/f66f1b4cbe9432865ed7832e5ff37e1227236881))
* **server:** embed and serve the built app with spa fallback ([771d719](https://github.com/hoshomoh/nooks/commit/771d7193d252f403c32b5effd50d8aca3dc2b415))
* **server:** serve instance service over connect ([99b9a55](https://github.com/hoshomoh/nooks/commit/99b9a55a64dcc11f1fdca8b379685f1f1fab8756))
* **store:** add access tokens, scoped to named lists ([31d9fd7](https://github.com/hoshomoh/nooks/commit/31d9fd72eead6850a3707098ff30f6f096176004))
* **store:** add groups, named sharing and activity ([51a7828](https://github.com/hoshomoh/nooks/commit/51a78286d04b85e746a09ca99e1ce470bce5eedb))
* **store:** add join and reset requests ([39e18f1](https://github.com/hoshomoh/nooks/commit/39e18f12f851dd67118f29f06ec29d8f750387a4))
* **store:** add lists and items ([aa39eb0](https://github.com/hoshomoh/nooks/commit/aa39eb0ca642ae33acd69eafe98fb13f6aebf786))
* **store:** add members and sessions ([ee07746](https://github.com/hoshomoh/nooks/commit/ee0774667f9b089a66df15bff6082c8a882be290))
* **store:** add store interface with sqlite and postgres drivers ([d39da6c](https://github.com/hoshomoh/nooks/commit/d39da6c5122e65ddc6362c1590c2eaf7e79694c1))
* stream changes and presence to the browsers watching ([d80bbc8](https://github.com/hoshomoh/nooks/commit/d80bbc8677605a98459c2885d194a638de513c0a))
* **views:** add the calendar ([36b5eb2](https://github.com/hoshomoh/nooks/commit/36b5eb280f5122c5211107cca766a17faab9b58b))
* **views:** add today and upcoming ([354b8f0](https://github.com/hoshomoh/nooks/commit/354b8f08a5a74e155f1e1ae37f0c85e428e985ca))
* **web:** add items from Today and Upcoming with a list and date from context ([554decf](https://github.com/hoshomoh/nooks/commit/554decfa5cca894f8acb15b3e7ddac43ef190a62))
* **web:** add routing with first run, sign in and password replacement ([391bb19](https://github.com/hoshomoh/nooks/commit/391bb19a5b9b80051821a23c1ed4a4477bcc0678))
* **web:** add the appearance settings page ([3aa8c4d](https://github.com/hoshomoh/nooks/commit/3aa8c4d58717a8e9e749b9de88b49e4f1275bd95))
* **web:** add the groups page ([b8d6ff9](https://github.com/hoshomoh/nooks/commit/b8d6ff97b1771cca60737bf9943cd3ea7dcd22fd))
* **web:** add the instance settings page ([0e70582](https://github.com/hoshomoh/nooks/commit/0e70582117e28895a665930db0f54ae88f70c8f9))
* **web:** add the list menu with duplicate, export and delete ([d7ffbd7](https://github.com/hoshomoh/nooks/commit/d7ffbd7eedeedf07f0c7e1a9234228f84973eb40))
* **web:** add the members page ([acc4c42](https://github.com/hoshomoh/nooks/commit/acc4c4245dd9d3875c7a74ec8a1e96da218428f8))
* **web:** add the public list page ([bbdd3da](https://github.com/hoshomoh/nooks/commit/bbdd3da6766c6ef2560221abeecb247e5cd34715))
* **web:** add the share dialog and the activity panel ([8bb28f8](https://github.com/hoshomoh/nooks/commit/8bb28f8019c58ef6c860427a74d3ce09852eea4e))
* **web:** add the sidebar, list view and command palette ([36dc897](https://github.com/hoshomoh/nooks/commit/36dc897a8347a5c08de2a7b9ca16904ab738b79e))
* **web:** adopt the new nooks mark and initialise shadcn ([a21e85b](https://github.com/hoshomoh/nooks/commit/a21e85ba9d979679f3ba92878c0f70c47233bb34))
* **web:** answer the instance address with the public list ([e62010b](https://github.com/hoshomoh/nooks/commit/e62010bbd89e600a3f9eae596e8d45d0bf91db79))
* **web:** apply nooks tokens, theme and connect transport ([a58ff24](https://github.com/hoshomoh/nooks/commit/a58ff24119358dcad1fa77ce4c7e1a9f2d33d8f9))
* **web:** apply the tick a visitor reached for before signing in ([8869347](https://github.com/hoshomoh/nooks/commit/8869347709e82b5b229da29132b994c6e4586ca0))
* **web:** approve or ignore a request from the activity panel ([7a73b1e](https://github.com/hoshomoh/nooks/commit/7a73b1e25c33006909c79970231dda267c792428))
* **web:** ask when two people wrote different things ([9907c09](https://github.com/hoshomoh/nooks/commit/9907c0960a131ce32cb6f5a2831e9df2669b2af4))
* **web:** collapse completed items behind a done-today row ([346a902](https://github.com/hoshomoh/nooks/commit/346a9028a912eb507a8188fdc485e8a02ed86b48))
* **web:** cut and revoke access tokens from settings ([be8278b](https://github.com/hoshomoh/nooks/commit/be8278b2cb56caa61c1d2ed023a1116ec8515a0f))
* **web:** delete an instance from about ([a95c894](https://github.com/hoshomoh/nooks/commit/a95c894b95f1a8bbc2fc5e511eebec5146d3c08c))
* **web:** draw icons with the icon library ([9539916](https://github.com/hoshomoh/nooks/commit/9539916b369219a99b3ef8e22489a2ae1c2a7897))
* **web:** edit an item's quantity and date on its own row ([8c3d520](https://github.com/hoshomoh/nooks/commit/8c3d520204b1b5b81560bb5bb2fd75d0da2aaca5))
* **web:** hide markdown markup and use the design system's calendar ([d11b48d](https://github.com/hoshomoh/nooks/commit/d11b48de23c4591940c475f8df478d416378dd78))
* **web:** keep working when the instance is out of reach ([a8190af](https://github.com/hoshomoh/nooks/commit/a8190afc61b5484cfc358c06055d91875b1a856c))
* **web:** let a member change their own name, email and password ([f844c7b](https://github.com/hoshomoh/nooks/commit/f844c7bc12b621896957b788b9f78c38a0792cce))
* **web:** let a member format the text they wrote ([a1d80d5](https://github.com/hoshomoh/nooks/commit/a1d80d59473be6b777eaf42126106e5e051c8c1d))
* **web:** let a member sign out ([5346663](https://github.com/hoshomoh/nooks/commit/53466633e3f482d7cf4e2e1c69ae0506af305a6c))
* **web:** let a token reach every list, and say what it may do ([e59925e](https://github.com/hoshomoh/nooks/commit/e59925ec2df797daba65ad83bc4f31f0a2e4271a))
* **web:** move what pushes and what arrives unasked ([9d8bde5](https://github.com/hoshomoh/nooks/commit/9d8bde59c54843fa6f5604d27febfe1ab65c0302))
* **web:** name types, add date-fns and localisation ([5f3515d](https://github.com/hoshomoh/nooks/commit/5f3515d4dfe9ad75f659c4182c358856c6e41765))
* **web:** open onto a list worth reading ([d0f6db9](https://github.com/hoshomoh/nooks/commit/d0f6db98439267da4edb8e1d8f867067729409a7))
* **web:** open what a search found, not where it lives ([ed3e409](https://github.com/hoshomoh/nooks/commit/ed3e40947e8e81b2e77fdf020c45d750a9739cb4))
* **web:** parse quantities and dates from the add row ([5a4cf41](https://github.com/hoshomoh/nooks/commit/5a4cf419b687b85df60eb58b2363889a6080f4be))
* **web:** print long lists two-up ([adb3ab6](https://github.com/hoshomoh/nooks/commit/adb3ab63bf130b29f86eb31774c26beac4c3ac86))
* **web:** read how far away a day is, not just its name ([50fdb21](https://github.com/hoshomoh/nooks/commit/50fdb217ec0a5a42b33323932107244d03ff275a))
* **web:** rebuild the list menus and seed a first list ([71118a8](https://github.com/hoshomoh/nooks/commit/71118a89fac40708626381254114356ff079fd58))
* **web:** rebuild the note editor on tiptap ([b13ab29](https://github.com/hoshomoh/nooks/commit/b13ab2946a70c878535715612ea912991643074d))
* **web:** say when the instance cannot be reached ([335e333](https://github.com/hoshomoh/nooks/commit/335e333462a5e18c7f347e2961d6ff25f1d36fb6))
* **web:** send what was queued when the tab reopens ([fa6ef6a](https://github.com/hoshomoh/nooks/commit/fa6ef6adb6471209e07f6f359186772701b78da3))
* **web:** set an item's date and quantity inside its menu ([139ad18](https://github.com/hoshomoh/nooks/commit/139ad18dce717406ffaa160186942016676ee0b7))
* **website:** document installing, configuring and backing up ([d846914](https://github.com/hoshomoh/nooks/commit/d84691483ccaeb4b7e2aba93eceeb8c7340e0b7b))
* **website:** follow the reader's theme ([fe8d06f](https://github.com/hoshomoh/nooks/commit/fe8d06f828fde4b251d0eb5cb6023f1c777c1eaa))
* **website:** generate the api reference from the spec ([3423328](https://github.com/hoshomoh/nooks/commit/3423328ba71c62f191e2c6ea1623e59e38cf8950))
* **website:** give the changelog the page the design drew ([2d84f24](https://github.com/hoshomoh/nooks/commit/2d84f24688babbd89cce18d3a780fe87df14ece4))
* **website:** make the reference read like documentation ([c137d09](https://github.com/hoshomoh/nooks/commit/c137d09790cd732cf29c5a905abd8c08bafe71aa))
* **website:** rebuild the site to the page design ([8186252](https://github.com/hoshomoh/nooks/commit/81862528d9afd27c4d1fe460c40a4c05562f37c8))
* **website:** scaffold the site on the app's own tokens ([4ddb228](https://github.com/hoshomoh/nooks/commit/4ddb2282fe5a1adc1bf72e156b8e78659d8e81c4))
* **website:** write the landing page ([c9781b2](https://github.com/hoshomoh/nooks/commit/c9781b27059c635d7b5a6bcb6d016867b2f98c70))


### Fixed

* **api:** refuse account-level changes made with a token ([b1b6dbc](https://github.com/hoshomoh/nooks/commit/b1b6dbcc6582eed41d4c4af57c4049e77d3e97d2))
* **build:** check the commit being pushed, not the worktree's own head ([87c5540](https://github.com/hoshomoh/nooks/commit/87c55408a7c3ff9d7f590dbb65eac9b9f24f52cd))
* **build:** let preflight reuse the checkout it made ([1103c36](https://github.com/hoshomoh/nooks/commit/1103c3664b2874009e7a9605be91974f4130389e))
* let the container write its own data directory ([a187e33](https://github.com/hoshomoh/nooks/commit/a187e333deae44846c70bd0dc49a1076b5079476))
* **web:** bring the visitor page up to the design ([248417f](https://github.com/hoshomoh/nooks/commit/248417f900bcd604ee4b1cfb53e81d27a8a0f86b))
* **web:** build settings as the design draws them ([cd116aa](https://github.com/hoshomoh/nooks/commit/cd116aa744ed54ae8b35285c3cfd6686a2e580dc))
* **web:** centre a note checkbox on the line beside it ([af07664](https://github.com/hoshomoh/nooks/commit/af07664cd94fb7f359fccea4cd5f19fb640c7742))
* **web:** converge on the instance when a change is refused ([a80acba](https://github.com/hoshomoh/nooks/commit/a80acba88b2c842894af392546afacd7a9190349))
* **web:** draw no rule under the add row on an empty dated view ([68cf120](https://github.com/hoshomoh/nooks/commit/68cf120269fc03a298c7e29f8d1b473190f5bbca))
* **web:** hide block shorthand always, size dialogs, and make ⌘K a command menu ([c11c46b](https://github.com/hoshomoh/nooks/commit/c11c46be05f30c112b8074e2ab116fc12ce791fe))
* **web:** keep rows clickable and print the list, not the page ([fa51789](https://github.com/hoshomoh/nooks/commit/fa51789034991680ba89ad9b9ed662d2a143fdbd))
* **web:** keep settings forms honest and make a group in one step ([371edaa](https://github.com/hoshomoh/nooks/commit/371edaaab5181cc9fb260327e91c363f5e237464))
* **web:** land on the right page after first run ([23a98c8](https://github.com/hoshomoh/nooks/commit/23a98c8713dfc221796cfaca206718d29e641227))
* **web:** make a row click open it, and say so with the pointer ([ea8347b](https://github.com/hoshomoh/nooks/commit/ea8347bcee94f85859dae802e7ef555e70eb123a))
* **web:** make screens read live data and repair the design tokens ([b17fb8b](https://github.com/hoshomoh/nooks/commit/b17fb8b0462dea878665991f98485c14dafccc92))
* **web:** make the done strike follow the theme ([47834e8](https://github.com/hoshomoh/nooks/commit/47834e8a1f97243e8499a09ebcaee4e95cfdf245))
* **web:** match the list view to the design ([0ad9f63](https://github.com/hoshomoh/nooks/commit/0ad9f6387d97ffb9d8c847bd2921c926e7f0a0fc))
* **web:** print the list instead of a blank page ([628a45b](https://github.com/hoshomoh/nooks/commit/628a45b89122b8216e2a58b33f2d69a22350663f))
* **web:** put the side sheet on its own layer ([1f8eab7](https://github.com/hoshomoh/nooks/commit/1f8eab76c578176a7d75e04facaf6ebbc599009e))
* **web:** repair navigation, the activity panel, the editor and the sheet ([223f185](https://github.com/hoshomoh/nooks/commit/223f18527debf3e8942d23cee68d4d48ececa455))
* **web:** show a note the way the note shows it ([ea445c4](https://github.com/hoshomoh/nooks/commit/ea445c4bab34afd36ecaa40bccf1ad49bee32a70))
* **web:** show a real error state when a screen cannot load ([43a7713](https://github.com/hoshomoh/nooks/commit/43a77136eff87dd08b75b274aa583581b70146fd))
* **web:** show the keyboard where it is ([942a7d5](https://github.com/hoshomoh/nooks/commit/942a7d5317aafc3082a94931a9c16879e49e7e37))
* **website:** give the api a section of its own ([cf03b6d](https://github.com/hoshomoh/nooks/commit/cf03b6ddc95cf216f8ab23f3811fb00b00b6d1e1))
* **website:** give the docs their copy buttons back ([be99ddb](https://github.com/hoshomoh/nooks/commit/be99ddbeac972179da4925dc10da6ae3b8cd4f8c))
* **website:** let the reference get back to the docs ([4d89aab](https://github.com/hoshomoh/nooks/commit/4d89aab56d54c33fd16b2cf336a85a5fef7ad83e))
* **website:** open the docs groups by default ([efcbb82](https://github.com/hoshomoh/nooks/commit/efcbb823238f31df565d8649af1305042b1f36fa))
* **website:** stop skipping deploys that changed the docs ([661b495](https://github.com/hoshomoh/nooks/commit/661b49546052165dd5298bdfba8e4aa0c8f748fb))
* **web:** stop the sign-in page opening with an error on it ([7298d21](https://github.com/hoshomoh/nooks/commit/7298d211e5a761320429c16edeb8b2b0e876995b))
* **web:** tick done items on the printed sheet ([8d10fd6](https://github.com/hoshomoh/nooks/commit/8d10fd6a1062a8d921ad50b844575d3db6e9c466))
* **web:** tokenise the type scale and unbreak the production build ([b4e7853](https://github.com/hoshomoh/nooks/commit/b4e7853f1c6a4110d16cd1e9947d830e4ec8572c))


### Changed

* **api:** stop deriving a note's first line ([ddd1e70](https://github.com/hoshomoh/nooks/commit/ddd1e7006abd0fb9df0f9a79f24deb961b841746))
* copy a list in one write, and parse its notes once ([f68eeb5](https://github.com/hoshomoh/nooks/commit/f68eeb5c179722ef99e1c84131038996706f32af))
* keep design values in one place ([473f826](https://github.com/hoshomoh/nooks/commit/473f8261c0a67583d898175e56c275ac50d1f4d3))
* move to an apps and packages monorepo layout ([5fa6a77](https://github.com/hoshomoh/nooks/commit/5fa6a7735073dbc838979b42fd7e9092924c7431))
* rename the app from Nook to Nooks ([57bbbcf](https://github.com/hoshomoh/nooks/commit/57bbbcf229bf13dde3a5620ae14797f36d6d3494))
* **store:** move to bun for typed multi-dialect queries ([27853bd](https://github.com/hoshomoh/nooks/commit/27853bd4dec34d50412f56dcf438b0ab5683f649))
* **web:** make theme an external store and add the ds layer ([e57a7d9](https://github.com/hoshomoh/nooks/commit/e57a7d9b28b73032c70e7aa80a2f1a724fb589f1))
* **web:** stop every screen paying for the editor ([6ee7b0f](https://github.com/hoshomoh/nooks/commit/6ee7b0f40b7cf7cbc54b3ade782a36af48f2c389))
