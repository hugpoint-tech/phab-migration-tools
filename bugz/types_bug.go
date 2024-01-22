package bugz

type Bug struct {
	ID                  int           `json:"id"`
	Product             string        `json:"product"`
	Status              string        `json:"status"`
	Priority            string        `json:"priority"`
	Severity            string        `json:"severity"`
	Component           string        `json:"component"`
	Platform            string        `json:"platform"`
	IsCCAccessible      bool          `json:"is_cc_accessible"`
	Blocks              []int         `json:"blocks"`
	DependsOn           []int         `json:"depends_on"`
	CreationTime        string        `json:"creation_time"`
	Resolution          string        `json:"resolution"`
	SeeAlso             []string      `json:"see_also"`
	Keywords            []string      `json:"keywords"`
	IsCreatorAccessible bool          `json:"is_creator_accessible"`
	Version             string        `json:"version"`
	Summary             string        `json:"summary"`
	URL                 string        `json:"url"`
	Groups              []string      `json:"groups"`
	Alias               []string      `json:"alias"`
	IsConfirmed         bool          `json:"is_confirmed"`
	LastChangeTime      string        `json:"last_change_time"`
	Flags               []interface{} `json:"flags"`
	DupeOf              int           `json:"dupe_of"`

	Creator       string     `json:"creator"`
	CreatorDetail UserDetail `json:"creator_detail"`

	AssignedTo       string     `json:"assigned_to"`
	AssignedToDetail UserDetail `json:"assigned_to_detail"`

	CC       []string     `json:"cc"`
	CCDetail []UserDetail `json:"cc_detail"`
}

type UserDetail struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	RealName string `json:"real_name"`
	Name     string `json:"name"`
}

type BugsResponse struct {
	Bugs []Bug `json:"bugs"`
}
