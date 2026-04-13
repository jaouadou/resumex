package theme

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/yuin/goldmark"

	"resumex-cli/internal/resume"
)

//go:embed assets/fonts/lmroman10-regular.otf assets/fonts/lmroman10-bold.otf assets/fonts/lmroman10-italic.otf
var fontFiles embed.FS

var markdown = goldmark.New()

type Options struct {
	Scale float64
}

type view struct {
	Title    string
	Fonts    fonts
	CSS      cssScale
	Hero     hero
	Summary  template.HTML
	Sections []section
}

type cssScale struct {
	BaseFontPX         template.CSS
	LayoutBottomPX     template.CSS
	SectionBottomPX    template.CSS
	SectionBodyMarginX template.CSS
	HeroMarginTopPX    template.CSS
	HeroMarginBottomPX template.CSS
	ContactGapY        template.CSS
	ContactGapX        template.CSS
	IconMarginRightPX  template.CSS
	ExperienceBottomPX template.CSS
	MetaGapPX          template.CSS
	SummaryBottomPX    template.CSS
	ListPaddingLeftPX  template.CSS
	LineBottomPX       template.CSS
	LineListMarginPX   template.CSS
	ReferenceBottomPX  template.CSS
}

type fonts struct {
	Regular template.URL
	Bold    template.URL
	Italic  template.URL
}

type hero struct {
	Name     string
	Contacts []contact
}

type contact struct {
	Icon  template.HTML
	Label string
	URL   string
}

type section struct {
	Title       string
	Experiences []experience
	Lines       []line
	References  []reference
}

type experience struct {
	Title      string
	Subtitle   string
	Date       string
	Summary    template.HTML
	Highlights []template.HTML
}

type line struct {
	Name  string
	Items []string
}

type reference struct {
	Name string
	Text string
}

func Render(r *resume.Resume) (string, error) {
	return RenderWithOptions(r, Options{})
}

func RenderWithOptions(r *resume.Resume, options Options) (string, error) {
	data, err := buildView(r)
	if err != nil {
		return "", err
	}
	data.CSS = buildCSSScale(options.Scale)

	var out bytes.Buffer
	if err := pageTemplate.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}

func buildCSSScale(scale float64) cssScale {
	if scale <= 0 {
		scale = 1
	}
	return cssScale{
		BaseFontPX:         cssPX(11, scale),
		LayoutBottomPX:     cssPX(40, scale),
		SectionBottomPX:    cssPX(18, scale),
		SectionBodyMarginX: cssPX(8, scale),
		HeroMarginTopPX:    cssPX(20, scale),
		HeroMarginBottomPX: cssPX(20, scale),
		ContactGapY:        cssPX(10, scale),
		ContactGapX:        cssPX(20, scale),
		IconMarginRightPX:  cssPX(5, scale),
		ExperienceBottomPX: cssPX(10, scale),
		MetaGapPX:          cssPX(16, scale),
		SummaryBottomPX:    cssPX(5, scale),
		ListPaddingLeftPX:  cssPX(20, scale),
		LineBottomPX:       cssPX(5, scale),
		LineListMarginPX:   cssPX(5, scale),
		ReferenceBottomPX:  cssPX(15, scale),
	}
}

func cssPX(value, scale float64) template.CSS {
	return template.CSS(fmt.Sprintf("%.2fpx", value*scale))
}

func buildView(r *resume.Resume) (view, error) {
	fonts, err := loadFonts()
	if err != nil {
		return view{}, err
	}

	data := view{
		Title:   defaultString(r.Basics.Name, "Resume") + " - Resume",
		Fonts:   fonts,
		Hero:    buildHero(r.Basics),
		Summary: markdownBlock(r.Basics.Summary),
	}

	if len(r.Work) > 0 {
		var experiences []experience
		for _, item := range r.Work {
			experiences = append(experiences, experience{
				Title:      item.Position,
				Subtitle:   item.Name,
				Date:       formatDateRange(item.StartDate, item.EndDate),
				Summary:    markdownBlock(item.Summary),
				Highlights: markdownList(item.Highlights),
			})
		}
		data.Sections = append(data.Sections, section{Title: "Experience", Experiences: experiences})
	}

	if len(r.Projects) > 0 {
		var experiences []experience
		for _, item := range r.Projects {
			experiences = append(experiences, experience{
				Title:      item.Name,
				Date:       formatDateRange(item.StartDate, item.EndDate),
				Summary:    markdownBlock(item.Description),
				Highlights: markdownList(item.Highlights),
			})
		}
		data.Sections = append(data.Sections, section{Title: "Projects", Experiences: experiences})
	}

	if len(r.Education) > 0 {
		var experiences []experience
		for _, item := range r.Education {
			subtitle := item.StudyType
			if item.Area != "" {
				subtitle = item.StudyType + " in " + item.Area
			}
			if item.Score != "" {
				subtitle += " (" + item.Score + ")"
			}
			experiences = append(experiences, experience{
				Title:      item.Institution,
				Subtitle:   subtitle,
				Date:       formatDateRange(item.StartDate, item.EndDate),
				Highlights: markdownList(item.Courses),
			})
		}
		data.Sections = append(data.Sections, section{Title: "Education", Experiences: experiences})
	}

	if len(r.Certificates) > 0 {
		var experiences []experience
		for _, item := range r.Certificates {
			experiences = append(experiences, experience{
				Title:    item.Name,
				Subtitle: item.Issuer,
				Date:     formatDate(item.Date),
			})
		}
		data.Sections = append(data.Sections, section{Title: "Certificates", Experiences: experiences})
	}

	if len(r.Publications) > 0 {
		var experiences []experience
		for _, item := range r.Publications {
			experiences = append(experiences, experience{
				Title:    item.Name,
				Subtitle: item.Publisher,
				Date:     formatDate(item.ReleaseDate),
				Summary:  markdownBlock(item.Summary),
			})
		}
		data.Sections = append(data.Sections, section{Title: "Publications", Experiences: experiences})
	}

	if len(r.Awards) > 0 {
		var experiences []experience
		for _, item := range r.Awards {
			experiences = append(experiences, experience{
				Title:    item.Title,
				Subtitle: item.Awarder,
				Date:     formatDate(item.Date),
				Summary:  markdownBlock(item.Summary),
			})
		}
		data.Sections = append(data.Sections, section{Title: "Awards", Experiences: experiences})
	}

	if len(r.Volunteer) > 0 {
		var experiences []experience
		for _, item := range r.Volunteer {
			experiences = append(experiences, experience{
				Title:      item.Position,
				Subtitle:   item.Organization,
				Date:       formatDateRange(item.StartDate, item.EndDate),
				Summary:    markdownBlock(item.Summary),
				Highlights: markdownList(item.Highlights),
			})
		}
		data.Sections = append(data.Sections, section{Title: "Volunteer", Experiences: experiences})
	}

	if len(r.Languages) > 0 {
		var lines []line
		for _, item := range r.Languages {
			lines = append(lines, line{Name: item.Language, Items: nonEmpty(item.Fluency)})
		}
		data.Sections = append(data.Sections, section{Title: "Languages", Lines: lines})
	}

	if len(r.Skills) > 0 {
		var lines []line
		for _, item := range r.Skills {
			lines = append(lines, line{Name: item.Name, Items: item.Keywords})
		}
		data.Sections = append(data.Sections, section{Title: "Skills", Lines: lines})
	}

	if len(r.Interests) > 0 {
		var lines []line
		for _, item := range r.Interests {
			lines = append(lines, line{Name: item.Name, Items: item.Keywords})
		}
		data.Sections = append(data.Sections, section{Title: "Interests", Lines: lines})
	}

	if len(r.References) > 0 {
		var refs []reference
		for _, item := range r.References {
			refs = append(refs, reference{Name: item.Name, Text: item.Reference})
		}
		data.Sections = append(data.Sections, section{Title: "References", References: refs})
	}

	return data, nil
}

func buildHero(b resume.Basics) hero {
	h := hero{Name: b.Name}
	location := joinNonEmpty(", ", b.Location.City, b.Location.CountryCode)
	if location != "" {
		h.Contacts = append(h.Contacts, contact{Icon: iconMapPin, Label: location})
	}
	if b.Email != "" {
		h.Contacts = append(h.Contacts, contact{Icon: iconEnvelope, Label: b.Email})
	}
	if b.Phone != "" {
		h.Contacts = append(h.Contacts, contact{Icon: iconPhone, Label: b.Phone})
	}
	if b.URL != "" {
		h.Contacts = append(h.Contacts, contact{Icon: iconLink, Label: b.URL, URL: b.URL})
	}

	for _, profile := range b.Profiles {
		switch strings.ToLower(profile.Network) {
		case "linkedin":
			h.Contacts = append(h.Contacts, contact{Icon: iconLinkedIn, Label: profile.Username, URL: "https://linkedin.com/in/" + profile.Username})
		case "github":
			h.Contacts = append(h.Contacts, contact{Icon: iconGithub, Label: profile.Username, URL: "https://github.com/" + profile.Username})
		case "twitter":
			h.Contacts = append(h.Contacts, contact{Icon: iconTwitter, Label: profile.Username, URL: "https://twitter.com/" + profile.Username})
		}
	}
	return h
}

func loadFonts() (fonts, error) {
	regular, err := fontDataURI("assets/fonts/lmroman10-regular.otf")
	if err != nil {
		return fonts{}, err
	}
	bold, err := fontDataURI("assets/fonts/lmroman10-bold.otf")
	if err != nil {
		return fonts{}, err
	}
	italic, err := fontDataURI("assets/fonts/lmroman10-italic.otf")
	if err != nil {
		return fonts{}, err
	}
	return fonts{Regular: regular, Bold: bold, Italic: italic}, nil
}

func fontDataURI(name string) (template.URL, error) {
	raw, err := fontFiles.ReadFile(name)
	if err != nil {
		return "", err
	}
	return template.URL("data:font/otf;base64," + base64.StdEncoding.EncodeToString(raw)), nil
}

func markdownBlock(input string) template.HTML {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}
	var out bytes.Buffer
	_ = markdown.Convert([]byte(markdownInput(input)), &out)
	return template.HTML(out.String())
}

func markdownInline(input string) template.HTML {
	rendered := strings.TrimSpace(string(markdownBlock(input)))
	if strings.HasPrefix(rendered, "<p>") && strings.HasSuffix(rendered, "</p>") {
		rendered = strings.TrimSuffix(strings.TrimPrefix(rendered, "<p>"), "</p>")
	}
	return template.HTML(rendered)
}

func markdownInput(input string) string {
	escaped := strings.ReplaceAll(input, "<", "&lt;")
	escaped = strings.ReplaceAll(escaped, ">", "&gt;")
	return strings.ReplaceAll(escaped, "\n", "  \n")
}

func markdownList(items []string) []template.HTML {
	if len(items) == 0 {
		return nil
	}
	rendered := make([]template.HTML, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		rendered = append(rendered, markdownInline(item))
	}
	return rendered
}

func formatDateRange(startDate, endDate string) string {
	return formatDate(startDate) + " - " + formatDate(endDate)
}

func formatDate(date string) string {
	date = strings.TrimSpace(date)
	if date == "" {
		return "Present"
	}

	layouts := []string{"2006-01-02", "2006-01", "2006"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, date); err == nil {
			return t.Format("January 2006")
		}
	}
	return date
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func nonEmpty(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return []string{value}
}

func joinNonEmpty(separator string, values ...string) string {
	var parts []string
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			parts = append(parts, value)
		}
	}
	return strings.Join(parts, separator)
}

var pageTemplate = template.Must(template.New("professional").Funcs(template.FuncMap{
	"join": strings.Join,
}).Parse(`<!DOCTYPE html>
<html>
<head>
  <title>{{.Title}}</title>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    @font-face {
      font-family: LatinModern;
      font-style: normal;
      font-weight: normal;
      src: url("{{.Fonts.Regular}}") format("opentype");
    }

    @font-face {
      font-family: LatinModern;
      font-weight: bold;
      src: url("{{.Fonts.Bold}}") format("opentype");
    }

    @font-face {
      font-family: LatinModern;
      font-style: italic;
      src: url("{{.Fonts.Italic}}") format("opentype");
    }

    @page {
      size: A4;
      margin: 0;
    }

    html {
      font-family: LatinModern, "Courier New", monospace;
      background: #fff;
      font-size: {{.CSS.BaseFontPX}};
      -webkit-print-color-adjust: exact;
      print-color-adjust: exact;
    }

    h2 {
      font-size: 1.65rem;
    }

    p {
      padding: 0;
      margin: 0;
    }

    p, li {
      font-size: 1.4rem;
      line-height: 1.5rem;
    }

    .secondary {
      color: #111;
    }

    a {
      color: inherit;
      text-decoration: none;
    }

    ul {
      list-style: none;
      margin: 0;
      padding: 0;
    }

    *,
    *::before,
    *::after {
      box-sizing: border-box;
    }

    .layout {
      max-width: 660px;
      margin: 0 auto {{.CSS.LayoutBottomPX}};
      line-height: calc(1ex / 0.32);
    }

    .section {
      max-width: 700px;
      margin: 0 auto {{.CSS.SectionBottomPX}};
    }

    .section h2 {
      margin: 0 0 3px;
      padding: 0;
      font-weight: 600;
    }

    .section hr {
      margin: 7px 0 3px;
      padding: 0;
    }

    .section-body {
      margin: 0 {{.CSS.SectionBodyMarginX}};
    }

    .title {
      font-size: 3rem;
      text-align: center;
      margin-top: {{.CSS.HeroMarginTopPX}};
      margin-bottom: {{.CSS.HeroMarginBottomPX}};
    }

    .basic-info {
      display: flex;
      gap: {{.CSS.ContactGapY}} {{.CSS.ContactGapX}};
      justify-content: center;
      flex-wrap: wrap;
    }

    .info {
      display: flex;
      align-items: center;
      font-size: 1.5rem;
    }

    .info svg {
      color: #000;
      margin-right: {{.CSS.IconMarginRightPX}};
      width: 10px;
      height: 10px;
      flex: 0 0 auto;
    }

    .experience {
      margin-bottom: {{.CSS.ExperienceBottomPX}};
    }

    .experience-meta {
      display: flex;
      justify-content: space-between;
      gap: {{.CSS.MetaGapPX}};
      margin-bottom: 2px;
    }

    .experience-title {
      font-weight: 600;
      font-size: 1.45rem;
      margin-bottom: 3px;
    }

    .date {
      font-style: italic;
      font-size: 1.4rem;
      white-space: nowrap;
    }

    .subtitle {
      font-style: italic;
      font-size: 1.4rem;
      margin-bottom: 3px;
    }

    .summary {
      margin-bottom: {{.CSS.SummaryBottomPX}};
    }

    .list {
      padding-left: {{.CSS.ListPaddingLeftPX}};
      line-height: 16px;
    }

    .list li::before {
      content: '•';
      display: inline-block;
      width: 1em;
      margin-left: -1em;
      line-height: 10px;
    }

    .one-line {
      margin-bottom: {{.CSS.LineBottomPX}};
      display: flex;
      align-items: baseline;
    }

    .one-line-name,
    .reference-name {
      font-weight: 600;
      font-size: 1.4rem;
    }

    .one-line-list {
      font-size: 1.4rem;
      margin-left: {{.CSS.LineListMarginPX}};
    }

    .reference {
      margin-bottom: {{.CSS.ReferenceBottomPX}};
    }

    .reference-name {
      margin-bottom: 5px;
    }
  </style>
</head>
<body>
  <main class="layout">
    <section class="section">
      <div class="section-body">
        <div class="title">{{.Hero.Name}}</div>
        <div class="secondary">
          <div class="basic-info">
            {{range .Hero.Contacts}}
              <div class="info">
                {{.Icon}}
                {{if .URL}}<a target="_blank" rel="noreferrer" href="{{.URL}}">{{.Label}}</a>{{else}}{{.Label}}{{end}}
              </div>
            {{end}}
          </div>
        </div>
      </div>
    </section>

    {{if .Summary}}
      <section class="section">
        <div class="section-body">
          <div class="secondary">{{.Summary}}</div>
        </div>
      </section>
    {{end}}

    {{range .Sections}}
      <section class="section">
        <h2>{{.Title}}</h2>
        <hr>
        <div class="section-body">
          {{range .Experiences}}
            <div class="experience">
              <div class="experience-meta">
                <div class="experience-title">{{.Title}}</div>
                <div class="secondary"><div class="date">{{.Date}}</div></div>
              </div>
              {{if .Subtitle}}<div class="subtitle">{{.Subtitle}}</div>{{end}}
              <div class="secondary">
                {{if .Summary}}<div class="summary">{{.Summary}}</div>{{end}}
                {{if .Highlights}}
                  <ul class="list">
                    {{range .Highlights}}<li>{{.}}</li>{{end}}
                  </ul>
                {{end}}
              </div>
            </div>
          {{end}}

          {{range .Lines}}
            <div class="one-line">
              <div class="one-line-name">{{.Name}}:</div>
              <div class="one-line-list"><div class="secondary">{{join .Items ", "}}</div></div>
            </div>
          {{end}}

          {{range .References}}
            <div class="reference">
              <div class="reference-name">{{.Name}}</div>
              <p>{{.Text}}</p>
            </div>
          {{end}}
        </div>
      </section>
    {{end}}
  </main>
</body>
</html>`))

const svgAttrs = `viewBox="0 0 16 16" aria-hidden="true" focusable="false"`

var (
	iconMapPin   = icon(`<path d="M8 1.3c-2.4 0-4.3 1.9-4.3 4.3 0 3.3 4.3 9.1 4.3 9.1s4.3-5.8 4.3-9.1c0-2.4-1.9-4.3-4.3-4.3zm0 6a1.7 1.7 0 1 1 0-3.4 1.7 1.7 0 0 1 0 3.4z"/>`)
	iconEnvelope = icon(`<path d="M1.5 3h13v10h-13V3zm1.2 1.5 5.3 4 5.3-4H2.7zm10.6 7V6L8 10 2.7 6v5.5h10.6z"/>`)
	iconGithub   = icon(`<path d="M8 .9a7.1 7.1 0 0 0-2.2 13.8c.4.1.5-.2.5-.4v-1.4c-2.1.5-2.6-.9-2.6-.9-.3-.8-.8-1-.8-1-.7-.5.1-.5.1-.5.8.1 1.1.8 1.1.8.7 1.1 1.8.8 2.2.6.1-.5.3-.8.5-1-1.7-.2-3.5-.9-3.5-3.8 0-.8.3-1.5.8-2-.1-.2-.3-1 .1-2 0 0 .6-.2 2.1.8.6-.2 1.2-.2 1.8-.2s1.2.1 1.8.2c1.4-1 2.1-.8 2.1-.8.4 1 .2 1.8.1 2 .5.5.8 1.2.8 2 0 2.9-1.8 3.6-3.5 3.8.3.2.5.7.5 1.3v1.9c0 .2.1.5.5.4A7.1 7.1 0 0 0 8 .9z"/>`)
	iconTwitter  = icon(`<path d="M14.7 4.4v.4c0 4.4-3.4 9.5-9.5 9.5-1.9 0-3.6-.6-5.1-1.5h.8c1.6 0 3-.5 4.1-1.4-1.5 0-2.7-1-3.1-2.3.2 0 .4.1.6.1.3 0 .6 0 .9-.1-1.5-.3-2.7-1.7-2.7-3.3.5.3 1 .4 1.5.4A3.4 3.4 0 0 1 1.2 1.7a9.5 9.5 0 0 0 6.9 3.5 3.4 3.4 0 0 1 5.8-3.1 6.7 6.7 0 0 0 2.1-.8 3.4 3.4 0 0 1-1.5 1.9 6.5 6.5 0 0 0 1.9-.5 7.2 7.2 0 0 1-1.7 1.7z"/>`)
	iconPhone    = icon(`<path d="M4.1 1.4 6 5.4 4.8 6.7c.8 1.6 2.1 2.9 3.7 3.7L9.8 9l4 1.9-.6 3.2c-.1.6-.6 1-1.2 1C5.8 15.1.9 10.2.9 4c0-.6.4-1.1 1-1.2l2.2-.4z"/>`)
	iconLink     = icon(`<path d="M6.6 10.8 5.4 12a2.3 2.3 0 1 1-3.3-3.3l2.2-2.2a2.3 2.3 0 0 1 3.3 0l.8.8-1.1 1.1-.8-.8a.8.8 0 0 0-1.1 0L3.2 9.8a.8.8 0 1 0 1.1 1.1l1.2-1.2 1.1 1.1zm2.8-5.6L10.6 4a2.3 2.3 0 1 1 3.3 3.3l-2.2 2.2a2.3 2.3 0 0 1-3.3 0l-.8-.8 1.1-1.1.8.8a.8.8 0 0 0 1.1 0l2.2-2.2a.8.8 0 1 0-1.1-1.1L10.5 6.3 9.4 5.2z"/>`)
	iconLinkedIn = icon(`<path d="M2.6 5.8h2.2v7.3H2.6V5.8zm1.1-3.6a1.3 1.3 0 1 1 0 2.6 1.3 1.3 0 0 1 0-2.6zm2.6 3.6h2.1v1h.1c.3-.6 1.1-1.2 2.2-1.2 2.4 0 2.8 1.6 2.8 3.6v3.9h-2.2V9.6c0-.8 0-1.9-1.2-1.9s-1.4.9-1.4 1.8v3.6H6.3V5.8z"/>`)
)

func icon(path string) template.HTML {
	return template.HTML(`<svg ` + svgAttrs + `>` + path + `</svg>`)
}
