package content

type Link struct {
	Link  string `json:"link"`
	Title string `json:"title"`
}

type Section struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Meta struct {
	Title   string   `yaml:"title" json:"title"`
	Links   []string `yaml:"links" json:"links"`
	Actions []string `yaml:"actions" json:"actions"`
}

type Item struct {
	Meta     *Meta      `json:"meta"`
	Content  string     `json:"content"`
	Sections []*Section `json:"sections,omitempty"`
	Links    []*Link    `json:"links,omitempty"`
}
