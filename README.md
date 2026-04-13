# resumex

`resumex` is a small Go CLI that turns a JSON Resume file into a PDF using a Go-native clone of the classic theme?

It implements  the rendering path needed for this job:

1. Read one JSON file.
2. Decode only the JSON Resume fields used by the classic theme.
3. Render a complete HTML document with embedded CSS and fonts.
4. Ask local Chrome or Chromium to print that HTML to PDF.

## Usage

```sh
go run ./cmd/resumex [-o output.pdf] [-chrome /path/to/chrome] [-scale 0.9] resume.json
```

Examples:

```sh
go run ./cmd/resumex ../jsonresumex/packages/test-fixtures/resume-sample.json -o /tmp/resume.pdf
go run ./cmd/resumex resume.json -scale 0.85 -o resume.pdf
go build -o resumex ./cmd/resumex
./resumex resume.json
```

If `-o` is omitted, the PDF is written beside the input file as `<input-basename>.pdf`.

If `-scale` is omitted, the theme renders at `1.0`. Use a smaller value, such as `0.9` or `0.85`, to reduce font size and vertical spacing so more content fits on a page.

If `-chrome` is omitted, the CLI tries to find Chrome or Chromium using common install paths and executable names. On macOS it checks `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`.

## Architecture

The code is split into four small pieces:

- `cmd/resumex`: CLI entrypoint. It parses flags, opens the input JSON, calls the theme renderer with the selected scale, and sends the resulting HTML to the PDF backend.
- `internal/resume`: Narrow JSON Resume model. It defines the fields consumed by the theme and decodes JSON into those types.
- `internal/theme`: Classic theme clone. It maps resume data into view structs, renders markdown with `goldmark`, formats JSON Resume dates, embeds the Latin Modern font files with `go:embed`, and produces one self-contained HTML document.
- `internal/pdf`: Chrome PDF backend. It writes the HTML to a temporary file, invokes headless Chrome or Chromium with `--print-to-pdf`, waits for a valid `%PDF` output, and cleans up temporary files.

The core data flow is:

```text
resume.json
  -> internal/resume.Decode
  -> internal/theme.RenderWithOptions
  -> temporary resume.html
  -> local Chrome/Chromium print-to-PDF
  -> output.pdf
```

## Theme Port

- The theme is implemented with Go `html/template`.
- The `-scale` flag adjusts the theme's base font size and spacing before PDF generation.
- The section order matches the classic theme: summary, experience, projects, education, certificates, publications, awards, volunteer, languages, skills, interests, references.
- Markdown is supported for summaries, descriptions, and bullet items through `github.com/yuin/goldmark`.
- Raw HTML in markdown input is escaped before rendering.
- The three Latin Modern font files used by the theme are embedded into the binary as data URLs.
- Icons are small inline SVG snippets, avoiding a separate icon dependency.

The goal is a close Go-native clone, not byte-for-byte identical React output.

## Runtime Requirements

- Go 1.22 or newer.
- Chrome or Chromium installed locally for PDF generation.
- One Go dependency: `github.com/yuin/goldmark`.

## Tests

There is one minimal integration test. It skips if Chrome or Chromium is not available.

```sh
go test ./...
```

The test runs the CLI against `testdata/resume-sample.json` and verifies that a non-trivial PDF starting with `%PDF` is produced.
