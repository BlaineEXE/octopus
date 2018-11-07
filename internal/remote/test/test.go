package test

// Clear takes a pointer to a result list as a param, returns the contents, and empties the list.
// e.g., Clear(&m.Commands)
func Clear(list *[]string) (contents []string) {
	contents = *list
	*list = []string{}
	return
}

// append to string slice in-place
func app(ss *[]string, s string) {
	*ss = append(*ss, s)
}
