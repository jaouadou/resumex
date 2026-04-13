package resume

import (
	"encoding/json"
	"io"
)

type Resume struct {
	Basics       Basics        `json:"basics"`
	Work         []Work        `json:"work"`
	Volunteer    []Volunteer   `json:"volunteer"`
	Education    []Education   `json:"education"`
	Awards       []Award       `json:"awards"`
	Certificates []Certificate `json:"certificates"`
	Publications []Publication `json:"publications"`
	Skills       []Skill       `json:"skills"`
	Languages    []Language    `json:"languages"`
	Interests    []Interest    `json:"interests"`
	References   []Reference   `json:"references"`
	Projects     []Project     `json:"projects"`
}

type Basics struct {
	Name     string    `json:"name"`
	Label    string    `json:"label"`
	Email    string    `json:"email"`
	Phone    string    `json:"phone"`
	URL      string    `json:"url"`
	Summary  string    `json:"summary"`
	Location Location  `json:"location"`
	Profiles []Profile `json:"profiles"`
}

type Location struct {
	City        string `json:"city"`
	CountryCode string `json:"countryCode"`
}

type Profile struct {
	Network  string `json:"network"`
	Username string `json:"username"`
	URL      string `json:"url"`
}

type Work struct {
	Name       string   `json:"name"`
	Position   string   `json:"position"`
	StartDate  string   `json:"startDate"`
	EndDate    string   `json:"endDate"`
	Summary    string   `json:"summary"`
	Highlights []string `json:"highlights"`
}

type Volunteer struct {
	Organization string   `json:"organization"`
	Position     string   `json:"position"`
	StartDate    string   `json:"startDate"`
	EndDate      string   `json:"endDate"`
	Summary      string   `json:"summary"`
	Highlights   []string `json:"highlights"`
}

type Education struct {
	Institution string   `json:"institution"`
	Area        string   `json:"area"`
	StudyType   string   `json:"studyType"`
	StartDate   string   `json:"startDate"`
	EndDate     string   `json:"endDate"`
	Score       string   `json:"score"`
	Courses     []string `json:"courses"`
}

type Award struct {
	Title   string `json:"title"`
	Date    string `json:"date"`
	Awarder string `json:"awarder"`
	Summary string `json:"summary"`
}

type Certificate struct {
	Name   string `json:"name"`
	Date   string `json:"date"`
	Issuer string `json:"issuer"`
}

type Publication struct {
	Name        string `json:"name"`
	Publisher   string `json:"publisher"`
	ReleaseDate string `json:"releaseDate"`
	Summary     string `json:"summary"`
}

type Skill struct {
	Name     string   `json:"name"`
	Keywords []string `json:"keywords"`
}

type Language struct {
	Language string `json:"language"`
	Fluency  string `json:"fluency"`
}

type Interest struct {
	Name     string   `json:"name"`
	Keywords []string `json:"keywords"`
}

type Reference struct {
	Name      string `json:"name"`
	Reference string `json:"reference"`
}

type Project struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	StartDate   string   `json:"startDate"`
	EndDate     string   `json:"endDate"`
	Highlights  []string `json:"highlights"`
}

func Decode(r io.Reader) (*Resume, error) {
	var parsed Resume
	if err := json.NewDecoder(r).Decode(&parsed); err != nil {
		return nil, err
	}
	return &parsed, nil
}
