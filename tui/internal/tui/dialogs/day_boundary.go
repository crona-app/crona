package dialogs

func fallbackTimezone(value string) string {
	if value == "" {
		return "daemon local time"
	}
	return value
}
