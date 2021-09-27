package sparta

func embeddedString(path string) (string, error) {
	vals, valErr := embeddedFS.ReadFile(path)
	if valErr != nil {
		return "", valErr
	}
	return string(vals), nil
}

func embeddedMustString(path string) string {
	vals, valErr := embeddedFS.ReadFile(path)
	if valErr != nil {
		panic("Failed to read file:" + path)
	}
	return string(vals)
}
