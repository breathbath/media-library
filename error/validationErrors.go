package error

type ValidationErrors map[string][]string

func NewValidationErrors() ValidationErrors {
	return make(map[string][]string)
}

func (ves ValidationErrors) Merge(ves2 ValidationErrors) {
	for field, ve := range ves2 {
		_, ok := ves[field]
		if !ok {
			ves[field] = []string{}
		}
		ves[field] = append(ves[field], ve...)
	}
}
