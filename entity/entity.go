package entity

import "math"

type Entity struct {
	Name string
	cost int
	cps  float64
	Icon rune
}

func (e Entity) Cost(n, owned int) float64 {
	if n <= 0 {
		return 0
	}

	nextPrice := float64(e.cost) * (math.Pow(1.15, float64(owned)))
	nthPrice := nextPrice * (math.Pow(1.15, float64(n-1)))

	return nthPrice
}

func (e Entity) Cps(n int) float64 {
	return float64(n) * e.cps
}

func newEntity(name string, cost int, cps float64, icon rune) Entity {
	return Entity{Name: name, cost: cost, cps: cps, Icon: icon}
}

func listEntities() []Entity {
	cursor := newEntity("cursor", 15, 0.1, 0x1F449)
	grandma := newEntity("grandma", 100, 1, 0x1F475)
	farm := newEntity("farm", 1100, 8, 0x1F69C)
	mine := newEntity("mine", 12000, 47, 0x1FAA8)
	factory := newEntity("factory", 130000, 260, 0x1F3E2)

	return []Entity{cursor, grandma, farm, mine, factory}
}

func Items() (map[string]Entity, []string) {
	lookup := make(map[string]Entity)
	items := make([]string, 0)

	for _, e := range listEntities() {
		lookup[e.Name] = e
		items = append(items, e.Name)
	}
	return lookup, items
}
