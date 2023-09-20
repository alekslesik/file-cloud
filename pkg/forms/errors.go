package forms

// Type to hold validation error messages for forms
type errors map[string][]string

// Add error messages for a given field to the map
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}


func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}
	return es[0]
}

// Sum all error messages to one message separated by sep
func (e errors) WholeErrorMessage(sep string) string {
	var m string

	for _, messages := range e {
		for _, message := range messages {
			m += m + sep + message
		}
	}

	return m
}
