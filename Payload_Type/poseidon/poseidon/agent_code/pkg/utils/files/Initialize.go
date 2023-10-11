package files

const FILE_CHUNK_SIZE = 512000 //Normal mythic chunk size

func Initialize() {
	// start listening for sending a file to Mythic ("download")
	go listenForSendFileToMythicMessages()
	// start listening for getting a file from Mythic ("upload")
	go listenForGetFromMythicMessages()
}
