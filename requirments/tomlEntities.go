package requirments

type tomlReq struct {
	Component   []Component
	Feature     []Feature
	Module      []Module
	Function    []Function
	Risk        []Risk
	Stakeholder []Stakeholder
	Format      []Format
	Goal        []Goal

	// This sub structure seems to be needed here and in entity. Sort this out.

}

type Entity struct {
	Id           string
	Status       string
	Title        string
	Author       []string
	Desc         string
	Spec         string
	File         []string
	Relationship []struct {
		Type   string
		Status string
		To     string
	}
	Test []struct {
		Reference string
		Framework string
		Desc      string
	}
}

type Component struct {
	Entity
}

type Feature struct {
	Entity
}

type Format struct {
	Entity
}

type Function struct {
	Entity
}

type Goal struct {
	Entity
}

type Module struct {
	Entity
}

type Risk struct {
	Entity
}

type Stakeholder struct {
	Entity
}
