# Changelog

## [1.6.0](https://github.com/hoshomoh/nooks/compare/v1.5.0...v1.6.0) (2026-09-25)


### New

* **api:** bring back a deleted list, and end the window it waits in ([6cca004](https://github.com/hoshomoh/nooks/commit/6cca00488265ac36ea1b5ff3a4d8ad321578fad5))
* **api:** choose what becomes of a removed member's lists ([a58ff48](https://github.com/hoshomoh/nooks/commit/a58ff482c7cd34fb4b4f29eeaf2493fe9a6ddf09))
* **api:** let the signup setting decide who may ask to join ([e58dc70](https://github.com/hoshomoh/nooks/commit/e58dc702ffb5eead62abd4d85cb01d10a8f30539))
* **live:** cap how many streams one member may hold open ([2d6ec54](https://github.com/hoshomoh/nooks/commit/2d6ec5411a52d8e337f825c390a9acf1dc95c2a3))
* **live:** one person edits a note, everybody else reads it ([628d68f](https://github.com/hoshomoh/nooks/commit/628d68f8c948969964d767c88487da55764860af))
* **mcp:** say how much a field holds before it is asked ([f0da375](https://github.com/hoshomoh/nooks/commit/f0da37594ef138f6d4cde6d1f08bda6cfaa0a373))
* **mcp:** tell an assistant how the tools fit together ([1832b8c](https://github.com/hoshomoh/nooks/commit/1832b8c45596d415e24409adac2071a773d20ed9))
* **requests:** bound who may be waiting, and what the panel draws ([2ed5ee4](https://github.com/hoshomoh/nooks/commit/2ed5ee466cb232c81ec552bc235c3044cd39f2bf))
* **store:** keep what a removed member gave everybody else ([fed20a7](https://github.com/hoshomoh/nooks/commit/fed20a74270745698a2d164c053ceb07bf99545b))
* **store:** sweep requests an admin answered a month ago ([a3d4ac7](https://github.com/hoshomoh/nooks/commit/a3d4ac7f0cab9e28e53a0b71398c0457e2eeb06b))
* **web:** say how much room is left in a bounded field ([013df3b](https://github.com/hoshomoh/nooks/commit/013df3b29f30ba1cfeaab1b0579c18b0f6fabd14))
* **web:** show pending reset requests beside the join requests ([3070637](https://github.com/hoshomoh/nooks/commit/3070637147225cec73d4f38555f71f0d9776e10e))


### Fixed

* **api:** bound every field a member writes into ([188726e](https://github.com/hoshomoh/nooks/commit/188726e7ff1fed2fb3e5acaf8fe5ae1a0fff4f4c))
* **api:** keep the cause of a failure out of the answer ([797aad4](https://github.com/hoshomoh/nooks/commit/797aad459190c7d9e5e019768da53a2989d10c2b))
* **api:** let only one person complete first run ([7f14b56](https://github.com/hoshomoh/nooks/commit/7f14b569fd67533d6dedad514c59da16876648ed))
* **api:** remember a member with no name on the public list ([29fdc8f](https://github.com/hoshomoh/nooks/commit/29fdc8fe63465a3aedeea2bf7ad96fed41c90ac6))
* **api:** stop telling a visitor which build is running ([f5250b1](https://github.com/hoshomoh/nooks/commit/f5250b11fe2ee4966a731f2da740208294ccc24c))
* **api:** tell a caller a contended write is worth trying again ([856d7b1](https://github.com/hoshomoh/nooks/commit/856d7b144f31eddd047e94685aed6c7e0f3fc033))
* **auth:** name the session cookie __Host- behind tls ([fb1bc92](https://github.com/hoshomoh/nooks/commit/fb1bc9290fcfd448e80a873132ec6c50d196228a))
* **ci:** run the checks CI runs, and say what is skipped ([7ac4cac](https://github.com/hoshomoh/nooks/commit/7ac4cac262e38f20fe90fd4f2752d41b6571d0eb))
* **ci:** set the test timeout from what the packages cost ([4f41ca4](https://github.com/hoshomoh/nooks/commit/4f41ca4db8c1a066b2db21739f98432f4ebcdf1b))
* **copy:** write the name in lower case where anybody reads it ([f31ac1a](https://github.com/hoshomoh/nooks/commit/f31ac1ae62f8fcd7ef9323cb783eabc364443411))
* **item:** compare the note where it is written, not before ([4b60715](https://github.com/hoshomoh/nooks/commit/4b607159754855f7972335d6738830a4aec7cc30))
* **live:** end a stream on a list its member can no longer see ([f4c7660](https://github.com/hoshomoh/nooks/commit/f4c76601991a35741eb3694fe58238a0a867ef9a))
* **live:** keep a note with whoever opened it first ([bec8511](https://github.com/hoshomoh/nooks/commit/bec851185a92c831c69ab5f0403fff3506fa19de))
* **mcp:** name every status a refused one could have been ([7490e63](https://github.com/hoshomoh/nooks/commit/7490e63087e257f63c8a35ab5f1b1faf4bb3f7c1))
* **mcp:** write the name in lower case where an assistant reads it ([95418b0](https://github.com/hoshomoh/nooks/commit/95418b02d6e3a637c34653340e9f81f5395763e8))
* **note:** keep punctuation somebody typed out of the markup ([b976598](https://github.com/hoshomoh/nooks/commit/b976598f1fd38709d6176e9e8f2aaa75f168f055))
* **proto:** write the name in lower case in the api comments ([8082d48](https://github.com/hoshomoh/nooks/commit/8082d487c965c2fcf8a3eca6b21ad5ae6ce3563a))
* **requests:** stop an approval to join standing open for ever ([f091c3d](https://github.com/hoshomoh/nooks/commit/f091c3d303a067cd2fdd716453d01346809a4f85))
* **requests:** tell the admin when somebody asks a second time ([7a54d69](https://github.com/hoshomoh/nooks/commit/7a54d6935f2633fce833de0ac3ae0ef674cd2386))
* **server:** refuse to be framed by another site ([2f3cd55](https://github.com/hoshomoh/nooks/commit/2f3cd5526c8b7bb7648842a147888a4e6395512a))
* **server:** write the backup snapshot beside the database ([9b97c23](https://github.com/hoshomoh/nooks/commit/9b97c23b5568ffb2dcd0403279f0050d002b0d57))
* **store:** bound how many connections postgres is asked for ([e8697aa](https://github.com/hoshomoh/nooks/commit/e8697aacf735b928e0082a691b5500f2fe2642d6))
* **store:** give every drawn list a settled order ([e2aa2bb](https://github.com/hoshomoh/nooks/commit/e2aa2bb023608f76d4706099b110f1f6161e13d4))
* **store:** give every item its own place on postgres ([78f6d51](https://github.com/hoshomoh/nooks/commit/78f6d51276292269246bef7eace6954af4e54fe9))
* **store:** give two items added at once different positions ([7f7f2f4](https://github.com/hoshomoh/nooks/commit/7f7f2f404905edcf8d117a4c27ff7704e1e47f03))
* **store:** hold the request queues one at a time on postgres ([dd6b13d](https://github.com/hoshomoh/nooks/commit/dd6b13d7512bae73f100fab783302791a8d9c60a))
* **store:** keep a list's counts true when people write at once ([b01b888](https://github.com/hoshomoh/nooks/commit/b01b888302fd7aa59e6f7031f8855fb9bab7da69))
* **store:** move a list's updated time when something on it changes ([186bf9d](https://github.com/hoshomoh/nooks/commit/186bf9dbd801b4006460393b385d4ffcb74ca409))
* **test:** stop the suite paying for bcrypt ([11d0112](https://github.com/hoshomoh/nooks/commit/11d0112a88ae0c7dff103c99a5ef835a43fc0f9a))
* **web:** give an icon button the row's hit area ([0f33a53](https://github.com/hoshomoh/nooks/commit/0f33a53305c98053ea78f623cb0a3fe937d42cfa))
* **web:** let a session change end a tick that is not its own ([0583dc5](https://github.com/hoshomoh/nooks/commit/0583dc5a64e5ed918138461f021fe3eb487ddde5))
* **web:** let asking again forget what was asked before ([8c621ab](https://github.com/hoshomoh/nooks/commit/8c621ab148f234dad3e8f02b81b7d7d6b735c2a7))
* **web:** write calendar months the way the design writes them ([78db8c3](https://github.com/hoshomoh/nooks/commit/78db8c390731c06f2a086f966e89522ceed0404a))
* write the name in lower case in the code, and hold it there ([83645d2](https://github.com/hoshomoh/nooks/commit/83645d2573edb85809509c7d99110df42b7090d2))


### Changed

* **api:** remove the conflict activity kind nothing could send ([1fd64cf](https://github.com/hoshomoh/nooks/commit/1fd64cf0c585fac845977e5b7e0e6b9f92c69139))
* **api:** state the field limits as constants rather than derive them ([99c4c9f](https://github.com/hoshomoh/nooks/commit/99c4c9f078091a771e249c5204039b706a4293c9))
* **design:** one token for how tall a floating list grows ([8444bd2](https://github.com/hoshomoh/nooks/commit/8444bd27eb7c4e30e51d4853ed6ab0d87f095c21))
* **proto:** declare the field limits where everything can read them ([8f8f8e4](https://github.com/hoshomoh/nooks/commit/8f8f8e47f1fd0feb6a4ffd0729ec3b9a79ccebb1))
* **public:** build the published page again only when it changed ([1122edd](https://github.com/hoshomoh/nooks/commit/1122eddf78fd9487bbb58b5bdcce837c57e43f95))
* **store:** keep the interface to what the app is promised ([8d631e4](https://github.com/hoshomoh/nooks/commit/8d631e40f3eb6b9b7313e9aef748fc1e66c3429f))
* **store:** remove the three functions nothing calls ([ae32b1e](https://github.com/hoshomoh/nooks/commit/ae32b1eb8fedff1096ddc14d9a61cd39b63cc56c))
* **test:** migrate one database and copy it ([e396087](https://github.com/hoshomoh/nooks/commit/e396087f7a3b5039259da38c99b15474041a1f26))
* **web:** draw the avatar chip in one place ([9c1e39e](https://github.com/hoshomoh/nooks/commit/9c1e39ef4fe0c6d5142240e29d67f183cdc1cd02))
* **website:** name the types in a signature ([c25790a](https://github.com/hoshomoh/nooks/commit/c25790a42ceea41380cccf365b5dcc6967f58ec3))

## [1.5.0](https://github.com/hoshomoh/nooks/compare/v1.4.1...v1.5.0) (2026-09-22)


### New

* **website:** show five lines of a release, the rest on github ([ea55bcd](https://github.com/hoshomoh/nooks/commit/ea55bcd247222cf63317d35b5531e3bb6a44ac14))


### Fixed

* **api:** record what failed, which nothing did ([7e77951](https://github.com/hoshomoh/nooks/commit/7e779519bbfbe7ba45fabf57357fc577e58e7b25))
* **api:** refresh everyone left when a member is removed ([e07e4e1](https://github.com/hoshomoh/nooks/commit/e07e4e10d042cce9ce756242aa911b03791e5c52))
* **api:** refuse a change from a token cut to read ([e7644bf](https://github.com/hoshomoh/nooks/commit/e7644bfd00d4901758610e2e7405405e82314ceb))
* **api:** stop a token learning that an item it cannot reach exists ([890a23c](https://github.com/hoshomoh/nooks/commit/890a23cc51894db4dfa6e73fe5d3f00bec11650b))
* **api:** stop telling a token about lists it cannot reach ([df068ac](https://github.com/hoshomoh/nooks/commit/df068ac71d75ab2021adc695ab94d4e5e0546707))
* **api:** tell whoever a group moved that their lists changed ([7296231](https://github.com/hoshomoh/nooks/commit/7296231b86eec7bd3ffd50ab37f2f3415daa362e))
* **api:** tell whoever a list was taken from that it went ([94c80d5](https://github.com/hoshomoh/nooks/commit/94c80d50b0315c52970c98f745fb4b3fde12c5c9))
* **auth:** sign other browsers out when a password changes ([5c95d3e](https://github.com/hoshomoh/nooks/commit/5c95d3e0ac301287db3bcaea2213e75e09687dfd))
* **backup:** refuse an access token the whole database ([2d0df78](https://github.com/hoshomoh/nooks/commit/2d0df78ff57efbd078d0b1d04961a63d99085b7a))
* **frontend:** serve the app under a content security policy ([a4eb5a7](https://github.com/hoshomoh/nooks/commit/a4eb5a742345c207e6a0139d2a733227dda90e65))
* **gateway:** carry the request headers the services read ([aef313d](https://github.com/hoshomoh/nooks/commit/aef313d4e271fca88f2ef5941d1dffeb1b4451b4))
* **live:** end a stream when its session does ([cb5da24](https://github.com/hoshomoh/nooks/commit/cb5da242f06e7fa3832c41a30e090a1814b0143d))
* **mcp:** say what the API said ([94bd522](https://github.com/hoshomoh/nooks/commit/94bd522d852de30b66e946f47b222bbf6d8f9132))
* **scripts:** stop claiming ci.sh runs everything CI runs ([f34606c](https://github.com/hoshomoh/nooks/commit/f34606c2f3b2b91fd1b609ef3c1fec5ddc14273e))
* **server:** bound a connection that is not sending anything ([7f29227](https://github.com/hoshomoh/nooks/commit/7f292270b440c6577560897f2bb49152dc8329d3))
* **server:** cap how much one request may send ([352c93b](https://github.com/hoshomoh/nooks/commit/352c93b388067be75811a4e20d4381df4a8e0a2f))
* **server:** send nosniff and a referrer policy ([92a2a1a](https://github.com/hoshomoh/nooks/commit/92a2a1adb8ebbaf011a872bc165b93435c68f033))
* **server:** sweep expired sessions, which nothing was doing ([dedac87](https://github.com/hoshomoh/nooks/commit/dedac877b403dc0300557fefa64fbb68ab9ea4e2))
* **store:** cap how many words a search uses ([6eee763](https://github.com/hoshomoh/nooks/commit/6eee7636601e156c13ab8eab40f00c1e5b3aebff))
* **store:** keep an unanswered request where an admin can see it ([f57de18](https://github.com/hoshomoh/nooks/commit/f57de18c519d8aed33332ecc614dfc715ea5d0b1))
* **store:** recount the lists a removed member added to ([ca879d1](https://github.com/hoshomoh/nooks/commit/ca879d1116636b700eb3612cf986c306f90b3807))
* **store:** refuse a batch of items spanning two lists ([48a205d](https://github.com/hoshomoh/nooks/commit/48a205d993a7feb7d8ecfada1c32713e5daff01b))
* **store:** settle the order of items that share a position ([b670284](https://github.com/hoshomoh/nooks/commit/b670284684ccd4b7059ba2212df247c796553d8a))
* **store:** sweep activity nothing can read ([ec1e2c2](https://github.com/hoshomoh/nooks/commit/ec1e2c2e753add4aa3cb9831f6c90b9438400262))
* **web:** clear the stored cache when a session ends ([eac8a02](https://github.com/hoshomoh/nooks/commit/eac8a02c2c45b0ef1fb66bd25eb206ada71518e1))
* **web:** filter the source scans by path, not by the glob's callback ([12aaf4b](https://github.com/hoshomoh/nooks/commit/12aaf4b31ad18181ecfafff7a3e7297535c81c54))
* **web:** put back the line that was edited out of a generated component ([fe8a7af](https://github.com/hoshomoh/nooks/commit/fe8a7afcb20a2b65b51fd13c536a668899a65a4f))
* **web:** reach the completed items the line counts ([037c004](https://github.com/hoshomoh/nooks/commit/037c004a2cc43408e8439390bf2f1f39194aff41))
* **web:** refresh the lists that go when a member is removed ([7229f0b](https://github.com/hoshomoh/nooks/commit/7229f0bca960193d7cc1245eaf66a982270ac977))
* **web:** say what removing a member actually takes ([8e69724](https://github.com/hoshomoh/nooks/commit/8e697243058f25307952d17ce87f060827765c30))
* **web:** size the public page's quantity badge with its token ([dc1667f](https://github.com/hoshomoh/nooks/commit/dc1667f29cb8b1a629df77b63c39302dda7bf5f5))
* **web:** stop crediting a tick to whoever added the item ([3a671cd](https://github.com/hoshomoh/nooks/commit/3a671cdcebaf8abe9872fb35feb1bca341ea4660))
* **web:** stop offering a tick on a list that is read-only ([300f069](https://github.com/hoshomoh/nooks/commit/300f069d322edd49554f459081f066a4f6dbc7a7))
* **web:** tell a member their role changed under them ([8e75b8c](https://github.com/hoshomoh/nooks/commit/8e75b8cc940b4c2b00bf6dbdcc870a73d57a4a07))


### Changed

* **api:** read one list to authorise an item, not all of them ([745e28a](https://github.com/hoshomoh/nooks/commit/745e28a0dd1916686435ff444643af69de7a9fdc))
* **events:** gather presence once, not once per watcher ([f2ede4e](https://github.com/hoshomoh/nooks/commit/f2ede4e8f101eb9e74d74a8d99dd6c1649ca5845))
* **live:** ask whether one list can be watched, not which ones can ([15950f9](https://github.com/hoshomoh/nooks/commit/15950f94136e0b09195d0cde2181c7e8a8669a18))
* **mcp:** drop a variable that called nothing to document what tools.go documents ([83f0c29](https://github.com/hoshomoh/nooks/commit/83f0c29c30f6b4a7c5f98c661216efb019a88a7e))
* **store:** count a list's items without opening its rows ([dd3598d](https://github.com/hoshomoh/nooks/commit/dd3598d45dbecb33807e85dfdb90fa1c38843013))
* **web:** name the event kinds the browser knows ([a107b42](https://github.com/hoshomoh/nooks/commit/a107b42a390fcc6af2b5ee9d21a02631767ee9a1))
* **web:** name the two types written into a signature ([73d14da](https://github.com/hoshomoh/nooks/commit/73d14da990a0d004b4a8c958e9b06396c9bde6b0))
* **web:** remove five exports nothing calls ([9f528f7](https://github.com/hoshomoh/nooks/commit/9f528f7163adda4485d505b751cbeb08353fbc59))
* **web:** write initialsOf once ([5a5ed04](https://github.com/hoshomoh/nooks/commit/5a5ed043a78478600e502a7b310f20bbf6f54b40))

## [1.4.1](https://github.com/hoshomoh/nooks/compare/v1.4.0...v1.4.1) (2026-09-16)


### Fixed

* **web:** make a menu answer the keys it prints ([8f7114a](https://github.com/hoshomoh/nooks/commit/8f7114a5775248b57d8db345153cd8b5429f945f))
* **web:** say what the date column is, and stop finished lists losing their menu ([3a2ff3e](https://github.com/hoshomoh/nooks/commit/3a2ff3e061e2d815760369b545fcb4324602857c))

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

* rename the app from Nook to nooks

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
* rename the app from Nook to nooks ([57bbbcf](https://github.com/hoshomoh/nooks/commit/57bbbcf229bf13dde3a5620ae14797f36d6d3494))
* **store:** move to bun for typed multi-dialect queries ([27853bd](https://github.com/hoshomoh/nooks/commit/27853bd4dec34d50412f56dcf438b0ab5683f649))
* **web:** make theme an external store and add the ds layer ([e57a7d9](https://github.com/hoshomoh/nooks/commit/e57a7d9b28b73032c70e7aa80a2f1a724fb589f1))
* **web:** stop every screen paying for the editor ([6ee7b0f](https://github.com/hoshomoh/nooks/commit/6ee7b0f40b7cf7cbc54b3ade782a36af48f2c389))
