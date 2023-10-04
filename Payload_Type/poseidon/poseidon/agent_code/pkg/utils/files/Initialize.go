package files

func Initialize() {
	go processSendToMythicChannel()
	go processGetFromMythicChannel()
}
