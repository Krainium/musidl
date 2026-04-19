package main

// musidl v2 — Global Music Downloader
// Searches iTunes (175 storefronts), Deezer, and MusicBrainz for any artist's
// complete worldwide catalog. Downloads FULL tracks via yt-dlp (YouTube primary,
// SoundCloud secondary). Zero-preview fallback — full songs only.
//
// Coverage: all 6 world regions · 175 Apple Music country storefronts
// Download sources: YouTube Music, SoundCloud (via yt-dlp)
// Author: krainium

import (
        "bufio"
        "encoding/json"
        "flag"
        "fmt"
        "net/http"
        "net/url"
        "os"
        "os/exec"
        "path/filepath"
        "sort"
        "strconv"
        "strings"
        "sync"
        "sync/atomic"
        "time"
)

// ── ANSI colours ─────────────────────────────────────────────────────────────

const (
        Reset   = "\033[0m"
        Bold    = "\033[1m"
        Dim     = "\033[2m"
        Red     = "\033[91m"
        Green   = "\033[92m"
        Yellow  = "\033[93m"
        Blue    = "\033[94m"
        Magenta = "\033[95m"
        Cyan    = "\033[96m"
        White   = "\033[97m"
)

// ── iTunes storefronts — ALL 175 Apple Music countries ────────────────────────

type storefront struct {
        Code   string
        Name   string
        Region string
}

var storefronts = []storefront{
        // ── North America (3) ──────────────────────────────────────────────────
        {"US", "United States", "North America"},
        {"CA", "Canada", "North America"},
        {"MX", "Mexico", "North America"},

        // ── Latin America & Caribbean (35) ────────────────────────────────────
        {"AI", "Anguilla", "Latin America"},
        {"AG", "Antigua & Barbuda", "Latin America"},
        {"AR", "Argentina", "Latin America"},
        {"BS", "Bahamas", "Latin America"},
        {"BB", "Barbados", "Latin America"},
        {"BZ", "Belize", "Latin America"},
        {"BM", "Bermuda", "Latin America"},
        {"BO", "Bolivia", "Latin America"},
        {"BR", "Brazil", "Latin America"},
        {"VG", "British Virgin Islands", "Latin America"},
        {"KY", "Cayman Islands", "Latin America"},
        {"CL", "Chile", "Latin America"},
        {"CO", "Colombia", "Latin America"},
        {"CR", "Costa Rica", "Latin America"},
        {"DM", "Dominica", "Latin America"},
        {"DO", "Dominican Republic", "Latin America"},
        {"EC", "Ecuador", "Latin America"},
        {"SV", "El Salvador", "Latin America"},
        {"GD", "Grenada", "Latin America"},
        {"GT", "Guatemala", "Latin America"},
        {"GY", "Guyana", "Latin America"},
        {"HN", "Honduras", "Latin America"},
        {"JM", "Jamaica", "Latin America"},
        {"MS", "Montserrat", "Latin America"},
        {"NI", "Nicaragua", "Latin America"},
        {"PA", "Panama", "Latin America"},
        {"PY", "Paraguay", "Latin America"},
        {"PE", "Peru", "Latin America"},
        {"KN", "Saint Kitts & Nevis", "Latin America"},
        {"LC", "Saint Lucia", "Latin America"},
        {"VC", "St. Vincent & Grenadines", "Latin America"},
        {"SR", "Suriname", "Latin America"},
        {"TC", "Turks & Caicos", "Latin America"},
        {"TT", "Trinidad & Tobago", "Latin America"},
        {"UY", "Uruguay", "Latin America"},
        {"VE", "Venezuela", "Latin America"},

        // ── Europe (48) ───────────────────────────────────────────────────────
        {"AL", "Albania", "Europe"},
        {"AM", "Armenia", "Europe"},
        {"AT", "Austria", "Europe"},
        {"AZ", "Azerbaijan", "Europe"},
        {"BY", "Belarus", "Europe"},
        {"BE", "Belgium", "Europe"},
        {"BA", "Bosnia & Herzegovina", "Europe"},
        {"BG", "Bulgaria", "Europe"},
        {"HR", "Croatia", "Europe"},
        {"CY", "Cyprus", "Europe"},
        {"CZ", "Czech Republic", "Europe"},
        {"DK", "Denmark", "Europe"},
        {"EE", "Estonia", "Europe"},
        {"FI", "Finland", "Europe"},
        {"FR", "France", "Europe"},
        {"GE", "Georgia", "Europe"},
        {"DE", "Germany", "Europe"},
        {"GR", "Greece", "Europe"},
        {"HU", "Hungary", "Europe"},
        {"IS", "Iceland", "Europe"},
        {"IE", "Ireland", "Europe"},
        {"IT", "Italy", "Europe"},
        {"KZ", "Kazakhstan", "Europe"},
        {"KG", "Kyrgyzstan", "Europe"},
        {"LV", "Latvia", "Europe"},
        {"LT", "Lithuania", "Europe"},
        {"LU", "Luxembourg", "Europe"},
        {"MK", "North Macedonia", "Europe"},
        {"MT", "Malta", "Europe"},
        {"MD", "Moldova", "Europe"},
        {"ME", "Montenegro", "Europe"},
        {"NL", "Netherlands", "Europe"},
        {"NO", "Norway", "Europe"},
        {"PL", "Poland", "Europe"},
        {"PT", "Portugal", "Europe"},
        {"RO", "Romania", "Europe"},
        {"RU", "Russia", "Europe"},
        {"RS", "Serbia", "Europe"},
        {"SK", "Slovakia", "Europe"},
        {"SI", "Slovenia", "Europe"},
        {"ES", "Spain", "Europe"},
        {"SE", "Sweden", "Europe"},
        {"CH", "Switzerland", "Europe"},
        {"TJ", "Tajikistan", "Europe"},
        {"TM", "Turkmenistan", "Europe"},
        {"TR", "Turkey", "Europe"},
        {"UA", "Ukraine", "Europe"},
        {"GB", "United Kingdom", "Europe"},
        {"UZ", "Uzbekistan", "Europe"},

        // ── Middle East (11) ──────────────────────────────────────────────────
        {"BH", "Bahrain", "Middle East"},
        {"IQ", "Iraq", "Middle East"},
        {"IL", "Israel", "Middle East"},
        {"JO", "Jordan", "Middle East"},
        {"KW", "Kuwait", "Middle East"},
        {"LB", "Lebanon", "Middle East"},
        {"OM", "Oman", "Middle East"},
        {"QA", "Qatar", "Middle East"},
        {"SA", "Saudi Arabia", "Middle East"},
        {"AE", "United Arab Emirates", "Middle East"},
        {"YE", "Yemen", "Middle East"},

        // ── Africa (43) ───────────────────────────────────────────────────────
        {"DZ", "Algeria", "Africa"},
        {"AO", "Angola", "Africa"},
        {"BJ", "Benin", "Africa"},
        {"BW", "Botswana", "Africa"},
        {"BF", "Burkina Faso", "Africa"},
        {"CM", "Cameroon", "Africa"},
        {"CV", "Cape Verde", "Africa"},
        {"TD", "Chad", "Africa"},
        {"CI", "Côte d'Ivoire", "Africa"},
        {"CD", "DR Congo", "Africa"},
        {"CG", "Republic of Congo", "Africa"},
        {"DJ", "Djibouti", "Africa"},
        {"EG", "Egypt", "Africa"},
        {"ET", "Ethiopia", "Africa"},
        {"GA", "Gabon", "Africa"},
        {"GM", "Gambia", "Africa"},
        {"GH", "Ghana", "Africa"},
        {"GN", "Guinea", "Africa"},
        {"GW", "Guinea-Bissau", "Africa"},
        {"KE", "Kenya", "Africa"},
        {"LS", "Lesotho", "Africa"},
        {"LR", "Liberia", "Africa"},
        {"LY", "Libya", "Africa"},
        {"MG", "Madagascar", "Africa"},
        {"MW", "Malawi", "Africa"},
        {"ML", "Mali", "Africa"},
        {"MR", "Mauritania", "Africa"},
        {"MU", "Mauritius", "Africa"},
        {"MA", "Morocco", "Africa"},
        {"MZ", "Mozambique", "Africa"},
        {"NA", "Namibia", "Africa"},
        {"NE", "Niger", "Africa"},
        {"NG", "Nigeria", "Africa"},
        {"RW", "Rwanda", "Africa"},
        {"ST", "São Tomé & Príncipe", "Africa"},
        {"SN", "Senegal", "Africa"},
        {"SC", "Seychelles", "Africa"},
        {"SL", "Sierra Leone", "Africa"},
        {"ZA", "South Africa", "Africa"},
        {"SZ", "Eswatini", "Africa"},
        {"TZ", "Tanzania", "Africa"},
        {"TN", "Tunisia", "Africa"},
        {"UG", "Uganda", "Africa"},
        {"ZM", "Zambia", "Africa"},
        {"ZW", "Zimbabwe", "Africa"},

        // ── Asia Pacific (35) ─────────────────────────────────────────────────
        {"AU", "Australia", "Asia Pacific"},
        {"BD", "Bangladesh", "Asia Pacific"},
        {"BT", "Bhutan", "Asia Pacific"},
        {"BN", "Brunei", "Asia Pacific"},
        {"KH", "Cambodia", "Asia Pacific"},
        {"CN", "China", "Asia Pacific"},
        {"FJ", "Fiji", "Asia Pacific"},
        {"FM", "Micronesia", "Asia Pacific"},
        {"HK", "Hong Kong", "Asia Pacific"},
        {"IN", "India", "Asia Pacific"},
        {"ID", "Indonesia", "Asia Pacific"},
        {"JP", "Japan", "Asia Pacific"},
        {"KR", "South Korea", "Asia Pacific"},
        {"LA", "Laos", "Asia Pacific"},
        {"MO", "Macao", "Asia Pacific"},
        {"MY", "Malaysia", "Asia Pacific"},
        {"MV", "Maldives", "Asia Pacific"},
        {"MN", "Mongolia", "Asia Pacific"},
        {"MM", "Myanmar", "Asia Pacific"},
        {"NP", "Nepal", "Asia Pacific"},
        {"NZ", "New Zealand", "Asia Pacific"},
        {"PG", "Papua New Guinea", "Asia Pacific"},
        {"PH", "Philippines", "Asia Pacific"},
        {"PK", "Pakistan", "Asia Pacific"},
        {"PW", "Palau", "Asia Pacific"},
        {"SG", "Singapore", "Asia Pacific"},
        {"LK", "Sri Lanka", "Asia Pacific"},
        {"TW", "Taiwan", "Asia Pacific"},
        {"TH", "Thailand", "Asia Pacific"},
        {"TL", "Timor-Leste", "Asia Pacific"},
        {"TO", "Tonga", "Asia Pacific"},
        {"VU", "Vanuatu", "Asia Pacific"},
        {"VN", "Vietnam", "Asia Pacific"},
        {"WS", "Samoa", "Asia Pacific"},
        {"PF", "French Polynesia", "Asia Pacific"},
}

// ── Track data model ─────────────────────────────────────────────────────────

type Track struct {
        ID         string
        Title      string
        Artist     string
        Album      string
        Duration   int
        Year       string
        Genre      string
        PreviewURL string
        ArtworkURL string
        Source     string
}

func trackKey(t Track) string {
        norm := func(s string) string {
                return strings.ToLower(strings.Join(strings.Fields(s), " "))
        }
        return norm(t.Title) + "|||" + norm(t.Artist)
}

// ── Print helpers ─────────────────────────────────────────────────────────────

var stdin = bufio.NewReader(os.Stdin)

func oinfo(msg string)    { fmt.Printf("  %s%s[*]%s %s\n", Bold, Cyan, Reset, msg) }
func osuccess(msg string) { fmt.Printf("  %s%s[+]%s %s%s%s\n", Bold, Green, Reset, Green, msg, Reset) }
func owarn(msg string)    { fmt.Printf("  %s%s[!]%s %s%s%s\n", Bold, Yellow, Reset, Yellow, msg, Reset) }
func oerror(msg string)   { fmt.Printf("  %s%s[-]%s %s%s%s\n", Bold, Red, Reset, Red, msg, Reset) }
func ostep(msg string)    { fmt.Printf("  %s%s[>]%s %s%s%s\n", Bold, Magenta, Reset, White, msg, Reset) }
func odivider()           { fmt.Printf("  %s%s%s\n", Dim, strings.Repeat("─", 66), Reset) }

func oheader(title string) {
        fmt.Println()
        odivider()
        fmt.Printf("  %s%s%s\n", Bold+White, title, Reset)
        odivider()
        fmt.Println()
}

func obanner() {
        b := Cyan
        y := Bold + Yellow
        r := Reset
        fmt.Println()
        fmt.Printf("  %s+══════════════════════════════════════════════════════════+%s\n", b, r)
        fmt.Printf("  %s|%s  %s███╗   ███╗██╗   ██╗███████╗██╗██████╗ ██╗      %s%s|%s\n", b, r, y, r, b, r)
        fmt.Printf("  %s|%s  %s████╗ ████║██║   ██║██╔════╝██║██╔══██╗██║      %s%s|%s\n", b, r, y, r, b, r)
        fmt.Printf("  %s|%s  %s██╔████╔██║██║   ██║███████╗██║██║  ██║██║      %s%s|%s\n", b, r, y, r, b, r)
        fmt.Printf("  %s|%s  %s██║╚██╔╝██║██║   ██║╚════██║██║██║  ██║██║      %s%s|%s\n", b, r, y, r, b, r)
        fmt.Printf("  %s|%s  %s██║ ╚═╝ ██║╚██████╔╝███████║██║██████╔╝███████╗ %s%s|%s\n", b, r, y, r, b, r)
        fmt.Printf("  %s|%s  %s╚═╝     ╚═╝ ╚═════╝ ╚══════╝╚═╝╚═════╝ ╚══════╝%s%s|%s\n", b, r, y, r, b, r)
        fmt.Printf("  %s|%s                                                            %s|%s\n", b, r, b, r)
        fmt.Printf("  %s|%s  %sGlobal Music Downloader  v2.0%s  %sgithub.com/krainium%s       %s|%s\n", b, r, White, r, Dim, r, b, r)
        fmt.Printf("  %s|%s  %siTunes 175 countries · Deezer · MusicBrainz · yt-dlp%s  %s|%s\n", b, r, Dim, r, b, r)
        fmt.Printf("  %s+══════════════════════════════════════════════════════════+%s\n", b, r)
        odivider()
        fmt.Println()
}

func prompt(msg string) string {
        fmt.Printf("  %s%s[>]%s %s%s%s: ", Bold, Magenta, Reset, White, msg, Reset)
        line, _ := stdin.ReadString('\n')
        return strings.TrimSpace(line)
}

// ── HTTP helper ───────────────────────────────────────────────────────────────

var httpClient = &http.Client{
        Timeout: 15 * time.Second,
        Transport: &http.Transport{
                MaxIdleConnsPerHost: 30,
                DisableCompression:  false,
        },
}

func httpGet(rawURL string) (*http.Response, error) {
        req, err := http.NewRequest("GET", rawURL, nil)
        if err != nil {
                return nil, err
        }
        req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
        req.Header.Set("Accept", "application/json, */*")
        req.Header.Set("Accept-Language", "en-US,en;q=0.9")
        return httpClient.Do(req)
}

// ── iTunes Search API ─────────────────────────────────────────────────────────

type itunesResp struct {
        ResultCount int `json:"resultCount"`
        Results     []struct {
                Kind             string `json:"kind"`
                TrackId          int    `json:"trackId"`
                TrackName        string `json:"trackName"`
                ArtistName       string `json:"artistName"`
                CollectionName   string `json:"collectionName"`
                PreviewUrl       string `json:"previewUrl"`
                ArtworkUrl100    string `json:"artworkUrl100"`
                TrackTimeMillis  int    `json:"trackTimeMillis"`
                ReleaseDate      string `json:"releaseDate"`
                PrimaryGenreName string `json:"primaryGenreName"`
        } `json:"results"`
}

func searchItunes(query, country string) []Track {
        params := url.Values{}
        params.Set("term", query)
        params.Set("entity", "song")
        params.Set("limit", "200")
        params.Set("country", country)
        params.Set("lang", "en_us")

        resp, err := httpGet("https://itunes.apple.com/search?" + params.Encode())
        if err != nil {
                return nil
        }
        defer resp.Body.Close()

        var data itunesResp
        if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
                return nil
        }

        var tracks []Track
        for _, r := range data.Results {
                if r.Kind != "song" || r.TrackName == "" {
                        continue
                }
                year := ""
                if len(r.ReleaseDate) >= 4 {
                        year = r.ReleaseDate[:4]
                }
                tracks = append(tracks, Track{
                        ID:         strconv.Itoa(r.TrackId),
                        Title:      r.TrackName,
                        Artist:     r.ArtistName,
                        Album:      r.CollectionName,
                        Duration:   r.TrackTimeMillis / 1000,
                        Year:       year,
                        Genre:      r.PrimaryGenreName,
                        PreviewURL: r.PreviewUrl,
                        ArtworkURL: r.ArtworkUrl100,
                        Source:     "iTunes/" + country,
                })
        }
        return tracks
}

// ── Deezer Search API ─────────────────────────────────────────────────────────

type deezerResp struct {
        Data []struct {
                Id       int    `json:"id"`
                Title    string `json:"title"`
                Preview  string `json:"preview"`
                Duration int    `json:"duration"`
                Artist   struct {
                        Name string `json:"name"`
                } `json:"artist"`
                Album struct {
                        Title       string `json:"title"`
                        CoverMedium string `json:"cover_medium"`
                } `json:"album"`
        } `json:"data"`
        Next  string `json:"next"`
        Total int    `json:"total"`
}

func searchDeezer(query string) []Track {
        var all []Track
        nextURL := "https://api.deezer.com/search?q=" + url.QueryEscape(`artist:"`+query+`"`) + "&limit=100"
        for nextURL != "" && len(all) < 500 {
                resp, err := httpGet(nextURL)
                if err != nil {
                        break
                }
                var data deezerResp
                if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
                        resp.Body.Close()
                        break
                }
                resp.Body.Close()
                for _, r := range data.Data {
                        if r.Title == "" {
                                continue
                        }
                        all = append(all, Track{
                                ID:         strconv.Itoa(r.Id),
                                Title:      r.Title,
                                Artist:     r.Artist.Name,
                                Album:      r.Album.Title,
                                Duration:   r.Duration,
                                PreviewURL: r.Preview,
                                ArtworkURL: r.Album.CoverMedium,
                                Source:     "deezer",
                        })
                }
                if data.Next != "" {
                        nextURL = data.Next
                } else {
                        break
                }
        }
        return all
}

// ── MusicBrainz API ───────────────────────────────────────────────────────────

type mbResp struct {
        Recordings []struct {
                ID    string `json:"id"`
                Title string `json:"title"`
                Length int   `json:"length"`
                ArtistCredit []struct {
                        Artist struct {
                                Name string `json:"name"`
                        } `json:"artist"`
                } `json:"artist-credit"`
                Releases []struct {
                        Title string `json:"title"`
                        Date  string `json:"date"`
                } `json:"releases"`
        } `json:"recordings"`
}

func searchMusicBrainz(query string) []Track {
        enc := url.QueryEscape(`artist:"` + query + `"`)
        apiURL := "https://musicbrainz.org/ws/2/recording/?query=" + enc + "&limit=100&fmt=json"

        req, err := http.NewRequest("GET", apiURL, nil)
        if err != nil {
                return nil
        }
        req.Header.Set("User-Agent", "musidl/2.0 (https://github.com/krainium)")
        req.Header.Set("Accept", "application/json")

        mbClient := &http.Client{Timeout: 20 * time.Second}
        resp, err := mbClient.Do(req)
        if err != nil {
                return nil
        }
        defer resp.Body.Close()

        var data mbResp
        if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
                return nil
        }

        var tracks []Track
        for _, r := range data.Recordings {
                if r.Title == "" {
                        continue
                }
                artistName := query
                if len(r.ArtistCredit) > 0 {
                        artistName = r.ArtistCredit[0].Artist.Name
                }
                album, year := "", ""
                if len(r.Releases) > 0 {
                        album = r.Releases[0].Title
                        if len(r.Releases[0].Date) >= 4 {
                                year = r.Releases[0].Date[:4]
                        }
                }
                tracks = append(tracks, Track{
                        ID:       r.ID,
                        Title:    r.Title,
                        Artist:   artistName,
                        Album:    album,
                        Duration: r.Length / 1000,
                        Year:     year,
                        Source:   "musicbrainz",
                })
        }
        return tracks
}

// ── Multi-source parallel search ──────────────────────────────────────────────

func searchAll(query string) []Track {
        type result struct{ tracks []Track }

        // 175 iTunes storefronts + Deezer + MusicBrainz
        total := len(storefronts) + 2
        ch := make(chan result, total)

        // iTunes: one goroutine per storefront (with rate-limit jitter)
        for i, sf := range storefronts {
                sf := sf
                delay := time.Duration(i/20) * 200 * time.Millisecond
                go func() {
                        if delay > 0 {
                                time.Sleep(delay)
                        }
                        ch <- result{searchItunes(query, sf.Code)}
                }()
        }

        // Deezer
        go func() { ch <- result{searchDeezer(query)} }()

        // MusicBrainz
        go func() { ch <- result{searchMusicBrainz(query)} }()

        seen := make(map[string]bool)
        var all []Track
        for i := 0; i < total; i++ {
                for _, t := range (<-ch).tracks {
                        k := trackKey(t)
                        if !seen[k] {
                                seen[k] = true
                                all = append(all, t)
                        }
                }
        }

        sort.Slice(all, func(i, j int) bool {
                if all[i].Year != all[j].Year {
                        return all[i].Year > all[j].Year
                }
                return strings.ToLower(all[i].Title) < strings.ToLower(all[j].Title)
        })
        return all
}

// ── yt-dlp detection & auto-install ──────────────────────────────────────────

func ytdlpCandidates() []string {
        return []string{
                "yt-dlp",
                os.ExpandEnv("$HOME/.local/bin/yt-dlp"),
                "/home/runner/workspace/.pythonlibs/bin/yt-dlp",
                "/usr/local/bin/yt-dlp",
                "/usr/bin/yt-dlp",
        }
}

// findYtdlp scans common locations and PATH for an existing yt-dlp binary.
func findYtdlp() string {
        for _, p := range ytdlpCandidates() {
                if path, err := exec.LookPath(p); err == nil {
                        return path
                }
        }
        return ""
}

// installYtdlp attempts a one-time pip install and returns the binary path.
// Prints progress so the user knows what is happening.
func installYtdlp() string {
        owarn("yt-dlp not found — installing automatically (one-time setup)...")
        pips := []string{"pip", "pip3", "python3 -m pip", "python -m pip"}
        anyPipFound := false
        for _, pip := range pips {
                parts := strings.Fields(pip)
                if _, err := exec.LookPath(parts[0]); err != nil {
                        continue
                }
                anyPipFound = true
                args := append(parts[1:], "install", "--quiet", "yt-dlp")
                cmd := exec.Command(parts[0], args...)
                cmd.Stdout = os.Stdout
                cmd.Stderr = os.Stderr
                if err := cmd.Run(); err == nil {
                        osuccess("yt-dlp installed successfully")
                        return findYtdlp()
                }
                owarn(fmt.Sprintf("%s install failed — trying next installer...", parts[0]))
        }
        if !anyPipFound {
                oerror("pip / pip3 not found on this system.")
                oerror("Install Python first, then run: pip install yt-dlp")
        }
        return ""
}

// checkAndUpdateYtdlp checks once per week whether a newer yt-dlp release is
// available on PyPI and, if so, upgrades silently via pip. All errors are
// swallowed so that a failed check never interrupts a download session.
// The timestamp is written only when the check completes without transient
// errors, so network outages do not suppress future check attempts.
// A human-readable status string is sent to ch when the check concludes;
// an empty string is sent when the check is skipped or encounters an error.
func checkAndUpdateYtdlp(ytdlpPath string, ch chan<- string) {
        home, err := os.UserHomeDir()
        if err != nil {
                ch <- ""
                return
        }
        cacheDir := filepath.Join(home, ".musidl")
        _ = os.MkdirAll(cacheDir, 0755)
        stampFile := filepath.Join(cacheDir, ".last_update_check")

        const week = 7 * 24 * time.Hour

        if data, err := os.ReadFile(stampFile); err == nil {
                if ts, err := time.Parse(time.RFC3339, strings.TrimSpace(string(data))); err == nil {
                        if time.Since(ts) < week {
                                ch <- ""
                                return
                        }
                }
        }

        client := &http.Client{Timeout: 10 * time.Second}
        resp, err := client.Get("https://pypi.org/pypi/yt-dlp/json")
        if err != nil {
                ch <- ""
                return
        }
        defer resp.Body.Close()
        if resp.StatusCode != http.StatusOK {
                ch <- ""
                return
        }

        var pypi struct {
                Info struct {
                        Version string `json:"version"`
                } `json:"info"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&pypi); err != nil {
                ch <- ""
                return
        }
        latest := strings.TrimSpace(pypi.Info.Version)
        if latest == "" {
                ch <- ""
                return
        }

        out, err := exec.Command(ytdlpPath, "--version").Output()
        if err != nil {
                ch <- ""
                return
        }
        installed := strings.TrimSpace(string(out))

        // Stamp the check time now — we have successfully determined whether an
        // update is needed; the outcome of the pip step is treated as best-effort.
        _ = os.WriteFile(stampFile, []byte(time.Now().UTC().Format(time.RFC3339)), 0644)

        if installed == latest {
                ch <- fmt.Sprintf("yt-dlp is up to date (%s)", installed)
                return
        }

        pips := []string{"pip", "pip3", "python3 -m pip", "python -m pip"}
        for _, pip := range pips {
                parts := strings.Fields(pip)
                if _, err := exec.LookPath(parts[0]); err != nil {
                        continue
                }
                args := append(parts[1:], "install", "--quiet", "--upgrade", "yt-dlp")
                cmd := exec.Command(parts[0], args...)
                if err := cmd.Run(); err == nil {
                        ch <- fmt.Sprintf("yt-dlp updated to %s", latest)
                        return
                }
        }
        ch <- ""
}

// tryUpgradeYtdlp checks PyPI unconditionally for a newer yt-dlp release and
// installs it if one is available. On success it returns the path to the
// (possibly relocated) yt-dlp binary, the installed version string, and true;
// otherwise it returns ("", "", false). All errors are swallowed so a failed
// attempt never blocks the caller. The weekly-check stamp is updated on success
// so the background check does not repeat the work immediately afterward.
func tryUpgradeYtdlp(ytdlpPath string) (string, string, bool) {
        client := &http.Client{Timeout: 10 * time.Second}
        resp, err := client.Get("https://pypi.org/pypi/yt-dlp/json")
        if err != nil {
                return "", "", false
        }
        defer resp.Body.Close()
        if resp.StatusCode != http.StatusOK {
                return "", "", false
        }

        var pypi struct {
                Info struct {
                        Version string `json:"version"`
                } `json:"info"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&pypi); err != nil {
                return "", "", false
        }
        latest := strings.TrimSpace(pypi.Info.Version)
        if latest == "" {
                return "", "", false
        }

        out, err := exec.Command(ytdlpPath, "--version").Output()
        if err != nil {
                return "", "", false
        }
        installed := strings.TrimSpace(string(out))
        if installed == latest {
                return "", "", false
        }

        pips := []string{"pip", "pip3", "python3 -m pip", "python -m pip"}
        for _, pip := range pips {
                parts := strings.Fields(pip)
                if _, err := exec.LookPath(parts[0]); err != nil {
                        continue
                }
                args := append(parts[1:], "install", "--quiet", "--upgrade", "yt-dlp")
                cmd := exec.Command(parts[0], args...)
                if err := cmd.Run(); err == nil {
                        // Update the weekly stamp so the scheduled check does not
                        // repeat the upgrade immediately after.
                        if home, err := os.UserHomeDir(); err == nil {
                                cacheDir := filepath.Join(home, ".musidl")
                                _ = os.MkdirAll(cacheDir, 0755)
                                stampFile := filepath.Join(cacheDir, ".last_update_check")
                                _ = os.WriteFile(stampFile, []byte(time.Now().UTC().Format(time.RFC3339)), 0644)
                        }
                        // Re-discover the binary: pip may have installed the new
                        // version at a different path than the original executable.
                        if newPath := findYtdlp(); newPath != "" {
                                return newPath, latest, true
                        }
                        return ytdlpPath, latest, true
                }
        }
        return "", "", false
}

// ── Downloader ────────────────────────────────────────────────────────────────

var (
        printMu sync.Mutex
        dlDone  int64
        dlFail  int64
        dlSkip  int64
)

// upgradeSession ensures that only one goroutine performs the yt-dlp upgrade
// check per download session. All other goroutines that encounter a failure
// block inside tryOnce until the single check completes, then reuse its result.
// noticeOnce guarantees the "upgraded" banner is printed at most once per
// session, and only when a post-upgrade retry actually succeeds.
type upgradeSession struct {
        once       sync.Once
        noticeOnce sync.Once
        newPath    string
        newVersion string
        upgraded   bool
}

func (s *upgradeSession) tryOnce(ytdlp string) (string, string, bool) {
        s.once.Do(func() {
                s.newPath, s.newVersion, s.upgraded = tryUpgradeYtdlp(ytdlp)
        })
        return s.newPath, s.newVersion, s.upgraded
}

// printUpgradeNotice emits a one-time banner when a post-upgrade retry
// succeeds. It is safe to call from multiple goroutines; only the first
// successful caller will print anything.
func (s *upgradeSession) printUpgradeNotice() {
        s.noticeOnce.Do(func() {
                printMu.Lock()
                defer printMu.Unlock()
                fmt.Printf("\n  %s%s[+]%s %s%syt-dlp upgraded to %s — retrying failed tracks%s\n\n",
                        Bold, Green, Reset, Bold, Green, s.newVersion, Reset)
        })
}

func fmtDur(secs int) string {
        if secs <= 0 {
                return "?:??"
        }
        return fmt.Sprintf("%d:%02d", secs/60, secs%60)
}

func safeName(s string) string {
        var b strings.Builder
        for _, r := range s {
                switch r {
                case '/', '\\', ':', '*', '?', '"', '<', '>', '|', '\x00':
                        b.WriteByte('_')
                default:
                        b.WriteRune(r)
                }
        }
        result := strings.TrimSpace(b.String())
        if len(result) > 180 {
                result = result[:180]
        }
        return result
}

func truncStr(s string, n int) string {
        runes := []rune(s)
        if len(runes) <= n {
                return s + strings.Repeat(" ", n-len(runes))
        }
        return string(runes[:n-1]) + "…"
}

func printResult(idx int, t Track, status, extra string) {
        printMu.Lock()
        defer printMu.Unlock()
        var col, sym string
        switch status {
        case "done":
                col, sym = Green, "✓"
        case "skip":
                col, sym = Cyan, "→"
        case "fail":
                col, sym = Red, "✗"
        }
        fmt.Printf("  %s%s[%s]%s  %s%3d%s  %-38s  %s%-5s%s  %s%s%s\n",
                Bold, col, sym, Reset,
                Cyan, idx, Reset,
                truncStr(t.Title, 38),
                Dim, fmtDur(t.Duration), Reset,
                Dim, extra, Reset)
}

// downloadFull tries multiple yt-dlp strategies to find a full track.
// Strategy order:
//   1. YouTube search: "ytsearch1:ARTIST - TITLE"
//   2. YouTube search: "ytsearch1:ARTIST TITLE"
//   3. YouTube search: "ytsearch1:ARTIST TITLE official audio"
//   4. SoundCloud search: "scsearch1:ARTIST TITLE"
func downloadFull(ytdlp string, t Track, outBase string) error {
        queries := []string{
                fmt.Sprintf("ytsearch1:%s - %s", t.Artist, t.Title),
                fmt.Sprintf("ytsearch1:%s %s", t.Artist, t.Title),
                fmt.Sprintf("ytsearch1:%s %s official audio", t.Artist, t.Title),
                fmt.Sprintf("scsearch1:%s %s", t.Artist, t.Title),
        }
        for _, q := range queries {
                cmd := exec.Command(ytdlp,
                        "--no-playlist",
                        "--extract-audio",
                        "--audio-format", "mp3",
                        "--audio-quality", "0",
                        "--output", outBase+".%(ext)s",
                        "--no-progress",
                        "--quiet",
                        "--no-warnings",
                        q,
                )
                if err := cmd.Run(); err == nil {
                        // Verify the file was created
                        if _, err2 := os.Stat(outBase + ".mp3"); err2 == nil {
                                return nil
                        }
                }
        }
        return fmt.Errorf("all download strategies failed")
}

func downloadOne(t Track, outDir string, rank int, ytdlp string, wg *sync.WaitGroup, upg *upgradeSession) {
        defer wg.Done()

        fname := safeName(fmt.Sprintf("%03d - %s.mp3", rank, t.Title))
        outPath := filepath.Join(outDir, fname)
        outBase := filepath.Join(outDir, safeName(fmt.Sprintf("%03d - %s", rank, t.Title)))

        if info, err := os.Stat(outPath); err == nil && info.Size() > 50000 {
                printResult(rank, t, "skip", "exists")
                atomic.AddInt64(&dlSkip, 1)
                return
        }

        dlErr := downloadFull(ytdlp, t, outBase)
        if dlErr != nil {
                // A non-zero exit from yt-dlp may mean the site changed its API.
                // Check unconditionally for a newer yt-dlp and, if one was
                // installed, retry the track once with the fresh binary.
                // Use the path returned by tryUpgradeYtdlp: pip may have placed
                // the new version at a different location than the original binary.
                if newPath, _, upgraded := upg.tryOnce(ytdlp); upgraded {
                        dlErr = downloadFull(newPath, t, outBase)
                        if dlErr == nil {
                                // Retry succeeded thanks to the upgrade — show a
                                // one-time session notice so users know why.
                                upg.printUpgradeNotice()
                        }
                }
        }

        if dlErr == nil {
                if info, _ := os.Stat(outPath); info != nil {
                        size := fmt.Sprintf("%.1fMB", float64(info.Size())/(1<<20))
                        printResult(rank, t, "done", size)
                        atomic.AddInt64(&dlDone, 1)
                } else {
                        printResult(rank, t, "fail", "no file")
                        atomic.AddInt64(&dlFail, 1)
                }
        } else {
                printResult(rank, t, "fail", "not found")
                atomic.AddInt64(&dlFail, 1)
                // Clean up any partial files
                os.Remove(outBase + ".mp3")
                os.Remove(outBase + ".webm")
                os.Remove(outBase + ".m4a")
                os.Remove(outBase + ".opus")
        }
}

// ── Selection parser ──────────────────────────────────────────────────────────

func parseSelection(input string, max int) []int {
        input = strings.TrimSpace(input)
        if input == "" || strings.ToLower(input) == "a" || strings.ToLower(input) == "all" {
                out := make([]int, max)
                for i := range out {
                        out[i] = i + 1
                }
                return out
        }

        var out []int
        seen := make(map[int]bool)
        add := func(n int) {
                if n >= 1 && n <= max && !seen[n] {
                        seen[n] = true
                        out = append(out, n)
                }
        }

        for _, part := range strings.Split(input, ",") {
                part = strings.TrimSpace(part)
                if strings.Contains(part, "-") {
                        bounds := strings.SplitN(part, "-", 2)
                        lo, err1 := strconv.Atoi(strings.TrimSpace(bounds[0]))
                        hi, err2 := strconv.Atoi(strings.TrimSpace(bounds[1]))
                        if err1 == nil && err2 == nil {
                                for n := lo; n <= hi; n++ {
                                        add(n)
                                }
                        }
                } else if n, err := strconv.Atoi(part); err == nil {
                        add(n)
                }
        }
        sort.Ints(out)
        return out
}

// ── Track table printer ───────────────────────────────────────────────────────

func printTable(tracks []Track) {
        fmt.Printf("  %s%s  %-4s  %-40s  %-22s  %-6s  %s%s\n",
                Bold+White, "#", "Dur", "Title", "Album", "Year", "Source", Reset)
        odivider()
        for i, t := range tracks {
                num := fmt.Sprintf("%d", i+1)
                dur := fmtDur(t.Duration)
                title := truncStr(t.Title, 40)
                album := truncStr(t.Album, 22)
                year := t.Year
                if year == "" {
                        year = "----"
                }
                src := t.Source
                if len(src) > 12 {
                        src = src[:12]
                }
                numCol := Cyan
                if (i+1)%2 == 0 {
                        numCol = Blue
                }
                fmt.Printf("  %s%s%-4s%s  %s%-4s%s  %s%-40s%s  %s%-22s%s  %s%-6s%s  %s%s%s\n",
                        Bold, numCol, num, Reset,
                        Dim, dur, Reset,
                        White, title, Reset,
                        Dim, album, Reset,
                        Yellow, year, Reset,
                        Dim, src, Reset)
        }
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
        workers := flag.Int("workers", 4, "parallel download workers")
        outDir := flag.String("out", "", "output directory (default: ~/Music/<artist>)")
        flag.Parse()

        obanner()

        ytdlp := findYtdlp()
        if ytdlp == "" {
                ytdlp = installYtdlp()
        }
        if ytdlp == "" {
                oerror("Could not install yt-dlp automatically.")
                oerror("Install manually: pip install yt-dlp")
                os.Exit(1)
        }
        osuccess(fmt.Sprintf("yt-dlp ready: %s", ytdlp))
        updateCh := make(chan string, 1)
        go checkAndUpdateYtdlp(ytdlp, updateCh)

        // ── Get artist name ───────────────────────────────────────────────────
        artist := strings.Join(flag.Args(), " ")
        if artist == "" {
                artist = prompt("Artist name")
        }
        if artist == "" {
                oerror("No artist specified.")
                os.Exit(1)
        }

        // ── Search across all sources ─────────────────────────────────────────
        oheader(fmt.Sprintf("Searching: %s%s%s", Bold+Yellow, artist, Reset))
        oinfo(fmt.Sprintf("Querying %d iTunes storefronts + Deezer + MusicBrainz...", len(storefronts)))

        start := time.Now()
        tracks := searchAll(artist)
        elapsed := time.Since(start)

        if len(tracks) == 0 {
                oerror(fmt.Sprintf("No tracks found for: %s", artist))
                os.Exit(0)
        }

        osuccess(fmt.Sprintf("Found %s%d tracks%s in %.1fs across all worldwide sources",
                Bold, len(tracks), Reset, elapsed.Seconds()))

        // First (non-blocking) chance to print the yt-dlp update status.
        // The goroutine has been running while the user typed the artist name
        // and while searchAll made its network requests, so it is often done
        // by this point.
        updatePrinted := false
        select {
        case msg := <-updateCh:
                updatePrinted = true
                if msg != "" {
                        oinfo(msg)
                }
        default:
        }
        fmt.Println()

        // ── Display catalog table ─────────────────────────────────────────────
        oheader("Track Catalog")
        printTable(tracks)
        fmt.Println()

        // ── Selection ─────────────────────────────────────────────────────────
        oinfo("Enter track numbers to download")
        oinfo("  Press Enter = all tracks")
        oinfo("  1,3,5     = specific tracks")
        oinfo("  1-20      = a range")
        fmt.Println()
        selInput := prompt("Selection")
        selected := parseSelection(selInput, len(tracks))
        if len(selected) == 0 {
                oerror("No valid selection.")
                os.Exit(0)
        }

        // ── Set up output directory ───────────────────────────────────────────
        dir := *outDir
        if dir == "" {
                home, _ := os.UserHomeDir()
                dir = filepath.Join(home, "Music", safeName(artist))
        }
        if err := os.MkdirAll(dir, 0755); err != nil {
                oerror(fmt.Sprintf("Cannot create directory: %v", err))
                os.Exit(1)
        }

        // Guaranteed drain: if the first non-blocking poll missed the result
        // (goroutine was still running), wait up to 2 s before downloads begin.
        // By this point the goroutine has had artist-prompt + searchAll +
        // table-reading + selection time — practically always done already.
        if !updatePrinted {
                select {
                case msg := <-updateCh:
                        if msg != "" {
                                oinfo(msg)
                        }
                case <-time.After(2 * time.Second):
                }
                fmt.Println()
        }

        // ── Download ──────────────────────────────────────────────────────────
        oheader(fmt.Sprintf("Downloading %d tracks → %s", len(selected), dir))
        oinfo(fmt.Sprintf("Workers: %d  |  Source priority: YouTube → SoundCloud", *workers))
        fmt.Println()

        dlDone, dlFail, dlSkip = 0, 0, 0
        sem := make(chan struct{}, *workers)
        var wg sync.WaitGroup
        upg := &upgradeSession{}

        for _, idx := range selected {
                t := tracks[idx-1]
                wg.Add(1)
                sem <- struct{}{}
                go func(t Track, rank int) {
                        defer func() { <-sem }()
                        downloadOne(t, dir, rank, ytdlp, &wg, upg)
                }(t, idx)
        }
        wg.Wait()

        // ── Summary ───────────────────────────────────────────────────────────
        fmt.Println()
        odivider()
        done := atomic.LoadInt64(&dlDone)
        fail := atomic.LoadInt64(&dlFail)
        skip := atomic.LoadInt64(&dlSkip)
        fmt.Printf("  %s%s[+]%s %sDone%s  %s✓ %d downloaded%s  %s→ %d skipped%s  %s✗ %d failed%s\n",
                Bold, Green, Reset,
                Bold, Reset,
                Green, done, Reset,
                Cyan, skip, Reset,
                Red, fail, Reset)
        odivider()
        fmt.Println()
}
