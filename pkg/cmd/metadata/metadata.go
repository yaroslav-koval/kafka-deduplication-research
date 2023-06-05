package metadata

// default version
// example: change before build
// - module:notification
// - version:4d7f728-dirty
// - build:20230124.125009.N.+0200
var (
	_version   = ""
	_module    = ""
	_buildDate = ""
)

type Meta struct {
	Version   string `json:"version"`
	Module    string `json:"module"`
	BuildDate string `json:"build_date"`
}

func New() Meta {
	return Meta{
		Version:   _version,
		Module:    _module,
		BuildDate: _buildDate,
	}
}
