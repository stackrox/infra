package create

func isQaDemoFlavor(flavorID string) bool {
	return flavorID == "qa-demo" || flavorID == "test-qa-demo"
}

func wasNameProvided(args []string) bool {
	return len(args) >= 2 && args[1] != ""
}
