package model

import "github.com/konveyor/tackle2-hub/settings"

//
// JSON field (data) type.
type JSON = []byte

var (
	Settings = &settings.Settings
)

//
// All builds all models.
// Models are enumerated such that each are listed after
// all the other models on which they may depend.
func All() []interface{} {
	return []interface{}{
		Application{},
		TechDependency{},
		Incident{},
		Analysis{},
		Issue{},
		Bucket{},
		BusinessService{},
		Dependency{},
		File{},
		Fact{},
		Identity{},
		Import{},
		ImportSummary{},
		ImportTag{},
		JobFunction{},
		MigrationWave{},
		Proxy{},
		Review{},
		Setting{},
		RuleSet{},
		Rule{},
		Stakeholder{},
		StakeholderGroup{},
		Tag{},
		TagCategory{},
		Target{},
		Task{},
		TaskGroup{},
		TaskReport{},
		Ticket{},
		Tracker{},
		ApplicationTag{},
		Questionnaire{},
		Assessment{},
		Archetype{},
	}
}
