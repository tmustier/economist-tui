# Demo fixtures

The `fair-exchange.txt` and `german-europe.txt` fixtures contain OCR excerpts from the Internet Archive scan of **The Economist, Sept 7 1940 (Vol. 139, Iss. 5063)**:

- Source text: https://archive.org/details/sim_economist_1940-09-07_139_5063
- OCR file: https://archive.org/download/sim_economist_1940-09-07_139_5063/sim_economist_1940-09-07_139_5063_djvu.txt

Additional fixtures include leader excerpts provided for demo mode:

- `lincoln-assassination.txt` — **The Economist, 1865 (archive volume)**
- `june-the-sixth.txt` — **The Economist, Jun 10th 1944**
- `atom-bomb.txt` — **The Economist, Aug 11th 1945**
- `policy-for-korea.txt` — **The Economist, Mar 31st 1951**
- `human-use-of-human-beings.txt` — **The Economist, Jul 14th 1951**
- `burnt-fingers-public-pulse.txt` — **The Economist, Nov 22nd 1952**
- `death-of-stalin.txt` — **The Economist, Mar 7th 1953**
- `rockets-for-reindeer.txt` — **The Economist, Dec 26th 1953**
- `electronic-abacus.txt` — **The Economist, Mar 13th 1954**
- `president-to-the-rescue.txt` — **The Economist, Oct 30th 1954**
- `algerian-dilemma.txt` — **The Economist, Oct 1st 1955**
- `freeing-electronics.txt` — **The Economist, Feb 4th 1956**
- `ends-and-means-at-suez.txt` — **The Economist, Sep 8th 1956**
- `hungary-european-future.txt` — **The Economist, Nov 10th 1956**
- `asian-milestone.txt` — **The Economist, Aug 31st 1957**
- `both-sides-moon.txt` — **The Economist, Oct 12th 1957**
- `riding-on-the-sputnik.txt` — **The Economist, Nov 9th 1957**
- `young-mans-america.txt` — **The Economist, Nov 12th 1960**

Fixture metadata (titles, subtitles, dates, ordering, sources) lives in `index.json`. Dates accept `YYYY-MM-DD` or human-readable formats like `Nov 12th 1960`. Subtitles are invented for illustration; think about what the article is saying and write a concise Economist-style subtitle rather than an obvious restatement. When a title is vague, it’s fine to be plainer (e.g., “Asian Milestone” is subtitled “Malaya’s independence” because that’s what it covers, even if the title alone doesn’t make it obvious).

## Adding fixtures

1. Find an OCR source on archive.org and copy a short excerpt (a few paragraphs) into a new `fixtures/<slug>.txt` file. End the excerpt with `■`.
2. Add an entry to `index.json` with `slug`, `title`, `subtitle`, `file`, `date`, and `source` (archive.org details page). Keep the list sorted by date.
3. Use short, readable subtitles (they’re curated, not extracted).
4. Run `go test ./internal/demo` to ensure the fixtures load and sort correctly.

They are included locally so demo mode works offline for all users. The Online Books Page notes that no issue or contribution copyright renewals were found for this serial; please verify suitability for your jurisdiction if reusing elsewhere, especially for the later excerpts.
