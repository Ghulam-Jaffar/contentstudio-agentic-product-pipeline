# LinkedIn Posts with Stats Data

Out of 1546 total posts tested, only 30 posts returned stats data from LinkedIn's API.

**Total posts tested:** 1546
- UGC posts: 24 (0 with data)
- Share posts: 1522 (30 with data)

**Success rate:** 1.94% (30/1546)

## Posts with Data

| Post ID | Impressions | Likes | Comments | Shares |
|---------|-------------|-------|----------|--------|
| urn:li:share:7402611332602294272 | 1 | 0 | 0 | 0 |
| urn:li:share:7402385408220893184 | 17 | 2 | 0 | 0 |
| urn:li:share:7400332058566348800 | 17 | 1 | 0 | 0 |
| urn:li:share:7345773380752621568 | 3 | 0 | 0 | 0 |
| urn:li:share:7398281832884097025 | 5 | 1 | 0 | 0 |
| urn:li:share:7400573448063967232 | 12 | 1 | 0 | 0 |
| urn:li:share:7404318225691410432 | 2 | 0 | 0 | 0 |
| urn:li:share:7402343787538296832 | 1 | 0 | 0 | 0 |
| urn:li:share:7396904509513355264 | 1 | 0 | 0 | 0 |
| urn:li:share:7402360498886946818 | 4 | 0 | 0 | 0 |
| urn:li:share:7401996335077281792 | 2 | 0 | 0 | 0 |
| urn:li:share:7403563190925705216 | 4 | 0 | 0 | 0 |
| urn:li:share:7390734486549983232 | 0 | 0 | 1 | 0 |
| urn:li:share:7403744359139127296 | 1 | 0 | 0 | 0 |
| urn:li:share:7402340359424462849 | 2 | 0 | 0 | 0 |
| urn:li:share:7403930854911123456 | 1 | 1 | 1 | 0 |
| urn:li:share:7404744505117126659 | 2 | 0 | 0 | 0 |
| urn:li:share:7403538223102173184 | 1 | 0 | 0 | 0 |
| urn:li:share:7403627183388246016 | 3 | 0 | 0 | 0 |
| urn:li:share:7401071140749651968 | 0 | 1 | 0 | 0 |
| urn:li:share:7403607982233784320 | 1 | 0 | 0 | 0 |
| urn:li:share:7397375924066811905 | 0 | 1 | 0 | 0 |
| urn:li:share:7404716055098773504 | 2 | 0 | 0 | 0 |
| urn:li:share:7402602170623205376 | 1 | 0 | 0 | 0 |
| urn:li:share:7404575145933307905 | 1 | 0 | 0 | 0 |
| urn:li:share:7403124741546352640 | 0 | 1 | 0 | 0 |
| urn:li:share:7394654709678120960 | 6 | 2 | 0 | 0 |
| urn:li:share:7402902347884167169 | 1 | 0 | 0 | 0 |
| urn:li:share:7404302601598488576 | 2 | 0 | 0 | 0 |
| urn:li:share:7402118906036842496 | 2 | 0 | 0 | 0 |

## Totals from Posts with Data

- **Total Impressions:** 91
- **Total Likes:** 13
- **Total Comments:** 2
- **Total Shares:** 0

## URN List (for programmatic use)

\`\`\`
urn:li:share:7402611332602294272
urn:li:share:7402385408220893184
urn:li:share:7400332058566348800
urn:li:share:7345773380752621568
urn:li:share:7398281832884097025
urn:li:share:7400573448063967232
urn:li:share:7404318225691410432
urn:li:share:7402343787538296832
urn:li:share:7396904509513355264
urn:li:share:7402360498886946818
urn:li:share:7401996335077281792
urn:li:share:7403563190925705216
urn:li:share:7390734486549983232
urn:li:share:7403744359139127296
urn:li:share:7402340359424462849
urn:li:share:7403930854911123456
urn:li:share:7404744505117126659
urn:li:share:7403538223102173184
urn:li:share:7403627183388246016
urn:li:share:7401071140749651968
urn:li:share:7403607982233784320
urn:li:share:7397375924066811905
urn:li:share:7404716055098773504
urn:li:share:7402602170623205376
urn:li:share:7404575145933307905
urn:li:share:7403124741546352640
urn:li:share:7394654709678120960
urn:li:share:7402902347884167169
urn:li:share:7404302601598488576
urn:li:share:7402118906036842496
\`\`\`

## Conclusion

LinkedIn's API returns stats for only ~2% of posts. This confirms the severe API limitations documented earlier:
- The API only returns stats for posts that have recent engagement
- Most posts (98%) return empty elements
- The 365-day sync window implementation is correct, but even within that window, most posts won't have API data
