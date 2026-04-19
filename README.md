# musidl — Global Music Downloader

Search any artist across every country on earth and download **full songs** in one command. No API keys. No accounts.

```
  +══════════════════════════════════════════════════════════+
  |  ███╗   ███╗██╗   ██╗███████╗██╗██████╗ ██╗      |
  |  ████╗ ████║██║   ██║██╔════╝██║██╔══██╗██║      |
  |  ██╔████╔██║██║   ██║███████╗██║██║  ██║██║      |
  |  ██║╚██╔╝██║██║   ██║╚════██║██║██║  ██║██║      |
  |  ██║ ╚═╝ ██║╚██████╔╝███████║██║██████╔╝███████╗ |
  |  ╚═╝     ╚═╝ ╚═════╝ ╚══════╝╚═╝╚═════╝ ╚══════╝|
  |                                                            |
  |  Global Music Downloader   github.com/krainium       |
  +══════════════════════════════════════════════════════════+
```

---

## What it does

1. Takes an artist name (prompt or argument)
2. Queries **179 sources simultaneously** in parallel across iTunes, Deezer, and MusicBrainz
3. Deduplicates and sorts the merged catalog (newest first)
4. Displays a numbered track table — pick individual tracks, ranges, or all
5. Downloads **full MP3 tracks** in parallel via yt-dlp with multi-strategy search:
   - **Strategy 1:** YouTube `ytsearch1:ARTIST - TITLE`
   - **Strategy 2:** YouTube `ytsearch1:ARTIST TITLE`
   - **Strategy 3:** YouTube `ytsearch1:ARTIST TITLE official audio`
   - **Strategy 4:** SoundCloud `scsearch1:ARTIST TITLE`

---

## Requirements

**Go 1.21+** to build. Zero external Go dependencies.

**yt-dlp** is required for full-track downloads. musidl manages it for you entirely:

| Situation | What musidl does |
|-----------|-----------------|
| yt-dlp not installed | Installs it automatically via pip before the first download |
| pip not available | Tells you exactly what to install, then exits cleanly |
| yt-dlp is a week old | Checks PyPI in the background while you search; upgrades silently if a newer version exists; reports the result before downloads begin |
| A download fails | Checks for a newer yt-dlp unconditionally, upgrades if one exists, and retries the failed track automatically |
| Recovery upgrade succeeds | Prints a one-time notice with the new version so you know what fixed it |
| Many tracks fail at once | Only one upgrade check runs — all goroutines share the result via `sync.Once` |

---

## Build

```bash
cd A/musidl
go build -o musidl musidl.go
```

---

## Usage

```bash
# Interactive (prompts for artist name)
./musidl

# Direct
./musidl "Burna Boy"
./musidl "Taylor Swift"
./musidl "BTS"

# Custom options
./musidl --workers 8 --out ~/Downloads/Music "Wizkid"
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--workers N` | 4 | Parallel download workers |
| `--out DIR` | `~/Music/<artist>` | Output directory |

### Selection syntax

| Input | Meaning |
|-------|---------|
| Enter (blank) | Download all tracks |
| `1-20` | Tracks 1 through 20 |
| `1,3,5,10` | Specific tracks |
| `1-10,15,20-25` | Mix of ranges and singles |

---

## Country Coverage

All 175 Apple Music storefronts queried in parallel:

| Region | Countries | Count |
|--------|-----------|-------|
| North America | US, CA, MX | 3 |
| Latin America & Caribbean | AI, AG, AR, BS, BB, BZ, BM, BO, BR, VG, KY, CL, CO, CR, DM, DO, EC, SV, GD, GT, GY, HN, JM, MS, NI, PA, PY, PE, KN, LC, VC, SR, TC, TT, UY, VE | 36 |
| Europe | AL, AM, AT, AZ, BY, BE, BA, BG, HR, CY, CZ, DK, EE, FI, FR, GE, DE, GR, HU, IS, IE, IT, KZ, KG, LV, LT, LU, MK, MT, MD, ME, NL, NO, PL, PT, RO, RU, RS, SK, SI, ES, SE, CH, TJ, TM, TR, UA, GB, UZ | 48 |
| Middle East | BH, IQ, IL, JO, KW, LB, OM, QA, SA, AE, YE | 11 |
| Africa | DZ, AO, BJ, BW, BF, CM, CV, TD, CI, CD, CG, DJ, EG, ET, GA, GM, GH, GN, GW, KE, LS, LR, LY, MG, MW, ML, MR, MU, MA, MZ, NA, NE, NG, RW, ST, SN, SC, SL, ZA, SZ, TZ, TN, UG, ZM, ZW | 45 |
| Asia Pacific | AU, BD, BT, BN, KH, CN, FJ, FM, HK, IN, ID, JP, KR, LA, MO, MY, MV, MN, MM, NP, NZ, PG, PH, PK, PW, SG, LK, TW, TH, TL, TO, VU, VN, WS, PF | 35 |

---

## Output

Files saved to `~/Music/<Artist>/` by default:
```
001 - Ye.mp3             (4.2 MB — full track)
002 - Last Last.mp3      (6.1 MB — full track)
003 - City Boys.mp3      (7.3 MB — full track)
```

Files already downloaded are skipped automatically (skip threshold: >50 KB).

---

## Source APIs

| Source | Auth Required | Notes |
|--------|-------------|-------|
| iTunes Search API | None | 200 results per storefront, 175 countries |
| Deezer Search API | None | Paginated, up to 500 tracks per artist |
| MusicBrainz API | None | 100 recordings per query, open database |
| YouTube (yt-dlp) | None | Primary full-track download engine |
| SoundCloud (yt-dlp) | None | Fallback download engine |

---

## Performance

| Metric | Value |
|--------|-------|
| Search time | ~2–5 seconds (179 parallel requests) |
| Tracks per search | 300–500+ depending on artist |
| Full track size | 3–10 MB per MP3 |
| Go dependencies | 0 |

